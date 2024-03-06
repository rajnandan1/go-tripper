package tripper

import (
	"fmt"
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
type Monitor struct {
	Name               string  // Name of the monitor
	CircuitOpen        bool    // Indicates whether the circuit is open or closed
	LastCapturedAt     int64   // Timestamp of the last captured event
	CircuitOpenedSince int64   // Timestamp when the circuit was opened
	FailureCount       int64   // Number of failures recorded
	SuccessCount       int64   // Number of successes recorded
	Threshold          float32 // Threshold value for triggering circuit open
	ThresholdType      string  // Type of threshold (e.g., percentage, count)
	MinimumCount       int64   // Minimum number of events required for monitoring
	IntervalInSeconds  int     // Interval in seconds for monitoring (should be non-zero and multiple of 60)
}

var circuits = make(map[string]*Monitor)

// AddMonitor adds a new monitor to the circuits map.
// It checks if the monitor with the given name already exists.
// It also validates the threshold type and threshold value.
// If the minimum count or interval is invalid, it returns an error.
// If the interval is not a multiple of 60, it returns an error.
// Parameters:
//   - monitor: The monitor to be added.
//
// Returns:
//   - error: An error if the monitor is invalid or if a monitor with the same name already exists.
func AddMonitor(monitor Monitor) error {
	if _, ok := circuits[monitor.Name]; ok {
		return fmt.Errorf("Monitor with name %s already exists", monitor.Name)
	}

	// check if the threshold type is valid
	validThresholdType := false
	for _, thType := range thresholdTypes {
		if thType == monitor.ThresholdType {
			validThresholdType = true
			break
		}
	}
	if !validThresholdType {
		return fmt.Errorf("invalid threshold type %s", monitor.ThresholdType)
	}
	//if the threshold type is percentage, check if the threshold is between 0 and 100
	if monitor.ThresholdType == ThresholdPercentage && (monitor.Threshold < 0 || monitor.Threshold > 100) {
		return fmt.Errorf("invalid threshold value %f for percentage type", monitor.Threshold)
	}
	// if the threshold type is count, check if the threshold is greater than 0
	if monitor.ThresholdType == ThresholdCount && monitor.Threshold <= 0 {
		return fmt.Errorf("invalid threshold value %f for count type", monitor.Threshold)
	}

	// if the minimum count is less than 1, return an error
	if monitor.MinimumCount < 1 {
		return fmt.Errorf("invalid minimum count %d", monitor.MinimumCount)
	}

	// if the interval is less than 60, return an error
	if monitor.IntervalInSeconds < 60 {
		return fmt.Errorf("invalid interval %d", monitor.IntervalInSeconds)
	}

	// if the interval is not a multiple of 60, return an error
	if monitor.IntervalInSeconds%60 != 0 {
		return fmt.Errorf("invalid interval %d", monitor.IntervalInSeconds)
	}

	circuits[monitor.Name] = &monitor
	return nil
}

// updateStatusByMonitor updates the status of a monitor based on the success or failure of a request.
// It takes a pointer to a Monitor struct and a boolean value indicating the success of the request.
// If the current timestamp is outside the monitor's interval, the success and failure counts are reset.
// The last captured timestamp is updated to the current timestamp.
// If the request is successful, the success count is incremented; otherwise, the failure count is incremented.
// The circuit open flag is set to false.
// If the total count of successes and failures is less than the minimum count, the function returns.
// If the threshold type is ThresholdCount, and the failure count is greater than the threshold,
// the circuit open flag is set to true and the circuit opened since timestamp is updated.
// If the threshold type is ThresholdPercentage, and the percentage of failures is greater than the threshold,
// the circuit open flag is set to true and the circuit opened since timestamp is updated.
func updateStatusByMonitor(monitor *Monitor, success bool) {

	currentTs := getStartOfIntervalTimestamp(monitor)

	//reset if not within a minute
	if currentTs-monitor.LastCapturedAt >= int64(monitor.IntervalInSeconds) {
		monitor.SuccessCount = 0
		monitor.FailureCount = 0
		monitor.CircuitOpenedSince = 0
	}

	monitor.LastCapturedAt = currentTs

	if success {
		monitor.SuccessCount++
	} else {
		monitor.FailureCount++
	}

	monitor.CircuitOpen = false

	//if not yet minimum retur
	if monitor.SuccessCount+monitor.FailureCount < monitor.MinimumCount {
		return
	}

	if monitor.ThresholdType == ThresholdCount {
		if float32(monitor.FailureCount) > monitor.Threshold {
			monitor.CircuitOpen = true
			monitor.CircuitOpenedSince = monitor.LastCapturedAt
		}
	} else {
		// if the threshold type is percentage, check if the percentage of failures is greater than the threshold
		totalRequests := monitor.FailureCount + monitor.SuccessCount
		failurePercentage := (monitor.FailureCount * 100) / totalRequests
		if float32(failurePercentage) > monitor.Threshold {
			monitor.CircuitOpen = true
			monitor.CircuitOpenedSince = monitor.LastCapturedAt
		}
	}
}

// UpdateMonitorStatusByName updates the status of a monitor by its name.
// It takes the name of the monitor and a boolean indicating the success status.
// It returns the updated monitor and an error, if any.
func UpdateMonitorStatusByName(name string, success bool) (*Monitor, error) {
	monitor, err := GetMonitorByName(name)
	if err != nil {
		return nil, err
	}
	updateStatusByMonitor(monitor, success)
	return monitor, nil
}

// IsCircuitOpen checks if a circuit with the given name is open or closed.
// It returns a boolean indicating whether the circuit is open or closed, and an error if the circuit does not exist.
func IsCircuitOpen(name string) (bool, error) {
	monitor, ok := circuits[name]
	if !ok {
		return false, fmt.Errorf("Monitor with name %s does not exist", name)
	}

	now := getStartOfIntervalTimestamp(monitor)
	if monitor.CircuitOpen && (now-monitor.CircuitOpenedSince >= int64(monitor.IntervalInSeconds)) {
		monitor.CircuitOpen = false
	}
	return monitor.CircuitOpen, nil
}

// GetMonitorByName retrieves a monitor by its name.
// It returns the monitor if found, otherwise it returns an error.
func GetMonitorByName(name string) (*Monitor, error) {
	monitor, ok := circuits[name]
	if !ok {
		return nil, fmt.Errorf("Monitor with name %s does not exist", name)
	}
	return monitor, nil
}

// GetAllMonitors returns a map of all monitors.
func GetAllMonitors() map[string]*Monitor {
	return circuits
}

// getStartOfIntervalTimestamp calculates the start of the interval timestamp based on the current time and the monitor's interval.
// It rounds the current time to the nearest minute, subtracts the difference between the rounded time and the current time,
// and calculates the remainder when divided by the monitor's interval in seconds. It then returns the start of the interval timestamp.
func getStartOfIntervalTimestamp(monitor *Monitor) int64 {
	currentTime := time.Now()
	roundedTime := currentTime.Round(time.Minute).Unix()
	diff := roundedTime - currentTime.Unix()
	if diff > 0 {
		roundedTime = roundedTime - 60
	}
	remainder := roundedTime % int64(monitor.IntervalInSeconds)
	startOfInterval := time.Unix(roundedTime-remainder, 0)
	return startOfInterval.Unix()
}
