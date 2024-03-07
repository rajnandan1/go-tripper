 
# Tripper

Tripper is a circuit breaker package for Go that allows you to monitor and control the status of circuits.

![Tripper](https://github.com/rajnandan1/go-tripper/assets/16224367/ba0b329c-6aa7-4d5f-aa35-4f7d28e9e3bf)

[![Coverage Status](https://coveralls.io/repos/github/rajnandan1/go-tripper/badge.svg?branch=main)](https://coveralls.io/github/rajnandan1/go-tripper?branch=main)

## Installation

To install Tripper, use `go get`:

```shell
go get github.com/rajnandan1/go-tripper
```

## Usage

Import the Tripper package in your Go code:

```go
import "github.com/rajnandan1/go-tripper"
```

### Creating a Tripper

To create a Tripper instance, use the `Configure` function:

```go
tripper := tripper.Configure(tripper.TripperOptions{})
```

### Adding a Monitor

To add a monitor to the Tripper, use the `AddMonitor` function:
#### Monitor With Count
```go
//Adding a monitor that will trip the circuit if count of failure is more than 10 in 1 minute
//for a minimum of 100 count
monitorOptions := tripper.MonitorOptions{
    Name:              "example-monitor",
    Threshold:         10,
    ThresholdType:     tripper.ThresholdCount,
    MinimumCount:      100,
    IntervalInSeconds: 60,
    OnCircuitOpen:     onCircuitOpenCallback,
    OnCircuitClosed:   onCircuitClosedCallback,
}

monitor, err := tripper.AddMonitor(monitorOptions)
if err != nil {
    fmt.Println("Failed to add monitor:", err)
    return
}
```

#### Monitor With Percentage
```go
//Adding a monitor that will trip the circuit if count of failure is more than 10% in 1 minute
//for a minimum of 100 count
monitorOptions := tripper.MonitorOptions{
    Name:              "example-monitor",
    Threshold:         10,
    ThresholdType:     tripper.ThresholdCount,
    MinimumCount:      100,
    IntervalInSeconds: 60,
    OnCircuitOpen:     onCircuitOpenCallback,
    OnCircuitClosed:   onCircuitClosedCallback,
}

monitor, err := tripper.AddMonitor(monitorOptions)
if err != nil {
    fmt.Println("Failed to add monitor:", err)
    return
}
```
#### Monitor with Callbacks
```go
func onCircuitOpenCallback(x tripper.CallbackEvent){
    fmt.Println("Callback OPEN")
	fmt.Println(x.FailureCount)
	fmt.Println(x.SuccessCount)
	fmt.Println(x.Timestamp)
}
func onCircuitClosedCallback(x tripper.CallbackEvent){
    fmt.Println("Callback Closed")
	fmt.Println(x.FailureCount)
	fmt.Println(x.SuccessCount)
	fmt.Println(x.Timestamp)
}
monitorOptions := tripper.MonitorOptions{
    Name:              "example-monitor",
    Threshold:         10,
    ThresholdType:     tripper.ThresholdCount,
    MinimumCount:      100,
    IntervalInSeconds: 60,
    OnCircuitOpen:     onCircuitOpenCallback,
    OnCircuitClosed:   onCircuitClosedCallback,
}

monitor, err := tripper.AddMonitor(monitorOptions)
if err != nil {
    fmt.Println("Failed to add monitor:", err)
    return
}
```
### Monitor Options

| Option              | Description                                                  | Required | Type       |
|---------------------|--------------------------------------------------------------|----------|------------|
| `Name`              | The name of the monitor.                                     | Required | `string`   |
| `Threshold`         | The threshold value for the monitor.                          | Required | `float32` |
| `ThresholdType`     | The type of threshold (`ThresholdCount` or `ThresholdRatio`). | Required | `string`  |
| `MinimumCount`      | The minimum number of events required for monitoring.         | Required | `int64`   |
| `IntervalInSeconds` | The time interval for monitoring in seconds.                  | Required | `int`     |
| `OnCircuitOpen`     | Callback function called when the circuit opens.              | Optional | `func()`  |
| `OnCircuitClosed`   | Callback function called when the circuit closes.             | Optional | `func()`  |

Circuits are reset after `IntervalInSeconds`

### Getting a Monitor

To get a monitor from the Tripper, use the `GetMonitor` function:

```go
monitor, err := tripper.GetMonitor("example-monitor")
if err != nil {
    fmt.Println("Monitor not found:", err)
    return
}
```

### Updating Monitor Status

To update the status of a monitor based on the success of an event, use the `UpdateStatus` function:

```go
monitor.UpdateStatus(true)  // Success event
monitor.UpdateStatus(false) // Failure event
```

### Checking Circuit Status

To check if a circuit is open or closed, use the `IsCircuitOpen` function:

```go
isOpen := monitor.IsCircuitOpen()
if isOpen {
    fmt.Println("Circuit is open")
} else {
    fmt.Println("Circuit is closed")
}
```

### Getting All Monitors

To get all monitors in the Tripper, use the `GetAllMonitors` function:

```go
monitors := tripper.GetAllMonitors()
for name, monitor := range monitors {
    fmt.Println("Monitor:", name)
    fmt.Println("Circuit Open:", monitor.IsCircuitOpen())
    fmt.Println("Success Count:", monitor.Get().SuccessCount)
    fmt.Println("Failure Count:", monitor.Get().FailureCount)
    fmt.Println()
}
```

### Example: HTTP Request with Circuit Breaker

Here's an example of using Tripper to handle HTTP requests with a circuit breaker:

```go
import (
    "fmt"
    "net/http"
    "time"

    "github.com/your-username/tripper"
)

func main() {
    // Create a Tripper instance
    tripper := tripper.Configure(tripper.TripperOptions{})

    // Add a monitor for the HTTP request circuit
    monitorOptions := tripper.MonitorOptions{
        Name:              "http-request-monitor",
        Threshold:         10,
        ThresholdType:     tripper.ThresholdCount,
        MinimumCount:      100,
        IntervalInSeconds: 60,
        OnCircuitOpen: func(t tripper.CallbackEvent) {
            fmt.Println("Circuit is open for HTTP requests. Triggered at:", time.Unix(t.Timestamp, 0))
        },
        OnCircuitClosed: func(t tripper.CallbackEvent) {
            fmt.Println("Circuit is closed for HTTP requests. Triggered at:", time.Unix(t.Timestamp, 0))
        },
    }

    _, err := tripper.AddMonitor(monitorOptions)
    if err != nil {
        fmt.Println("Failed to add monitor:", err)
        return
    }

    // Make an HTTP request with circuit breaker protection
    client := http.Client{
        Transport: tripper.NewTransport(http.DefaultTransport),
    }

    resp, err := client.Get("https://api.example.com/data")
    if err != nil {
        fmt.Println("HTTP request failed:", err)
        return
    }
    defer resp.Body.Close()

    // Process the response
    // ...

    fmt.Println("HTTP request successful!")
}
```

## License

This project is licensed under the [MIT License](LICENSE).
 