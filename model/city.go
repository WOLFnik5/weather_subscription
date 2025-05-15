package model

import (
	"github.com/WOLFnik5/weather_subscriber/db"
)

type City struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Country   string `json:"country"`
	CreatedAt string `json:"created_at"`
}

func ListCities(limit, offset int) ([]City, error) {
	query := `SELECT id, name, country, created_at FROM cities ORDER BY id LIMIT ? OFFSET ?`
	rows, err := db.DB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cities []City
	for rows.Next() {
		var c City
		if err := rows.Scan(&c.ID, &c.Name, &c.Country, &c.CreatedAt); err != nil {
			return nil, err
		}
		cities = append(cities, c)
	}

	return cities, nil
}
