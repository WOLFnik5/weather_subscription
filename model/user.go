package model

import (
	"database/sql"

	"github.com/WOLFnik5/weather_subscriber/db"
)

type User struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

func getOrCreateUserIDByEmail(email string) (int64, error) {
	var id int64
	err := db.DB.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&id)
	if err == sql.ErrNoRows {
		res, err := db.DB.Exec("INSERT INTO users (email) VALUES (?)", email)
		if err != nil {
			return 0, err
		}
		userID, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}
		return userID, nil
	} else if err != nil {
		return 0, err
	}
	return id, nil
}
