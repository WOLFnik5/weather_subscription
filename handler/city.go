package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/WOLFnik5/weather_subscriber/model"
)

func HandleListCities(w http.ResponseWriter, r *http.Request) {
	// Отримуємо параметри limit та offset
	query := r.URL.Query()
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	limit := 10 // значення за замовчуванням
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	cities, err := model.ListCities(limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve cities: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cities)
}
