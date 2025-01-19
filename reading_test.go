package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"tryy3.dev/weight-api-golang/api"
)

func TestWeightReadings(t *testing.T) {
	testCases := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{name: "SingleReading", values: []float64{1500}, expected: 1500},
		{name: "AverageReading", values: []float64{1500, 1001, 1005, 1002, 998, 994}, expected: 1000 /*(1001+1005+1002+998+994) / 5*/},
	}

	for _, testCase := range testCases {
		wr := api.NewWeightReadings(30, 10)

		for _, value := range testCase.values {
			wr.AddReading(value)
		}

		assert.Equal(t, wr.AverageReading(), testCase.expected)
	}
}
