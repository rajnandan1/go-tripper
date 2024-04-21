 
# Tripper

Tripper is a circuit breaker package for Go that allows you to circuit and control the status of circuits.

![Tripper](https://github.com/rajnandan1/go-tripper/assets/16224367/ba0b329c-6aa7-4d5f-aa35-4f7d28e9e3bf)

[![Coverage Status](https://coveralls.io/repos/github/rajnandan1/go-tripper/badge.svg?branch=main)](https://coveralls.io/github/rajnandan1/go-tripper?branch=main) [![Go Reference](https://pkg.go.dev/badge/github.com/rajnandan1/go-tripper.svg)](https://pkg.go.dev/github.com/rajnandan1/go-tripper)

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



### Adding a Circuit

To add a circuit to the Tripper, use the `ConfigureCircuit` function:
#### Circuit With Count
```go
//Adding a circuit that will trip the circuit if count of failure is more than 10 in 1 minute
//for a minimum of 100 count
circuitOptions := tripper.CircuitOptions{
    Name:              "example-circuit",
    Threshold:         10,
    ThresholdType:     tripper.ThresholdCount,
    MinimumCount:      100,
    IntervalInSeconds: 60,
    OnCircuitOpen:     onCircuitOpenCallback,
    OnCircuitClosed:   onCircuitClosedCallback,
}

circuit, err := tripper.ConfigureCircuit(circuitOptions)
if err != nil {
    fmt.Println("Failed to add circuit:", err)
    return
}
```

#### Circuit With Percentage
```go
//Adding a circuit that will trip the circuit if count of failure is more than 10% in 1 minute
//for a minimum of 100 count
circuitOptions := tripper.CircuitOptions{
    Name:              "example-circuit",
    Threshold:         10,
    ThresholdType:     tripper.ThresholdPercentage,
    MinimumCount:      100,
    IntervalInSeconds: 60,
    OnCircuitOpen:     onCircuitOpenCallback,
    OnCircuitClosed:   onCircuitClosedCallback,
}

circuit, err := tripper.ConfigureCircuit(circuitOptions)
if err != nil {
    fmt.Println("Failed to add circuit:", err)
    return
}
```

#### Circuit With Consecutive Errors
```go
//Adding a circuit that will trip the circuit if 10 consecutive erros occur in 1 minute
//for a minimum of 100 count
circuitOptions := tripper.CircuitOptions{
    Name:              "example-circuit",
    Threshold:         10,
    ThresholdType:     tripper.ThresholdConsecutive,
    MinimumCount:      100,
    IntervalInSeconds: 60,
    OnCircuitOpen:     onCircuitOpenCallback,
    OnCircuitClosed:   onCircuitClosedCallback,
}

circuit, err := tripper.ConfigureCircuit(circuitOptions)
if err != nil {
    fmt.Println("Failed to add circuit:", err)
    return
}
```

#### Circuit with Callbacks
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
circuitOptions := tripper.CircuitOptions{
    Name:              "example-circuit",
    Threshold:         10,
    ThresholdType:     tripper.ThresholdCount,
    MinimumCount:      100,
    IntervalInSeconds: 60,
    OnCircuitOpen:     onCircuitOpenCallback,
    OnCircuitClosed:   onCircuitClosedCallback,
}

circuit, err := tripper.ConfigureCircuit(circuitOptions)
if err != nil {
    fmt.Println("Failed to add circuit:", err)
    return
}
```
### Circuit Options

| Option              | Description                                                  | Required | Type       |
|---------------------|--------------------------------------------------------------|----------|------------|
| `Name`              | The name of the circuit.                                     | Required | `string`   |
| `Threshold`         | The threshold value for the circuit.                          | Required | `float32` |
| `ThresholdType`     | The type of threshold (`ThresholdCount` or `ThresholdPercentage`or `ThresholdConsecutive`). | Required | `string`  |
| `MinimumCount`      | The minimum number of events required for monitoring.         | Required | `int64`   |
| `IntervalInSeconds` | The time interval for monitoring in seconds.                  | Required | `int`     |
| `OnCircuitOpen`     | Callback function called when the circuit opens.              | Optional | `func()`  |
| `OnCircuitClosed`   | Callback function called when the circuit closes.             | Optional | `func()`  |

Circuits are reset after `IntervalInSeconds`

### Updating Circuit Status

To update the status of a circuit based on the success of an event, use the `UpdateStatus` function:

```go
circuit.UpdateStatus(true)  // Success event
circuit.UpdateStatus(false) // Failure event
```

### Checking Circuit Status

To check if a circuit is open or closed, use the `IsCircuitOpen` function:

```go
isOpen := circuit.IsCircuitOpen()
if isOpen {
    fmt.Println("Circuit is open")
} else {
    fmt.Println("Circuit is closed")
}
```
 

### Example: HTTP Request with Circuit Breaker

Here's an example of using Tripper to handle HTTP requests with a circuit breaker:

```go
package main

import (
    "fmt"
    "time"

    "github.com/rajnandan1/go-tripper"
)

func main() {
    // Configure the circuit options
    circuitOptions := tripper.CircuitOptions{
        Name:              "MyServiceMonitor",
        Threshold:         50, // Threshold for triggering circuit open (percentage)
        ThresholdType:     tripper.ThresholdPercentage,
        MinimumCount:      10,  // Minimum number of events required for monitoring
        IntervalInSeconds: 60,  // Interval in seconds for monitoring
        OnCircuitOpen: func(event tripper.CallbackEvent) {
            fmt.Println("Circuit opened:", event)
        },
        OnCircuitClosed: func(event tripper.CallbackEvent) {
            fmt.Println("Circuit closed:", event)
        },
    }

    // Create a new circuit
    circuit, err := tripper.ConfigureCircuit(circuitOptions)
    if err != nil {
        fmt.Println("Error creating circuit:", err)
        return
    }

    // Simulate calling a service
    for i := 0; i < 20; i++ {
        success := i%2 == 0 // Simulate failures every other call
        circuit.UpdateStatus(success)
        time.Sleep(time.Second)
    }

    // Check if the circuit is open
    if circuit.IsCircuitOpen() {
        fmt.Println("Circuit is open!")
    } else {
        fmt.Println("Circuit is closed.")
    }
}
```

## License

This project is licensed under the [MIT License](LICENSE).
 