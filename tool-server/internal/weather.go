package internal

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
)

// A simple struct to represent our weather data
type WeatherInfo struct {
	Location     string `json:"location"`
	Temperature  int    `json:"temperature_celsius"`
	Condition    string `json:"condition"`
	WindSpeedKPH int    `json:"wind_speed_kph"`
}

// GetWeatherHandler handles the API request for weather
func GetWeatherHandler(w http.ResponseWriter, r *http.Request) {
	// Get the location from the query parameter
	location := r.URL.Query().Get("location")
	if location == "" {
		log.Println("Tool Server: Missing 'location' query parameter")
		http.Error(w, "Missing 'location' query parameter", http.StatusBadRequest)
		return
	}

	log.Printf("Tool Server: Received request for weather in %s\n", location)

	// --- In a real app, you would call a real weather API here ---
	// For this example, we'll just generate some fake data.
	weather := WeatherInfo{
		Location:     location,
		Temperature:  rand.Intn(25) + 5, // Random temp between 5 and 29
		Condition:    []string{"Sunny", "Cloudy", "Rainy", "Windy"}[rand.Intn(4)],
		WindSpeedKPH: rand.Intn(30),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(weather)
}
