package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Create a type for our weight reading
type WeightReading struct {
	Value float64
	Time  time.Time
}

func SetupWebserver(readings *WeightReadings) {
	// Handler for getting the latest weight
	http.HandleFunc("/api/v1/weight", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(WeightReading{
			Value: readings.AverageReading(),
			Time:  time.Now(),
		})
	})

	// Start the server
	log.Printf("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
