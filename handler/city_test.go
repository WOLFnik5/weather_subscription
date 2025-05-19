package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/WOLFnik5/weather_subscriber/db"
	"github.com/WOLFnik5/weather_subscriber/model"
	"github.com/stretchr/testify/assert"
)

// --- Тести для HandleListCities ---

func TestHandleListCities_Success_DefaultParams(t *testing.T) {
	clearTables() // Викликаємо для очищення та заповнення містами

	req, err := http.NewRequest("GET", "/cities", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleListCities)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var cities []model.City
	err = json.NewDecoder(rr.Body).Decode(&cities)
	assert.NoError(t, err)

	// У clearTables ми додаємо 2 міста.
	assert.Len(t, cities, 2, "Expected 2 cities with default params")
}

func TestHandleListCities_Success_WithParams(t *testing.T) {
	clearTables() // Очищаємо та додаємо 2 міста
	// Додамо ще кілька міст, щоб пагінація мала сенс
	// (припускаючи, що db.DB ініціалізовано і доступно)
	if db.DB != nil {
		_, err := db.DB.Exec("INSERT INTO cities (id, name, country) VALUES (3, 'City C', 'Country C'), (4, 'City D', 'Country D'), (5, 'City E', 'Country E')")
		assert.NoError(t, err, "Failed to insert additional cities for testing pagination")
	} else {
		t.Fatal("DB connection not initialized for TestHandleListCities_Success_WithParams")
	}

	// Тепер у нас 5 міст (2 з clearTables + 3 щойно додані)
	// IDs будуть 1, 2, 3, 4, 5

	tests := []struct {
		name            string
		queryParams     string
		expectedLength  int
		expectedFirstID int64 // Перевіримо ID першого міста для порядку
	}{
		{
			name:            "Limit 2, Offset 0",
			queryParams:     "?limit=2&offset=0",
			expectedLength:  2,
			expectedFirstID: 1,
		},
		{
			name:            "Limit 2, Offset 2",
			queryParams:     "?limit=2&offset=2",
			expectedLength:  2,
			expectedFirstID: 3,
		},
		{
			name:            "Limit 10, Offset 0 (get all 5)",
			queryParams:     "?limit=10&offset=0",
			expectedLength:  5,
			expectedFirstID: 1,
		},
		{
			name:            "Limit 3, Offset 3 (get remaining 2)",
			queryParams:     "?limit=3&offset=3",
			expectedLength:  2,
			expectedFirstID: 4,
		},
		{
			name:            "Limit 5, Offset 5 (empty result)",
			queryParams:     "?limit=5&offset=5",
			expectedLength:  0,
			expectedFirstID: 0, // Немає першого ID
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/cities"+tc.queryParams, nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(HandleListCities)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)

			var cities []model.City
			err = json.NewDecoder(rr.Body).Decode(&cities)
			assert.NoError(t, err)
			assert.Len(t, cities, tc.expectedLength)

			if tc.expectedLength > 0 && tc.expectedFirstID > 0 {
				assert.Equal(t, tc.expectedFirstID, cities[0].ID, "First city ID mismatch")
			}
		})
	}
}

// --- Інтеграційні тести для City за допомогою Suite ---
// Ми можемо додати ці тести до існуючого HandlerTestSuite
// або створити окремий CityHandlerTestSuite. Для простоти,
// припустимо, що ми додаємо їх до HandlerTestSuite в handler_test.go,
// просто переконавшись, що маршрут /cities доданий до suite.router.

// Приклад тесту, який можна додати до HandlerTestSuite:
func (suite *HandlerTestSuite) TestIntegration_ListCities_Pagination() {
	// clearTables вже викликається в suite.SetupTest() і додає 2 міста.
	// Додамо ще кілька для тестування пагінації.
	if db.DB != nil {
		_, err := db.DB.Exec("INSERT INTO cities (id, name, country) VALUES (3, 'Paged City 1', 'PageLand'), (4, 'Paged City 2', 'PageLand'), (5, 'Paged City 3', 'PageLand')")
		suite.Require().NoError(err, "Failed to insert additional cities for pagination test")
	} else {
		suite.T().Fatal("DB connection not initialized for TestIntegration_ListCities_Pagination")
	}
	// Тепер у нас 5 міст: Test City (1), Another City (2), Paged City 1 (3), Paged City 2 (4), Paged City 3 (5)

	// Запит: limit=2, offset=1
	// Очікуємо: Another City (2), Paged City 1 (3)
	req, _ := http.NewRequest("GET", "/cities?limit=2&offset=1", nil)
	rr := httptest.NewRecorder()
	suite.router.ServeHTTP(rr, req) // Використовуємо router з suite

	suite.Assert().Equal(http.StatusOK, rr.Code)
	var cities []model.City
	err := json.Unmarshal(rr.Body.Bytes(), &cities)
	suite.Require().NoError(err)
	suite.Assert().Len(cities, 2, "Expected 2 cities with limit=2, offset=1")
	if len(cities) == 2 {
		suite.Assert().Equal("Another City", cities[0].Name) // ID 2
		suite.Assert().Equal("Paged City 1", cities[1].Name) // ID 3
	}
}

// Щоб запустити тести з city_test.go, ви можете просто виконати `go test`
// у каталозі `handler`. Якщо ви додали TestIntegration_ListCities_Pagination
// до HandlerTestSuite, він запуститься разом з іншими тестами suite.
