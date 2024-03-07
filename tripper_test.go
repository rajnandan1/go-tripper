package tripper

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddMonitor(t *testing.T) {
	// Test case 1: Add a new monitor successfully
	// Expected output: No error
	monitorOptions := MonitorOptions{
		Name:              "test",
		Threshold:         65,
		MinimumCount:      20,
		IntervalInSeconds: 120,
		ThresholdType:     ThresholdPercentage,
	}
	tripperOpts := TripperOptions{}
	tripper := Configure(tripperOpts)
	m, err := tripper.AddMonitor(monitorOptions)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	r, errOk := tripper.GetMonitor("test")
	assert.NoError(t, errOk)
	assert.Equal(t, r, m)

	// Test case 2: Add a monitor with an existing name
	// Expected output: An error
	_, err = tripper.AddMonitor(monitorOptions)
	assert.Error(t, err)
	assert.EqualError(t, err, "Monitor with name test already exists")
	// Test case 3: Add a monitor with an invalid threshold type
	// Expected output: An error
	monitorOptions.ThresholdType = "invalid"
	monitorOptions.Name = "invalid"
	_, err = tripper.AddMonitor(monitorOptions)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid threshold type invalid")
	// Test case 4: Add a monitor with an invalid threshold value for percentage type
	// Expected output: An error
	monitorOptions.ThresholdType = ThresholdPercentage
	monitorOptions.Threshold = -1
	_, err = tripper.AddMonitor(monitorOptions)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid threshold value -1.000000 for percentage type")
	// Test case 5: Add a monitor with an invalid threshold value for count type
	// Expected output: An error
	monitorOptions.ThresholdType = ThresholdCount
	monitorOptions.Threshold = 0
	_, err = tripper.AddMonitor(monitorOptions)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid threshold value 0.000000 for count type")
	// Test case 6: Add a monitor with an invalid minimum count
	// Expected output: An error
	monitorOptions.ThresholdType = ThresholdPercentage
	monitorOptions.Threshold = 65
	monitorOptions.MinimumCount = 0
	_, err = tripper.AddMonitor(monitorOptions)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid minimum count 0")
	// Test case 7: Add a monitor with an invalid interval
	// Expected output: An error
	monitorOptions.MinimumCount = 20
	monitorOptions.IntervalInSeconds = 59
	_, err = tripper.AddMonitor(monitorOptions)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid interval 59")
	// Test case 8: Add a monitor with an interval that is not a multiple of 60
	// Expected output: An error
	monitorOptions.IntervalInSeconds = 61
	_, err = tripper.AddMonitor(monitorOptions)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid interval 61")
	// Test case 9: Add a monitor with a valid threshold type and threshold value
	// Expected output: No error
	monitorOptions.IntervalInSeconds = 120
	monitorOptions.Name = "test2"
	monitorOptions.ThresholdType = ThresholdCount
	monitorOptions.Threshold = 10
	_, err = tripper.AddMonitor(monitorOptions)
	assert.NoError(t, err)

	// Test case 10: Add a monitor with a  threshold type count and minimum value less than threshold
	// Expected output: An error
	monitorOptions.MinimumCount = 5
	monitorOptions.Threshold = 6
	monitorOptions.Name = "test10"
	_, err = tripper.AddMonitor(monitorOptions)
	assert.Error(t, err)
	assert.EqualError(t, err, "minimum count should be greater than threshold")
}

func TestUpdateStatus(t *testing.T) {
	// Test case 1: Update status with success=true
	// Expected output: Success count incremented
	callBackCalledOpen := false
	callBackCalledClosed := false
	monitorOptions := MonitorOptions{
		Name:              "test",
		Threshold:         50,
		MinimumCount:      4,
		IntervalInSeconds: 60,
		ThresholdType:     ThresholdPercentage,
		OnCircuitOpen: func(x CallbackEvent) {
			callBackCalledOpen = true
		},
		OnCircuitClosed: func(x CallbackEvent) {
			callBackCalledClosed = true
		},
	}
	tripperOpts := TripperOptions{}
	tripper := Configure(tripperOpts)
	m, err := tripper.AddMonitor(monitorOptions)
	assert.NoError(t, err)
	assert.NotNil(t, m)

	m.UpdateStatus(true)
	assert.Equal(t, int64(1), m.Get().SuccessCount)
	assert.Equal(t, int64(0), m.Get().FailureCount)

	// Test case 2: Update status with success=false
	// Expected output: Failure count incremented
	m.UpdateStatus(false)
	assert.Equal(t, int64(1), m.Get().SuccessCount)
	assert.Equal(t, int64(1), m.Get().FailureCount)

	// Test case 3: Update status with success=true and minimum count not reached
	// Expected output: Success count incremented, circuit not opened

	m.UpdateStatus(true)
	assert.Equal(t, int64(2), m.Get().SuccessCount)
	assert.Equal(t, int64(1), m.Get().FailureCount)
	assert.False(t, m.Get().CircuitOpen)

	// Test case 4: Update status with success=false and minimum count not reached
	// Expected output: Failure count incremented, circuit not opened
	m.UpdateStatus(false)
	assert.Equal(t, int64(2), m.Get().SuccessCount)
	assert.Equal(t, int64(2), m.Get().FailureCount)
	assert.True(t, m.Get().CircuitOpen)
	assert.True(t, callBackCalledOpen)
	assert.NotZero(t, m.Get().CircuitOpenedSince)

	// Test case 5: Update status with success=false and failure count greater than threshold (ThresholdType=ThresholdCount)
	// Expected output: Failure count incremented, circuit opened
	m.UpdateStatus(false)
	assert.Equal(t, int64(2), m.Get().SuccessCount)
	assert.Equal(t, int64(3), m.Get().FailureCount)
	assert.True(t, m.Get().CircuitOpen)
	assert.NotZero(t, m.Get().CircuitOpenedSince)

	m.UpdateStatus(true)
	m.UpdateStatus(true)
	m.UpdateStatus(true)
	m.UpdateStatus(true)
	m.UpdateStatus(true)
	assert.False(t, m.Get().CircuitOpen)
	assert.True(t, callBackCalledClosed)

	// Test case 6: Update status with success=false and failure percentage greater than threshold (ThresholdType=ThresholdPercentage)
	// Expected output: Failure count incremented, circuit opened
	m.Get().SuccessCount = 0
	m.Get().FailureCount = 0
	m.Get().Options.ThresholdType = ThresholdCount
	m.Get().Options.Threshold = 2 //two failures would trigger the circuit
	m.Get().Options.MinimumCount = 2
	m.Get().CircuitOpen = false
	m.Get().CircuitOpenedSince = 0

	m.UpdateStatus(false)
	assert.Equal(t, int64(0), m.Get().SuccessCount)
	assert.Equal(t, int64(1), m.Get().FailureCount)
	assert.False(t, m.Get().CircuitOpen)

	m.UpdateStatus(false)
	m.UpdateStatus(false)
	m.UpdateStatus(true)
	assert.Equal(t, int64(1), m.Get().SuccessCount)
	assert.Equal(t, int64(3), m.Get().FailureCount)
	assert.True(t, m.Get().CircuitOpen)
	assert.NotZero(t, m.Get().CircuitOpenedSince)

	m.Get().Options.Threshold = 20
	m.UpdateStatus(true)
	assert.False(t, m.Get().CircuitOpen)
	assert.True(t, callBackCalledClosed)
}

func TestInterval(t *testing.T) {
	callBackCalledClosed := false
	monitorOptionsX := MonitorOptions{
		Name:              "x",
		Threshold:         2,
		MinimumCount:      4,
		IntervalInSeconds: 60,
		ThresholdType:     ThresholdCount,
		OnCircuitClosed: func(x CallbackEvent) {
			callBackCalledClosed = true
		},
	}
	tripperOptsX := TripperOptions{}
	tripperX := Configure(tripperOptsX)
	mx, err := tripperX.AddMonitor(monitorOptionsX)
	assert.NoError(t, err)
	mx.Get().Options.IntervalInSeconds = 1
	mx.UpdateStatus(false)
	mx.UpdateStatus(false)
	mx.UpdateStatus(true)
	mx.UpdateStatus(false)
	assert.Equal(t, int64(1), mx.Get().SuccessCount)
	assert.Equal(t, int64(3), mx.Get().FailureCount)
	assert.True(t, mx.Get().CircuitOpen)
	assert.NotZero(t, mx.Get().CircuitOpenedSince)
	time.Sleep(1100 * time.Millisecond)
	assert.Equal(t, int64(0), mx.Get().SuccessCount)
	assert.Equal(t, int64(0), mx.Get().FailureCount)
	assert.False(t, mx.Get().CircuitOpen)
	assert.True(t, callBackCalledClosed)
	assert.Zero(t, mx.Get().CircuitOpenedSince)
}

func TestUpdateStatusRaceSingle(t *testing.T) {
	monitorOptionsX := MonitorOptions{
		Name:              "x",
		Threshold:         2,
		MinimumCount:      4,
		IntervalInSeconds: 60,
		ThresholdType:     ThresholdCount,
	}
	tripperOptsX := TripperOptions{}
	tripperX := Configure(tripperOptsX)
	_, err := tripperX.AddMonitor(monitorOptionsX)
	assert.NoError(t, err)
	mx, _ := tripperX.GetMonitor("x")
	numGoroutines := 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {

			mx.UpdateStatus((mx.Get().SuccessCount+mx.Get().FailureCount)%2 == 0)
			// Decrement the wait group counter
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Equal(t, int64(100), mx.Get().SuccessCount+mx.Get().FailureCount)
}

func TestUpdateStatusRaceTwo(t *testing.T) {
	monitorOptionsX := MonitorOptions{
		Name:              "x",
		Threshold:         2,
		MinimumCount:      4,
		IntervalInSeconds: 60,
		ThresholdType:     ThresholdCount,
	}
	tripperOptsX := TripperOptions{}
	tripperX := Configure(tripperOptsX)
	_, err := tripperX.AddMonitor(monitorOptionsX)
	assert.NoError(t, err)
	mx, _ := tripperX.GetMonitor("x")

	monitorOptionsX1 := MonitorOptions{
		Name:              "x1",
		Threshold:         2,
		MinimumCount:      4,
		IntervalInSeconds: 60,
		ThresholdType:     ThresholdCount,
	}

	_, errx1 := tripperX.AddMonitor(monitorOptionsX1)
	assert.NoError(t, errx1)
	mx1, _ := tripperX.GetMonitor("x1")

	numGoroutines := 100
	status := false
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {

			status = !status
			mx.UpdateStatus(status)
			mx1.UpdateStatus(status)
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Equal(t, int64(200), mx.Get().SuccessCount+mx.Get().FailureCount+mx1.Get().SuccessCount+mx1.Get().FailureCount)

}
func TestGetMonitor(t *testing.T) {
	tripperOpts := TripperOptions{}
	tripper := Configure(tripperOpts)

	// Test case 1: Get an existing monitor
	// Expected output: Monitor and no error
	monitorOptions := MonitorOptions{
		Name:              "test",
		Threshold:         65,
		MinimumCount:      20,
		IntervalInSeconds: 120,
		ThresholdType:     ThresholdPercentage,
	}
	_, err := tripper.AddMonitor(monitorOptions)
	assert.NoError(t, err)

	m, err := tripper.GetMonitor("test")
	assert.NoError(t, err)
	assert.NotNil(t, m)

	// Test case 2: Get a non-existing monitor
	// Expected output: Error
	_, err = tripper.GetMonitor("non-existing")
	assert.Error(t, err)
	assert.EqualError(t, err, "Monitor with name non-existing does not exist")
}
func TestIsCircuitOpen(t *testing.T) {
	// Test case 1: Circuit is closed
	// Expected output: false
	monitorOptions := MonitorOptions{
		Name:              "test",
		Threshold:         65,
		MinimumCount:      20,
		IntervalInSeconds: 120,
		ThresholdType:     ThresholdPercentage,
	}
	tripperOpts := TripperOptions{}
	tripper := Configure(tripperOpts)
	m, err := tripper.AddMonitor(monitorOptions)
	assert.NoError(t, err)
	assert.NotNil(t, m)

	result := m.IsCircuitOpen()
	assert.False(t, result)

	// Test case 2: Circuit is open
	// Expected output: true
	m.Get().CircuitOpen = true
	result = m.IsCircuitOpen()
	assert.True(t, result)
}
func TestGetAllMonitors(t *testing.T) {
	tripperOpts := TripperOptions{}
	tripper := Configure(tripperOpts)

	// Test case 1: Get all monitors when there are no monitors
	// Expected output: Empty map
	monitors := tripper.GetAllMonitors()
	assert.Empty(t, monitors)

	// Test case 2: Get all monitors when there are multiple monitors
	// Expected output: Map with all monitors
	monitorOptions1 := MonitorOptions{
		Name:              "test1",
		Threshold:         65,
		MinimumCount:      20,
		IntervalInSeconds: 120,
		ThresholdType:     ThresholdPercentage,
	}
	monitorOptions2 := MonitorOptions{
		Name:              "test2",
		Threshold:         70,
		MinimumCount:      30,
		IntervalInSeconds: 180,
		ThresholdType:     ThresholdPercentage,
	}
	_, err1 := tripper.AddMonitor(monitorOptions1)
	assert.NoError(t, err1)
	_, err2 := tripper.AddMonitor(monitorOptions2)
	assert.NoError(t, err2)

	monitors = tripper.GetAllMonitors()
	assert.Len(t, monitors, 2)
	assert.Contains(t, monitors, "test1")
	assert.Contains(t, monitors, "test2")
}
