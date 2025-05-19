package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/WOLFnik5/weather_subscriber/db"
	"github.com/WOLFnik5/weather_subscriber/model"
	"github.com/gorilla/mux"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestMain(m *testing.M) {
	// mysqlUserFromEnv := "mykola"
	// mysqlPasswordFromEnv := "4444"
	//  mysqlHost := "db"

	// if err == nil {
	// 	log.Println("Loaded environment variables from .env_test")
	// } else {
	// 	log.Println("No .env_test file found or error loading it, using existing environment variables or defaults.")
	// }
	mysqlUserFromEnv := os.Getenv("MYSQL_USER")
	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlPort := os.Getenv("MYSQL_PORT")
	testDBName := os.Getenv("MYSQL_DATABASE")

	// mysqlHostLocal := "127.0.0.1"
	// mysqlPort := "3306"
	// testDBName := "weather_subscriber_test"

	log.Printf("TestMain: Attempting to connect to test DB: user=%s, host=%s, port=%s, dbname=%s\n", mysqlUserFromEnv, mysqlHost, mysqlPort, testDBName)
	// os.Setenv("MYSQL_USER", mysqlUserFromEnv)
	// os.Setenv("MYSQL_PASSWORD", mysqlPasswordFromEnv)
	// os.Setenv("MYSQL_HOST", mysqlHost)
	// os.Setenv("MYSQL_PORT", mysqlPort)
	// os.Setenv("MYSQL_DATABASE", testDBName)
	err := db.Connect()

	if err != nil {
		log.Printf("TestMain: Failed to connect to DB with host '%s' (error: %v)", mysqlHost, err)
		// os.Setenv("MYSQL_USER", mysqlUserFromEnv)
		// os.Setenv("MYSQL_PASSWORD", mysqlPasswordFromEnv)
		// os.Setenv("MYSQL_HOST", mysqlHostLocal)
		// os.Setenv("MYSQL_PORT", mysqlPort)
		// os.Setenv("MYSQL_DATABASE", testDBName)
		// err = db.Connect()
		panic(fmt.Sprintf("TestMain: Failed to connect to database for testing: %v. \nEnsure the database '%s' exists and user '%s' has ALL PRIVILEGES on it. \nEnsure migrations have been applied to '%s'.", err, testDBName, mysqlUserFromEnv, testDBName))
		// if err != nil {
		// 	// Встановлюємо MYSQL_ROOT_PASSWORD, якщо він не встановлений, для повідомлення паніки
		// 	// Це значення з вашого .env файлу
		// 	if os.Getenv("MYSQL_ROOT_PASSWORD") == "" {
		// 		os.Setenv("MYSQL_ROOT_PASSWORD", "3333")
		// 	}
		// 	panic(fmt.Sprintf("TestMain: Failed to connect to any database for testing: %v. \nEnsure the database '%s' exists and user '%s' has ALL PRIVILEGES on it. \nEnsure migrations have been applied to '%s'. \nYou might need to run: \n1. docker-compose up -d db \n2. docker-compose exec db mysql -uroot -p%s -e \"CREATE DATABASE IF NOT EXISTS %s;\" \n3. docker-compose exec db mysql -uroot -p%s -e \"GRANT ALL PRIVILEGES ON %s.* TO '%s'@'%%';\" \n4. (Using PowerShell/CMD): docker-compose run --rm --entrypoint \"\" migrate migrate -path=/migrations -database=\"mysql://%s:%s@tcp(db:3306)/%s?parseTime=true&multiStatements=true\" up",
		// 		err, testDBName, mysqlUserFromEnv, testDBName, os.Getenv("MYSQL_ROOT_PASSWORD"), testDBName, os.Getenv("MYSQL_ROOT_PASSWORD"), testDBName, mysqlUserFromEnv, mysqlUserFromEnv, mysqlPasswordFromEnv, testDBName))
		// }
	}
	log.Println("TestMain: Successfully connected to test DB.")

	clearTables() // Очищаємо таблиці один раз після підключення та міграцій

	exitVal := m.Run()
	os.Exit(exitVal)
}

func clearTables() {
	if db.DB == nil {
		log.Println("clearTables: DB connection is nil. Skipping.")
		return
	}
	log.Println("clearTables: Clearing tables in test DB...")

	db.DB.Exec("DELETE FROM subscriptions")
	// db.DB.Exec("DELETE FROM weather_forecasts") // Якщо ця таблиця використовується
	db.DB.Exec("DELETE FROM users")
	db.DB.Exec("DELETE FROM cities")

	db.DB.Exec("ALTER TABLE users AUTO_INCREMENT = 1")
	db.DB.Exec("ALTER TABLE cities AUTO_INCREMENT = 1")
	db.DB.Exec("ALTER TABLE subscriptions AUTO_INCREMENT = 1")
	// db.DB.Exec("ALTER TABLE weather_forecasts AUTO_INCREMENT = 1")

	// Додамо кілька міст для тестів. Використовуємо ON DUPLICATE KEY UPDATE для ідемпотентності.
	_, err := db.DB.Exec("INSERT INTO cities (id, name, country) VALUES (1, 'Test City', 'Test Country'), (2, 'Another City', 'Test Country') ON DUPLICATE KEY UPDATE name=VALUES(name), country=VALUES(country)")
	if err != nil {
		log.Printf("clearTables: Warning - could not seed cities for testing: %s. This might be OK if tables were just created by migrations.", err.Error())
	} else {
		log.Println("clearTables: Seeded initial cities.")
	}
}

type HandlerTestSuite struct {
	suite.Suite
	router *mux.Router
}

func (suite *HandlerTestSuite) SetupSuite() {
	log.Println("HandlerTestSuite.SetupSuite: Setting up router.")

	suite.Require().NotNil(db.DB, "HandlerTestSuite.SetupSuite: Database connection should have been established by TestMain and be available globally via db.DB")

	suite.router = mux.NewRouter()
	suite.router.HandleFunc("/subscriptions", HandleCreateSubscription).Methods("POST")
	suite.router.HandleFunc("/subscriptions", HandleListSubscriptions).Methods("GET")
	suite.router.HandleFunc("/cities", HandleListCities).Methods("GET")
}

func (suite *HandlerTestSuite) TearDownSuite() {
	log.Println("HandlerTestSuite.TearDownSuite: Tearing down.")
}

func (suite *HandlerTestSuite) SetupTest() {
	log.Println("HandlerTestSuite.SetupTest: Clearing tables for a new test in suite.")
	clearTables()
}

func (suite *HandlerTestSuite) TestIntegration_CreateAndListSubscriptions() {
	log.Println("HandlerTestSuite.TestIntegration_CreateAndListSubscriptions: Running...")
	// 1. Створення підписки
	payload := subscriptionInput{
		Email:     "integration@example.com",
		CityID:    1, // Припускаємо, місто з ID 1 існує (засіяне clearTables)
		Frequency: "hourly",
	}
	body, _ := json.Marshal(payload)

	reqCreate, _ := http.NewRequest("POST", "/subscriptions", bytes.NewBuffer(body))
	reqCreate.Header.Set("Content-Type", "application/json")
	rrCreate := httptest.NewRecorder()
	suite.router.ServeHTTP(rrCreate, reqCreate)

	suite.Assert().Equal(http.StatusCreated, rrCreate.Code)
	var createdSub model.Subscription
	err := json.Unmarshal(rrCreate.Body.Bytes(), &createdSub)
	suite.Require().NoError(err)
	suite.Assert().Equal(payload.Email, createdSub.Email)
	suite.Assert().Equal(payload.CityID, createdSub.CityID)

	// 2. Отримання списку підписок
	reqList, _ := http.NewRequest("GET", "/subscriptions", nil)
	rrList := httptest.NewRecorder()
	suite.router.ServeHTTP(rrList, reqList)

	suite.Assert().Equal(http.StatusOK, rrList.Code)
	var subs []model.Subscription
	err = json.Unmarshal(rrList.Body.Bytes(), &subs)
	suite.Require().NoError(err)
	suite.Assert().Len(subs, 1, "Expected one subscription in the list")
	if len(subs) == 1 {
		suite.Assert().Equal(payload.Email, subs[0].Email)
		suite.Assert().Equal(payload.CityID, subs[0].CityID)
		suite.Assert().Equal(payload.Frequency, subs[0].Frequency)
	}
}

func TestHandlerTestSuite(t *testing.T) {
	log.Println("TestHandlerTestSuite: Kicking off the suite.")
	suite.Run(t, new(HandlerTestSuite))
}

func TestHandleCreateSubscription_Success(t *testing.T) {
	clearTables()

	payload := subscriptionInput{
		Email:     "test@example.com",
		CityID:    1, // Припускаємо, що місто з ID 1 існує
		Frequency: "daily",
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "/subscriptions", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handlerFunc := http.HandlerFunc(HandleCreateSubscription)
	handlerFunc.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code, "Expected status code 201")

	var createdSubscription model.Subscription
	err = json.NewDecoder(rr.Body).Decode(&createdSubscription)
	assert.NoError(t, err, "Should be able to decode response")
	assert.Equal(t, payload.Email, createdSubscription.Email)
	assert.Equal(t, payload.CityID, createdSubscription.CityID)
	assert.Equal(t, payload.Frequency, createdSubscription.Frequency)
	assert.NotZero(t, createdSubscription.ID, "Subscription ID should not be zero")
}

func TestHandleCreateSubscription_InvalidInput(t *testing.T) {
	clearTables()

	tests := []struct {
		name           string
		payload        string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Bad JSON",
			payload:        `{"email": "test@example.com", "city_id": 1, "frequency": "daily"`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid input",
		},
		{
			name:           "Missing email",
			payload:        `{"city_id": 1, "frequency": "daily"}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Missing required fields",
		},
		{
			name:           "Missing city_id",
			payload:        `{"email": "test@example.com", "frequency": "daily"}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Missing required fields",
		},
		{
			name:           "Missing frequency",
			payload:        `{"email": "test@example.com", "city_id": 1}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Missing required fields",
		},
		{
			name:           "Invalid frequency",
			payload:        `{"email": "test@example.com", "city_id": 1, "frequency": "weekly"}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid frequency, must be 'daily' or 'hourly'",
		},
		{
			name:           "Non-existent city_id",
			payload:        `{"email": "test@example.com", "city_id": 999, "frequency": "daily"}`,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to create subscription",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/subscriptions", strings.NewReader(tc.payload))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handlerFunc := http.HandlerFunc(HandleCreateSubscription)
			handlerFunc.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.expectedBody)
		})
	}
}

func TestHandleListSubscriptions_Success(t *testing.T) {
	clearTables()

	sub1 := model.Subscription{Email: "user1@example.com", CityID: 1, Frequency: "daily"}
	sub2 := model.Subscription{Email: "user2@example.com", CityID: 2, Frequency: "hourly"}

	errSub1 := model.CreateSubscription(&sub1)
	assert.NoError(t, errSub1, "Failed to create subscription 1 for test")
	errSub2 := model.CreateSubscription(&sub2)
	assert.NoError(t, errSub2, "Failed to create subscription 2 for test")

	req, err := http.NewRequest("GET", "/subscriptions", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handlerFunc := http.HandlerFunc(HandleListSubscriptions)
	handlerFunc.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var subscriptions []model.Subscription
	err = json.NewDecoder(rr.Body).Decode(&subscriptions)
	assert.NoError(t, err)
	assert.Len(t, subscriptions, 2, "Expected 2 subscriptions")

	foundSub1 := false
	foundSub2 := false
	for _, s := range subscriptions {
		if s.Email == sub1.Email && s.CityID == sub1.CityID {
			foundSub1 = true
		}
		if s.Email == sub2.Email && s.CityID == sub2.CityID {
			foundSub2 = true
		}
	}
	assert.True(t, foundSub1, "Subscription 1 not found in list")
	assert.True(t, foundSub2, "Subscription 2 not found in list")
}

func TestHandleListSubscriptions_Empty(t *testing.T) {
	clearTables()

	req, err := http.NewRequest("GET", "/subscriptions", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handlerFunc := http.HandlerFunc(HandleListSubscriptions)
	handlerFunc.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var subscriptions []model.Subscription
	err = json.NewDecoder(rr.Body).Decode(&subscriptions)
	assert.NoError(t, err)
	assert.Len(t, subscriptions, 0, "Expected 0 subscriptions")
}
