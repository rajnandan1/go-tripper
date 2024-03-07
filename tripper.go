package tripper

import (
	"fmt"
	"sync"
	"time"
)

// threshold type can be only COUNT or PERCENTAGE
// ThresholdCount represents a threshold type based on count.
const (
	ThresholdCount      = "COUNT"
	ThresholdPercentage = "PERCENTAGE"
)

var thresholdTypes = []string{ThresholdCount, ThresholdPercentage}

// Monitor represents a monitoring entity that tracks the status of a circuit.
type Monitor interface {
	Get() *MonitorImplementation
	UpdateStatus(success bool)
	IsCircuitOpen() bool
}

// TripperOptions represents options for configuring the Tripper.
type TripperOptions struct {
}

// TripperImplementation represents the implementation of the Tripper interface.
type TripperImplementation struct {
	Circuits map[string]Monitor
	Options  TripperOptions
}

// Tripper represents a circuit breaker tripper.
type Tripper interface {
	AddMonitor(monitorOptions MonitorOptions) (Monitor, error)
	GetMonitor(name string) (Monitor, error)
	GetAllMonitors() map[string]Monitor
}

// MonitorOptions represents options for configuring a Monitor.
type MonitorOptions struct {
	Name              string  // Name of the monitor
	Threshold         float32 // Threshold value for triggering circuit open
	ThresholdType     string  // Type of threshold (e.g., percentage, count)
	MinimumCount      int64   // Minimum number of events required for monitoring
	IntervalInSeconds int     // Interval in seconds for monitoring (should be non-zero and multiple of 60)
	OnCircuitOpen     func(t CallbackEvent)
	OnCircuitClosed   func(t CallbackEvent)
}

// MonitorImplementation represents the implementation of the Monitor interface.
type MonitorImplementation struct {
	Options            MonitorOptions
	FailureCount       int64 // Number of failures recorded
	SuccessCount       int64 // Number of successes recorded
	CircuitOpen        bool  // Indicates whether the circuit is open or closed
	LastCapturedAt     int64 // Timestamp of the last captured event
	CircuitOpenedSince int64 // Timestamp when the circuit was opened
	Ticker             *time.Ticker
	Mutex              sync.Mutex
}

// CallbackEvent represents an event callback for the circuit monitor.
type CallbackEvent struct {
	Timestamp    int64
	SuccessCount int64
	FailureCount int64
}

// Configure configures the Tripper with the provided options.
func Configure(p TripperOptions) Tripper {
	return &TripperImplementation{
		Options:  p,
		Circuits: make(map[string]Monitor),
	}
}

// ConfigureNewMonitor creates and configures a new Monitor with the provided options.
func ConfigureNewMonitor(p MonitorOptions) Monitor {
	return &MonitorImplementation{
		Options: p,
	}
}

// Get returns the MonitorImplementation.
func (t *MonitorImplementation) Get() *MonitorImplementation {
	return t
}

// AddMonitor adds a new Monitor with the provided options.
func (t *TripperImplementation) AddMonitor(monitorOptions MonitorOptions) (Monitor, error) {
	if _, ok := t.Circuits[monitorOptions.Name]; ok {
		return nil, fmt.Errorf("Monitor with name %s already exists", monitorOptions.Name)
	}

	// check if the threshold type is valid
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

	// if the interval is less than 60, return an error
	if monitorOptions.IntervalInSeconds < 60 {
		return nil, fmt.Errorf("invalid interval %d", monitorOptions.IntervalInSeconds)
	}

	// if the interval is not a multiple of 60, return an error
	if monitorOptions.IntervalInSeconds%60 != 0 {
		return nil, fmt.Errorf("invalid interval %d", monitorOptions.IntervalInSeconds)
	}

	// configure the monitor
	m := ConfigureNewMonitor(MonitorOptions{
		Name:              monitorOptions.Name,
		Threshold:         monitorOptions.Threshold,
		ThresholdType:     monitorOptions.ThresholdType,
		MinimumCount:      monitorOptions.MinimumCount,
		IntervalInSeconds: monitorOptions.IntervalInSeconds,
		OnCircuitOpen:     monitorOptions.OnCircuitOpen,
		OnCircuitClosed:   monitorOptions.OnCircuitClosed,
	})

	//start ticker

	t.Circuits[monitorOptions.Name] = m
	return m, nil
}

// GetMonitor returns the Monitor with the provided name.
func (t *TripperImplementation) GetMonitor(name string) (Monitor, error) {
	monitor, ok := t.Circuits[name]
	if !ok {
		return nil, fmt.Errorf("Monitor with name %s does not exist", name)
	}
	return monitor, nil
}

// UpdateStatus updates the status of the Monitor based on the success of the event.
func (m *MonitorImplementation) UpdateStatus(success bool) {
	//add a lock here
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	if m.Ticker == nil {
		m.Ticker = time.NewTicker(time.Duration(m.Options.IntervalInSeconds) * time.Second)
		go func() {
			for range m.Ticker.C {
				m.SuccessCount = 0
				m.FailureCount = 0
				m.CircuitOpenedSince = 0
				m.CircuitOpen = false
				m.Options.OnCircuitClosed(CallbackEvent{
					Timestamp:    getTimestamp(),
					SuccessCount: m.SuccessCount,
					FailureCount: m.FailureCount,
				})
			}
		}()
	}

	m.LastCapturedAt = getTimestamp()
	if success {
		m.SuccessCount++
	} else {
		m.FailureCount++
	}
	if m.SuccessCount+m.FailureCount < m.Options.MinimumCount {
		return
	}
	if m.Options.ThresholdType == ThresholdCount {
		if float32(m.FailureCount) >= m.Options.Threshold {
			m.CircuitOpen = true
			m.CircuitOpenedSince = m.LastCapturedAt
		} else {
			m.CircuitOpen = false
			m.CircuitOpenedSince = 0
		}
	} else {
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
	}
	if m.CircuitOpen {

		if m.Options.OnCircuitOpen != nil {

			m.Options.OnCircuitOpen(CallbackEvent{
				Timestamp:    m.LastCapturedAt,
				SuccessCount: m.SuccessCount,
				FailureCount: m.FailureCount,
			})
		}

	} else {
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
func (m *MonitorImplementation) IsCircuitOpen() bool {
	return m.CircuitOpen
}

// GetAllMonitors returns all the monitors in the Tripper.
func (t *TripperImplementation) GetAllMonitors() map[string]Monitor {
	return t.Circuits
}

// getTimestamp returns the current timestamp in Unix format.
func getTimestamp() int64 {
	currentTime := time.Now()
	return currentTime.Unix()
}
