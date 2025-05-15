package model

import (
	"database/sql"
	"errors"

	"github.com/WOLFnik5/weather_subscriber/db"
)

type Subscription struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	CityID    int64  `json:"city_id"`
	Frequency string `json:"frequency"`
	CreatedAt string `json:"created_at"`
}

type User struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

func getOrCreateUserIDByEmail(email string) (int64, error) {
	var id int64
	// Спочатку пробуємо знайти користувача за email
	err := db.DB.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&id)
	if err == sql.ErrNoRows {
		// Якщо користувача немає, вставляємо нового
		res, err := db.DB.Exec("INSERT INTO users (email) VALUES (?)", email)
		if err != nil {
			return 0, err // Якщо є помилка при вставці, повертаємо її
		}
		// Повертаємо ID нового користувача
		userID, err := res.LastInsertId()
		if err != nil {
			return 0, err // Якщо не вдалося отримати LastInsertId
		}
		return userID, nil
	} else if err != nil {
		// Якщо сталася інша помилка при виконанні запиту
		return 0, err
	}
	// Якщо користувач знайдений, повертаємо його ID
	return id, nil
}

func CreateSubscription(s *Subscription) error {
	userID, err := getOrCreateUserIDByEmail(s.Email)
	if err != nil {
		return err
	}

	query := `INSERT INTO subscriptions (user_id, city_id, frequency) VALUES (?, ?, ?)`
	res, err := db.DB.Exec(query, userID, s.CityID, s.Frequency)
	if err != nil {
		return err
	}
	s.ID, err = res.LastInsertId()
	return err
}

func ListSubscriptions() ([]Subscription, error) {
	rows, err := db.DB.Query(`
		SELECT s.id, u.email, s.city_id, s.frequency, s.created_at
		FROM subscriptions s
		JOIN users u ON s.user_id = u.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subs := []Subscription{}
	for rows.Next() {
		s := Subscription{}
		err := rows.Scan(&s.ID, &s.Email, &s.CityID, &s.Frequency, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, nil
}

func ValidateFrequency(freq string) error {
	if freq != "daily" && freq != "hourly" {
		return errors.New("invalid frequency, must be 'daily' or 'hourly'")
	}
	return nil
}
