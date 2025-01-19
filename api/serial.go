package api

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tarm/serial"
)

type WeightSerial struct {
	config *serial.Config
	port   *serial.Port
}

func NewSerial() (*WeightSerial, error) {
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

	return &WeightSerial{config: config, port: port}, nil
}

func (s *WeightSerial) Close() {
	s.port.Close()
}

func (s *WeightSerial) Read(weightChan chan<- float64) {
	buf := make([]byte, 32)
	// Regex to match "Weight: {number} g" where number can be negative float
	weightRegex := regexp.MustCompile(`Weight:\s*([-]?\d*\.?\d+)\s*g`)

	for {
		n, err := s.port.Read(buf)
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

				weightChan <- weight
				// fmt.Printf("Weight: %.2f g\n", weight)
			}
		}
	}
}
