package tripper

import (
	"fmt"
	"sync"
	"time"
)

// threshold type can be only COUNT or PERCENTAGE
// ThresholdCount represents a threshold type based on count.
const (
	ThresholdCount       = "COUNT"
	ThresholdPercentage  = "PERCENTAGE"
	ThresholdConsecutive = "CONSECUTIVE"
)

var thresholdTypes = []string{ThresholdCount, ThresholdPercentage, ThresholdConsecutive}

// Circuit represents a monitoring entity that tracks the status of a circuit.
type Circuit interface {
	UpdateStatus(success bool)
	IsCircuitOpen() bool
	Data() CircuitData
}

// CircuitOptions represents options for configuring a Circuit.
type CircuitOptions struct {
	Name              string  // Name of the circuit
	Threshold         float32 // Threshold value for triggering circuit open
	ThresholdType     string  // Type of threshold (e.g., percentage, count)
	MinimumCount      int64   // Minimum number of events required for monitoring
	IntervalInSeconds int     // Interval in seconds for monitoring (should be non-zero and multiple of 60)
	OnCircuitOpen     func(t CallbackEvent)
	OnCircuitClosed   func(t CallbackEvent)
}
type CircuitData struct {
	SuccessCount       int64
	FailureCount       int64
	IsCircuitOpen      bool
	CircuitOpenedSince int64
}

// CircuitImplementation represents the implementation of the Circuit interface.
type CircuitImplementation struct {
	Options            CircuitOptions
	FailureCount       int64 // Number of failures recorded
	SuccessCount       int64 // Number of successes recorded
	CircuitOpen        bool  // Indicates whether the circuit is open or closed
	LastCapturedAt     int64 // Timestamp of the last captured event
	CircuitOpenedSince int64 // Timestamp when the circuit was opened
	ConsecutiveCounter int64
	Ticker             *time.Ticker
	Mutex              sync.Mutex
	XMutex             sync.Mutex
}

// CallbackEvent represents an event callback for the circuit.
type CallbackEvent struct {
	Timestamp    int64
	SuccessCount int64
	FailureCount int64
}

func (m *CircuitImplementation) Data() CircuitData {
	return CircuitData{
		SuccessCount:       m.SuccessCount,
		FailureCount:       m.FailureCount,
		IsCircuitOpen:      m.CircuitOpen,
		CircuitOpenedSince: m.CircuitOpenedSince,
	}
}

// ConfigureCircuit creates and configures a new Circuit with the provided options.
func ConfigureCircuit(monitorOptions CircuitOptions) (Circuit, error) {
	validThresholdType := false
	for _, thType := range thresholdTypes {
		if thType == monitorOptions.ThresholdType {
			validThresholdType = true
			break
		}
	}
	if !validThresholdType {
		return nil, fmt.Errorf("invalid threshold type %s", monitorOptions.ThresholdType)
	}
	//if the threshold type is percentage, check if the threshold is between 0 and 100
	if monitorOptions.ThresholdType == ThresholdPercentage && (monitorOptions.Threshold < 0 || monitorOptions.Threshold > 100) {
		return nil, fmt.Errorf("invalid threshold value %f for percentage type", monitorOptions.Threshold)
	}
	// if the threshold type is count, check if the threshold is greater than 0
	if monitorOptions.ThresholdType == ThresholdCount && monitorOptions.Threshold <= 0 {
		return nil, fmt.Errorf("invalid threshold value %f for count type", monitorOptions.Threshold)
	}

	// if the minimum count is less than 1, return an error
	if monitorOptions.MinimumCount < 1 {
		return nil, fmt.Errorf("invalid minimum count %d", monitorOptions.MinimumCount)
	}

	//if threshold is type count then minimum count should be greater than threshold
	if monitorOptions.ThresholdType == ThresholdCount && monitorOptions.MinimumCount <= int64(monitorOptions.Threshold) {
		return nil, fmt.Errorf("minimum count should be greater than threshold")
	}

	// if the interval is less than 5, return an error
	if monitorOptions.IntervalInSeconds < 5 {
		return nil, fmt.Errorf("invalid interval %d", monitorOptions.IntervalInSeconds)
	}

	newMonitor := &CircuitImplementation{
		Options:            monitorOptions,
		ConsecutiveCounter: 0,
	}
	newMonitor.Ticker = time.NewTicker(time.Duration(monitorOptions.IntervalInSeconds) * time.Second)
	go func() {

		for range newMonitor.Ticker.C {

			newMonitor.SuccessCount = 0
			newMonitor.FailureCount = 0
			newMonitor.CircuitOpenedSince = 0
			newMonitor.ConsecutiveCounter = 0
			newMonitor.CircuitOpen = false
			if newMonitor.Options.OnCircuitClosed != nil {
				newMonitor.Options.OnCircuitClosed(CallbackEvent{
					Timestamp:    getTimestamp(),
					SuccessCount: newMonitor.SuccessCount,
					FailureCount: newMonitor.FailureCount,
				})
			}
		}
	}()
	return newMonitor, nil

}

// UpdateStatus updates the status of the Circuit based on the success of the event.
func (m *CircuitImplementation) UpdateStatus(success bool) {
	//add a lock here
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	m.LastCapturedAt = getTimestamp()
	if success {
		m.ConsecutiveCounter = 0
		m.SuccessCount++
	} else {
		m.ConsecutiveCounter++
		m.FailureCount++
	}
	if m.SuccessCount+m.FailureCount < m.Options.MinimumCount {
		return
	}

	currentStateOfCircuit := m.CircuitOpen

	if m.Options.ThresholdType == ThresholdCount {
		if float32(m.FailureCount) >= m.Options.Threshold {
			m.CircuitOpen = true
			m.CircuitOpenedSince = m.LastCapturedAt
		} else {
			m.CircuitOpen = false
			m.CircuitOpenedSince = 0
		}
	} else if m.Options.ThresholdType == ThresholdPercentage {
		// if the threshold type is percentage, check if the percentage of failures is greater than the threshold
		totalRequests := m.FailureCount + m.SuccessCount
		failurePercentage := (m.FailureCount * 100) / totalRequests
		if float32(failurePercentage) >= m.Options.Threshold {
			m.CircuitOpen = true
			m.CircuitOpenedSince = m.LastCapturedAt
		} else {
			m.CircuitOpen = false
			m.CircuitOpenedSince = 0

		}
	} else if m.Options.ThresholdType == ThresholdConsecutive {
		if m.ConsecutiveCounter >= int64(m.Options.Threshold) {
			m.CircuitOpen = true
			m.CircuitOpenedSince = m.LastCapturedAt
		} else {
			m.CircuitOpen = false
			m.CircuitOpenedSince = 0
		}

	}
	if currentStateOfCircuit != m.CircuitOpen && m.CircuitOpen {

		if m.Options.OnCircuitOpen != nil {

			m.Options.OnCircuitOpen(CallbackEvent{
				Timestamp:    m.LastCapturedAt,
				SuccessCount: m.SuccessCount,
				FailureCount: m.FailureCount,
			})
		}

	} else if currentStateOfCircuit != m.CircuitOpen && !m.CircuitOpen {
		if m.Options.OnCircuitClosed != nil {
			m.Options.OnCircuitClosed(CallbackEvent{
				Timestamp:    m.LastCapturedAt,
				SuccessCount: m.SuccessCount,
				FailureCount: m.FailureCount,
			})
		}

	}

}

// IsCircuitOpen returns true if the circuit is open, false otherwise.
func (m *CircuitImplementation) IsCircuitOpen() bool {
	return m.CircuitOpen
}

// getTimestamp returns the current timestamp in Unix format.
func getTimestamp() int64 {
	currentTime := time.Now()
	return currentTime.Unix()
}
