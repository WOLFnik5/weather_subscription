package router

import (
	"net/http"

	"github.com/WOLFnik5/weather_subscriber/handler"
	"github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}).Methods("GET")

	r.HandleFunc("/subscriptions", handler.HandleCreateSubscription).Methods("POST")
	r.HandleFunc("/subscriptions", handler.HandleListSubscriptions).Methods("GET")
	//r.HandleFunc("/cities", handler.HandleListCities).Methods("GET")

	return r
}
