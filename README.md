# Package tripper

The `tripper` is a lightweight, 0 dependencies package that provides functionality for monitoring the status of a circuit. It allows you to track the success and failure counts of events and determine if a circuit should be open or closed based on configurable thresholds. This can be useful in scenarios where you want to implement circuit breaker patterns or manage the availability of services.

## Constants

- `ThresholdCount` represents a threshold type based on count.
- `ThresholdPercentage` represents a threshold type based on percentage.

## Types

### Monitor

The `Monitor` struct represents a monitoring entity that tracks the status of a circuit. It contains the following fields:

- `Name`: Name of the monitor.
- `CircuitOpen`: Indicates whether the circuit is open or closed.
- `LastCapturedAt`: Timestamp of the last captured event.
- `CircuitOpenedSince`: Timestamp when the circuit was opened.
- `FailureCount`: Number of failures recorded.
- `SuccessCount`: Number of successes recorded.
- `Threshold`: Threshold value for triggering circuit open.
- `ThresholdType`: Type of threshold (e.g., percentage, count).
- `MinimumCount`: Minimum number of events required for monitoring.
- `IntervalInSeconds`: Interval in seconds for monitoring (should be non-zero and a multiple of 60).

## Functions

### AddMonitor

```go
func AddMonitor(monitor Monitor) error
```

`AddMonitor` adds a new monitor to the circuits map. It checks if a monitor with the given name already exists. It also validates the threshold type and threshold value. If the minimum count or interval is invalid, it returns an error. If the interval is not a multiple of 60, it returns an error.

### UpdateMonitorStatusByName

```go
func UpdateMonitorStatusByName(name string, success bool) (*Monitor, error)
```

`UpdateMonitorStatusByName` updates the status of a monitor by its name. It takes the name of the monitor and a boolean indicating the success status. It returns the updated monitor and an error, if any.

### IsCircuitOpen

```go
func IsCircuitOpen(name string) (bool, error)
```

`IsCircuitOpen` checks if a circuit with the given name is open or closed. It returns a boolean indicating whether the circuit is open or closed, and an error if the circuit does not exist.

### GetMonitorByName

```go
func GetMonitorByName(name string) (*Monitor, error)
```

`GetMonitorByName` retrieves a monitor by its name. It returns the monitor if found, otherwise it returns an error.

### GetAllMonitors

```go
func GetAllMonitors() map[string]*Monitor
```

`GetAllMonitors` returns a map of all monitors.

## How to Use

1. Import the `tripper` package into your Go code:

```go
import "github.com/rajnandan1/go-tripper"
```

2. Create a new `Monitor` instance with the desired configuration:

```go
monitor := tripper.Monitor{
	Name:               "MyMonitor",
	Threshold:          0.5,
	ThresholdType:      tripper.ThresholdPercentage,
	MinimumCount:       10,
	IntervalInSeconds:  60,
}
```

3. Add the monitor to the circuits map using the `AddMonitor` function:

```go
err := tripper.AddMonitor(monitor)
if err != nil {
	// Handle error
}
```

4. Update the status of the monitor based on the success or failure of events using the `UpdateMonitorStatusByName` function:

```go
success := true // Or false, depending on the event result
updatedMonitor, err := tripper.UpdateMonitorStatusByName("MyMonitor", success)
if err != nil {
	// Handle error
}
```

5. Check if a circuit is open or closed using the `IsCircuitOpen` function:

```go
isOpen, err := tripper.IsCircuitOpen("MyMonitor")
if err != nil {
	// Handle error
}
if isOpen {
	// Circuit is open, take appropriate action
} else {
	// Circuit is closed, continue normal operation
}
```

6. Retrieve a monitor by its name using the `GetMonitorByName` function:

```go
monitor, err := tripper.GetMonitorByName("MyMonitor")
if err != nil {
	// Handle error
}
// Use the monitor as needed
```

7. Get a map of all monitors using the `GetAllMonitors` function:

```go
monitors := tripper.GetAllMonitors()
for name, monitor := range monitors {
	// Process each monitor
}
```
