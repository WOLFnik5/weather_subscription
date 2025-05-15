package model

import (
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

	var subs []Subscription
	for rows.Next() {
		var s Subscription
		if err := rows.Scan(&s.ID, &s.Email, &s.CityID, &s.Frequency, &s.CreatedAt); err != nil {
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
