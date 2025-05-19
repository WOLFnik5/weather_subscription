package model

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/WOLFnik5/weather_subscriber/db"
	"github.com/stretchr/testify/assert"
)

func TestModel_getOrCreateUserIDByEmail_CreateNew(t *testing.T) {
	asserter := assert.New(t)
	mockDB, mock, err := sqlmock.New()
	asserter.NoError(err, "Failed to create sqlmock")
	defer mockDB.Close()

	originalDB := db.DB
	db.DB = mockDB
	defer func() { db.DB = originalDB }()

	email := "new.user.for.test@example.com"
	expectedNewUserID := int64(12345)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM users WHERE email = ?")).
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (email) VALUES (?)")).
		WithArgs(email).
		WillReturnResult(sqlmock.NewResult(expectedNewUserID, 1))

	userID, errFunc := getOrCreateUserIDByEmail(email)
	asserter.NoError(errFunc)
	asserter.Equal(expectedNewUserID, userID)

	asserter.NoError(mock.ExpectationsWereMet())
}

func TestModel_getOrCreateUserIDByEmail_ExistingUser(t *testing.T) {
	asserter := assert.New(t)
	mockDB, mock, err := sqlmock.New()
	asserter.NoError(err, "Failed to create sqlmock")
	defer mockDB.Close()

	originalDB := db.DB
	db.DB = mockDB
	defer func() { db.DB = originalDB }()

	email := "existing.user.for.test@example.com"
	expectedExistingUserID := int64(56789)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(expectedExistingUserID)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM users WHERE email = ?")).
		WithArgs(email).
		WillReturnRows(rows)

	userID, errFunc := getOrCreateUserIDByEmail(email)
	asserter.NoError(errFunc)
	asserter.Equal(expectedExistingUserID, userID)

	asserter.NoError(mock.ExpectationsWereMet())
}
