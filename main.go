package main

import (
	"log"
	"net/http"

	"github.com/WOLFnik5/weather_subscriber/db"
	"github.com/WOLFnik5/weather_subscriber/router"
)

func main() {
	// Підключення до бази
	err := db.Connect()
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}

	// Запуск роутера
	r := router.SetupRouter()
	log.Println("Server is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
