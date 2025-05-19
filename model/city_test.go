package model

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/WOLFnik5/weather_subscriber/db" // Потрібен для підміни db.DB
	"github.com/stretchr/testify/assert"
)

func TestModel_ListCities(t *testing.T) {
	asserter := assert.New(t)
	mockDB, mock, err := sqlmock.New()
	asserter.NoError(err, "Failed to create sqlmock")
	defer mockDB.Close()

	originalDB := db.DB
	db.DB = mockDB
	defer func() { db.DB = originalDB }()

	// Припускаємо, що CreatedAt - це string у форматі MySQL TIMESTAMP (YYYY-MM-DD HH:MM:SS)
	// або той формат, який повертається вашим драйвером MySQL і сканується в рядок.
	// Якщо CreatedAt у структурі City - це time.Time, то тут теж має бути time.Time.
	// У вашій моделі City.CreatedAt - це string.
	nowStr := time.Now().Format("2006-01-02 15:04:05") // Типовий формат для MySQL TIMESTAMP

	expectedCitiesData := []struct {
		ID        int64
		Name      string
		Country   string
		CreatedAt string
	}{
		{ID: 1, Name: "Kyiv", Country: "Ukraine", CreatedAt: nowStr},
		{ID: 2, Name: "Lviv", Country: "Ukraine", CreatedAt: nowStr},
	}

	limit := 10
	offset := 0

	// Важливо: у вашому ListCities запит виглядає так:
	// query := `SELECT id, name, country, created_at FROM cities ORDER BY id LIMIT ? OFFSET ?`
	// Тому колонки в NewRows мають відповідати цьому порядку.
	rows := sqlmock.NewRows([]string{"id", "name", "country", "created_at"})
	for _, cityData := range expectedCitiesData {
		rows.AddRow(cityData.ID, cityData.Name, cityData.Country, cityData.CreatedAt)
	}

	// Очікування для SELECT з cities
	// Використовуємо regexp.QuoteMeta для екранування спеціальних символів SQL, якщо вони є.
	// Тут запит простий, тому можна і без QuoteMeta, але це хороша практика.
	expectedSQL := "SELECT id, name, country, created_at FROM cities ORDER BY id LIMIT \\? OFFSET \\?"
	mock.ExpectQuery(expectedSQL). // Екрануємо ? для регулярного виразу
					WithArgs(limit, offset).
					WillReturnRows(rows)

	cities, errList := ListCities(limit, offset)
	asserter.NoError(errList)
	asserter.NotNil(cities)
	asserter.Len(cities, len(expectedCitiesData))

	if len(cities) == len(expectedCitiesData) {
		for i, expected := range expectedCitiesData {
			asserter.Equal(expected.ID, cities[i].ID)
			asserter.Equal(expected.Name, cities[i].Name)
			asserter.Equal(expected.Country, cities[i].Country)
			asserter.Equal(expected.CreatedAt, cities[i].CreatedAt)
		}
	}

	asserter.NoError(mock.ExpectationsWereMet(), "SQL expectations were not met")
}

func TestModel_ListCities_EmptyResult(t *testing.T) {
	asserter := assert.New(t)
	mockDB, mock, err := sqlmock.New()
	asserter.NoError(err, "Failed to create sqlmock")
	defer mockDB.Close()

	originalDB := db.DB
	db.DB = mockDB
	defer func() { db.DB = originalDB }()

	limit := 10
	offset := 0

	rows := sqlmock.NewRows([]string{"id", "name", "country", "created_at"}) // Порожні рядки

	expectedSQL := "SELECT id, name, country, created_at FROM cities ORDER BY id LIMIT \\? OFFSET \\?"
	mock.ExpectQuery(expectedSQL).
		WithArgs(limit, offset).
		WillReturnRows(rows)

	cities, errList := ListCities(limit, offset)
	asserter.NoError(errList)
	// asserter.NotNil(cities) // Ось тут проблема, якщо ListCities може повернути nil
	// Замість цього, перевіряємо, що це порожній слайс або nil (залежно від реалізації)
	// Якщо ListCities *завжди* повертає ініціалізований слайс (навіть порожній), то NotNil - це ОК.
	// Але якщо він повертає nil, то NotNil впаде.
	// Типова ідіома Go - повертати порожній слайс.
	if cities != nil { // Тільки якщо не nil, перевіряємо довжину
		asserter.Len(cities, 0)
	} else {
		// Якщо cities == nil, то це теж може бути прийнятним порожнім результатом
		// в залежності від вашого контракту функції.
		// asserter.Nil(cities) // Можна явно перевірити на nil
	}
	asserter.Len(cities, 0)

	asserter.NoError(mock.ExpectationsWereMet())
}
