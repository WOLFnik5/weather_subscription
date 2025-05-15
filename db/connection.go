package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Connect() error {
	var err error
	dsn := "root:1111@tcp(db:3306)/weather_service?parseTime=true"
	for i := 0; i < 10; i++ {
		DB, err = sql.Open("mysql", dsn)
		if err == nil {
			err = DB.Ping()
			if err == nil {
				return err
			}
		}
		fmt.Println("Waiting for DB... retrying in 2 seconds")
		time.Sleep(2 * time.Second)
	}
	return err
}
