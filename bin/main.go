package main

import (
	"log"

	"tryy3.dev/weight-api-golang/api"
)

func main() {
	readings := api.NewWeightReadings(30, 20)
	serial, err := api.NewSerial()
	if err != nil {
		log.Fatal(err)
	}
	defer serial.Close()
	weightChan := make(chan float64, 1)

	go serial.Read(weightChan)

	go func() {
		for weight := range weightChan {
			readings.AddReading(weight)
			// fmt.Printf("Latest weight at %v: %.2f g - Raw: %v\n",
			// 	time.Now().Format("15:04:05"),
			// 	readings.AverageReading(),
			// 	readings.RawReadings())
		}
	}()

	api.SetupWebserver(readings)
}
