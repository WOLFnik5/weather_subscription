package handler

import (
	"encoding/json"
	"net/http"

	"github.com/WOLFnik5/weather_subscriber/model"
)

type subscriptionInput struct {
	Email     string `json:"email"`
	CityID    int64  `json:"city_id"`
	Frequency string `json:"frequency"`
}

func HandleCreateSubscription(w http.ResponseWriter, r *http.Request) {
	var input subscriptionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	if input.Email == "" || input.CityID == 0 || input.Frequency == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	if err := model.ValidateFrequency(input.Frequency); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sub := &model.Subscription{
		Email:     input.Email,
		CityID:    input.CityID,
		Frequency: input.Frequency,
	}

	if err := model.CreateSubscription(sub); err != nil {
		http.Error(w, "Failed to create subscription: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sub)
}

func HandleListSubscriptions(w http.ResponseWriter, r *http.Request) {
	subs, err := model.ListSubscriptions()
	if err != nil {
		http.Error(w, "Failed to retrieve subscriptions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subs)
}
