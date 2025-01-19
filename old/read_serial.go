package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tarm/serial"
)

// Create a type for our weight reading
type WeightReading struct {
	Value float64
	Time  time.Time
}

// Collection of weight readings
type WeightReadings struct {
	Readings []WeightReading
	Mutex    sync.Mutex
}

// New instance of weight readings
func NewWeightReadings() *WeightReadings {
	return &WeightReadings{
		Readings: []WeightReading{},
		Mutex:    sync.Mutex{},
	}
}

// Add a new weight reading to the collection
func (wr *WeightReadings) AddReading(reading WeightReading) {
	wr.Mutex.Lock()
	defer wr.Mutex.Unlock()

	// Remove the oldest reading if the collection is too large
	if len(wr.Readings) > 30 {
		wr.Readings = wr.Readings[1:]
	}

	// Add the reading to the collection
	wr.Readings = append(wr.Readings, reading)
}

// Global variable to store latest reading
var latestReading WeightReading

func initializeSerial() (*serial.Port, error) {
	config := &serial.Config{
		Name:        "/dev/ttyACM0",
		Baud:        9600,
		ReadTimeout: time.Millisecond * 100,
	}

	port, err := serial.OpenPort(config)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port: %v", err)
	}

	// Flush the buffer by reading until no more data
	buf := make([]byte, 128)
	for {
		n, err := port.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				break // EOF means no more data to read
			}
			return nil, fmt.Errorf("failed to flush serial port: %v", err)
		}
		if n == 0 {
			break // No more data to read, buffer is empty
		}
	}

	return port, nil
}

func readSerial(port *serial.Port, weightChan chan<- WeightReading) {
	buf := make([]byte, 128)
	// Regex to match "Weight: {number} g" where number can be negative float
	weightRegex := regexp.MustCompile(`Weight:\s*([-]?\d*\.?\d+)\s*g`)

	for {
		n, err := port.Read(buf)
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("Failed to read from serial port: %v", err)
				continue
			}
		}
		if n > 0 {
			data := string(buf[:n])
			// Clean the data (remove any extra whitespace/newlines)
			data = strings.TrimSpace(data)

			// Check if data matches expected format
			matches := weightRegex.FindStringSubmatch(data)
			if len(matches) == 2 { // matches[0] is full string, matches[1] is the number
				weight, err := strconv.ParseFloat(matches[1], 64)
				if err != nil {
					log.Printf("Failed to parse weight value: %v", err)
					continue
				}

				// Correction on the weight, don't need negative numbers
				if weight < 0 {
					weight = 0
				}

				weightChan <- WeightReading{
					Value: weight,
					Time:  time.Now(),
				}
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func setupWebserver(weightChan <-chan WeightReading) {
	// Handler for getting the latest weight
	http.HandleFunc("/api/v1/weight", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(latestReading)
	})

	// Start the server
	log.Printf("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func main() {
	weightReadings := NewWeightReadings()

	port, err := initializeSerial()
	if err != nil {
		log.Fatal(err)
	}
	defer port.Close()

	weightChan := make(chan WeightReading, 1)

	// Start the serial reading goroutine
	go readSerial(port, weightChan)

	// Start a goroutine to update the latest reading
	go func() {
		for weight := range weightChan {
			weightReadings.AddReading(weight)
			fmt.Printf("Latest weight at %v: %.2f g\n",
				weight.Time.Format("15:04:05"),
				weight.Value)
		}
	}()

	// Start the webserver (this will block)
	setupWebserver(weightChan)
}
