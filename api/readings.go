package api

import (
	"math"
	"slices"
	"sync"
)

type WeightReadings struct {
	readings           []float64
	mutex              sync.Mutex
	maxSize            int
	thresholdDeviation float64
}

func NewWeightReadings(maxSize int, thresholdDeviation float64) *WeightReadings {
	return &WeightReadings{
		readings:           []float64{},
		mutex:              sync.Mutex{},
		maxSize:            maxSize,
		thresholdDeviation: thresholdDeviation,
	}
}

func (wr *WeightReadings) AddReading(reading float64) {
	wr.mutex.Lock()
	defer wr.mutex.Unlock()

	// Remove the oldest reading if the collection is too large
	if len(wr.readings) > wr.maxSize {
		wr.readings = wr.readings[1:]
	}

	wr.readings = append(wr.readings, reading)
}

// Get the middle value of readings
func middleReading(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	return data[len(data)/2]
}

func (wr *WeightReadings) RawReadings() []float64 {
	wr.mutex.Lock()
	data := make([]float64, len(wr.readings))
	copy(data[:], wr.readings[:])
	wr.mutex.Unlock()

	return data
}

// Get the latest reading with different filterings
func (wr *WeightReadings) AverageReading() float64 {
	// Make a copy of the readings
	wr.mutex.Lock()
	data := make([]float64, len(wr.readings))
	copy(data[:], wr.readings[:])
	wr.mutex.Unlock()

	slices.Sort(data)

	middle := middleReading(data)

	filteredReadings := []float64{}
	for _, reading := range data {
		// Calculate the Z-Score
		diff := math.Abs(reading - middle)
		if diff <= wr.thresholdDeviation {
			filteredReadings = append(filteredReadings, reading)
		}
	}

	sum := 0.0
	count := 0
	for _, reading := range filteredReadings {
		sum += reading
		count++
	}

	return sum / float64(count)
}
