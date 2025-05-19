package model

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/WOLFnik5/weather_subscriber/db"
	"github.com/stretchr/testify/assert"
)

func TestValidateFrequency(t *testing.T) {
	asserter := assert.New(t)
	tests := []struct {
		name      string
		frequency string
		expectErr bool
	}{
		{"Valid daily", "daily", false},
		{"Valid hourly", "hourly", false},
		{"Invalid weekly", "weekly", true},
		{"Empty frequency", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateFrequency(tc.frequency)
			if tc.expectErr {
				asserter.Error(err)
				asserter.EqualError(err, "invalid frequency, must be 'daily' or 'hourly'")
			} else {
				asserter.NoError(err)
			}
		})
	}
}

func TestModel_CreateSubscription_NewUser(t *testing.T) {
	asserter := assert.New(t)
	mockDB, mock, err := sqlmock.New()
	asserter.NoError(err, "Failed to create sqlmock")
	defer mockDB.Close()

	originalDB := db.DB
	db.DB = mockDB
	defer func() { db.DB = originalDB }()

	subInput := &Subscription{
		Email:     "test@example.com",
		CityID:    123,
		Frequency: "daily",
	}
	newUserID := int64(1)
	newSubscriptionID := int64(10)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM users WHERE email = ?")).
		WithArgs(subInput.Email).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (email) VALUES (?)")).
		WithArgs(subInput.Email).
		WillReturnResult(sqlmock.NewResult(newUserID, 1))

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO subscriptions (user_id, city_id, frequency) VALUES (?, ?, ?)")).
		WithArgs(newUserID, subInput.CityID, subInput.Frequency).
		WillReturnResult(sqlmock.NewResult(newSubscriptionID, 1))

	errCreate := CreateSubscription(subInput)
	asserter.NoError(errCreate)
	asserter.Equal(newSubscriptionID, subInput.ID, "Subscription ID should be set to the new ID")

	asserter.NoError(mock.ExpectationsWereMet())
}

func TestModel_CreateSubscription_ExistingUser(t *testing.T) {
	asserter := assert.New(t)
	mockDB, mock, err := sqlmock.New()
	asserter.NoError(err, "Failed to create sqlmock")
	defer mockDB.Close()

	originalDB := db.DB
	db.DB = mockDB
	defer func() { db.DB = originalDB }()

	subInput := &Subscription{
		Email:     "existing@example.com",
		CityID:    456,
		Frequency: "hourly",
	}
	existingUserID := int64(5)
	newSubscriptionID := int64(11)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(existingUserID)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM users WHERE email = ?")).
		WithArgs(subInput.Email).
		WillReturnRows(rows)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO subscriptions (user_id, city_id, frequency) VALUES (?, ?, ?)")).
		WithArgs(existingUserID, subInput.CityID, subInput.Frequency).
		WillReturnResult(sqlmock.NewResult(newSubscriptionID, 1))

	errCreate := CreateSubscription(subInput)
	asserter.NoError(errCreate)
	asserter.Equal(newSubscriptionID, subInput.ID)

	asserter.NoError(mock.ExpectationsWereMet())
}

func TestModel_ListSubscriptions(t *testing.T) {
	asserter := assert.New(t)
	mockDB, mock, err := sqlmock.New()
	asserter.NoError(err, "Failed to create sqlmock")
	defer mockDB.Close()

	originalDB := db.DB
	db.DB = mockDB
	defer func() { db.DB = originalDB }()

	nowStr := time.Now().Format("2006-01-02 15:04:05")

	expectedSubsData := []struct {
		ID        int64
		Email     string
		CityID    int64
		Frequency string
		CreatedAt string
	}{
		{ID: 1, Email: "user1@example.com", CityID: 10, Frequency: "daily", CreatedAt: nowStr},
		{ID: 2, Email: "user2@example.com", CityID: 20, Frequency: "hourly", CreatedAt: nowStr},
	}

	rows := sqlmock.NewRows([]string{"s.id", "u.email", "s.city_id", "s.frequency", "s.created_at"})
	for _, data := range expectedSubsData {
		rows.AddRow(data.ID, data.Email, data.CityID, data.Frequency, data.CreatedAt)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT s.id, u.email, s.city_id, s.frequency, s.created_at
		FROM subscriptions s
		JOIN users u ON s.user_id = u.id`)).WillReturnRows(rows)

	subs, errList := ListSubscriptions()
	asserter.NoError(errList)
	asserter.Len(subs, 2)

	if len(subs) == 2 {
		asserter.Equal(expectedSubsData[0].Email, subs[0].Email)
		asserter.Equal(expectedSubsData[0].Frequency, subs[0].Frequency)
		asserter.Equal(expectedSubsData[0].CreatedAt, subs[0].CreatedAt)

		asserter.Equal(expectedSubsData[1].Email, subs[1].Email)
		asserter.Equal(expectedSubsData[1].Frequency, subs[1].Frequency)
		asserter.Equal(expectedSubsData[1].CreatedAt, subs[1].CreatedAt)
	}

	asserter.NoError(mock.ExpectationsWereMet())
}

func TestModel_GetOrCreateUserIDByEmail_CreateNew(t *testing.T) {
	asserter := assert.New(t)
	mockDB, mock, err := sqlmock.New()
	asserter.NoError(err, "Failed to create sqlmock")
	defer mockDB.Close()

	originalDB := db.DB
	db.DB = mockDB
	defer func() { db.DB = originalDB }()

	email := "newuser@example.com"
	expectedUserID := int64(100)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM users WHERE email = ?")).
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (email) VALUES (?)")).
		WithArgs(email).
		WillReturnResult(sqlmock.NewResult(expectedUserID, 1))

	userID, errFunc := getOrCreateUserIDByEmail(email)
	asserter.NoError(errFunc)
	asserter.Equal(expectedUserID, userID)

	asserter.NoError(mock.ExpectationsWereMet())
}

func TestModel_GetOrCreateUserIDByEmail_GetExisting(t *testing.T) {
	asserter := assert.New(t)
	mockDB, mock, err := sqlmock.New()
	asserter.NoError(err, "Failed to create sqlmock")
	defer mockDB.Close()

	originalDB := db.DB
	db.DB = mockDB
	defer func() { db.DB = originalDB }()

	email := "existinguser@example.com"
	expectedUserID := int64(200)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(expectedUserID)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM users WHERE email = ?")).
		WithArgs(email).
		WillReturnRows(rows)

	userID, errFunc := getOrCreateUserIDByEmail(email)
	asserter.NoError(errFunc)
	asserter.Equal(expectedUserID, userID)

	asserter.NoError(mock.ExpectationsWereMet())
}
