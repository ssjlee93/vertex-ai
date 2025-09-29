package main

import (
	"fmt"
	"log"
	"net/http"
	"tool-server/internal"
)

func main() {
	http.HandleFunc("/getWeather", internal.GetWeatherHandler)
	port := "8080"
	fmt.Printf("Tool server listening on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
