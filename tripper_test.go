package tripper

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddMonitor(t *testing.T) {
	// Test case 1: Add a new monitor successfully
	// Expected output: No error
	monitorOptions := CircuitOptions{
		Name:              "test",
		Threshold:         65,
		MinimumCount:      20,
		IntervalInSeconds: 120,
		ThresholdType:     ThresholdPercentage,
	}
	_, err := ConfigureCircuit(monitorOptions)
	assert.NoError(t, err)

	// Test case 3: Add a monitor with an invalid threshold type
	// Expected output: An error
	monitorOptions.ThresholdType = "invalid"
	monitorOptions.Name = "invalid"
	_, err = ConfigureCircuit(monitorOptions)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid threshold type invalid")
	// Test case 4: Add a monitor with an invalid threshold value for percentage type
	// Expected output: An error
	monitorOptions.ThresholdType = ThresholdPercentage
	monitorOptions.Threshold = -1
	_, err = ConfigureCircuit(monitorOptions)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid threshold value -1.000000 for percentage type")
	// Test case 5: Add a monitor with an invalid threshold value for count type
	// Expected output: An error
	monitorOptions.ThresholdType = ThresholdCount
	monitorOptions.Threshold = 0
	_, err = ConfigureCircuit(monitorOptions)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid threshold value 0.000000 for count type")
	// Test case 6: Add a monitor with an invalid minimum count
	// Expected output: An error
	monitorOptions.ThresholdType = ThresholdPercentage
	monitorOptions.Threshold = 65
	monitorOptions.MinimumCount = 0
	_, err = ConfigureCircuit(monitorOptions)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid minimum count 0")
	// Test case 7: Add a monitor with an invalid interval
	// Expected output: An error
	monitorOptions.MinimumCount = 20
	monitorOptions.IntervalInSeconds = 2
	_, err = ConfigureCircuit(monitorOptions)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid interval 2")
	// Test case 8: Add a monitor with an interval that is not a multiple of 60
	// Expected output: An error

	// Test case 9: Add a monitor with a valid threshold type and threshold value
	// Expected output: No error
	monitorOptions.IntervalInSeconds = 120
	monitorOptions.Name = "test2"
	monitorOptions.ThresholdType = ThresholdCount
	monitorOptions.Threshold = 10
	_, err = ConfigureCircuit(monitorOptions)
	assert.NoError(t, err)

	// Test case 10: Add a monitor with a  threshold type count and minimum value less than threshold
	// Expected output: An error
	monitorOptions.MinimumCount = 5
	monitorOptions.Threshold = 6
	monitorOptions.Name = "test10"
	_, err = ConfigureCircuit(monitorOptions)
	assert.Error(t, err)
	assert.EqualError(t, err, "minimum count should be greater than threshold")
}

func TestUpdateStatus(t *testing.T) {
	// Test case 1: Update status with success=true
	// Expected output: Success count incremented
	callBackCalledOpen := false
	callBackCalledClosed := false
	monitorOptions := CircuitOptions{
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
	m, err := ConfigureCircuit(monitorOptions)
	assert.NoError(t, err)
	assert.NotNil(t, m)

	m.UpdateStatus(true)
	assert.Equal(t, int64(1), m.Data().SuccessCount)
	assert.Equal(t, int64(0), m.Data().FailureCount)

	// Test case 2: Update status with success=false
	// Expected output: Failure count incremented
	m.UpdateStatus(false)
	assert.Equal(t, int64(1), m.Data().SuccessCount)
	assert.Equal(t, int64(1), m.Data().FailureCount)

	// Test case 3: Update status with success=true and minimum count not reached
	// Expected output: Success count incremented, circuit not opened

	m.UpdateStatus(true)
	assert.Equal(t, int64(2), m.Data().SuccessCount)
	assert.Equal(t, int64(1), m.Data().FailureCount)
	assert.False(t, m.Data().IsCircuitOpen)

	// Test case 4: Update status with success=false and minimum count not reached
	// Expected output: Failure count incremented, circuit not opened
	m.UpdateStatus(false)
	assert.Equal(t, int64(2), m.Data().SuccessCount)
	assert.Equal(t, int64(2), m.Data().FailureCount)
	assert.True(t, m.Data().IsCircuitOpen)
	assert.True(t, callBackCalledOpen)
	assert.NotZero(t, m.Data().CircuitOpenedSince)

	// Test case 5: Update status with success=false and failure count greater than threshold (ThresholdType=ThresholdCount)
	// Expected output: Failure count incremented, circuit opened
	m.UpdateStatus(false)
	assert.Equal(t, int64(2), m.Data().SuccessCount)
	assert.Equal(t, int64(3), m.Data().FailureCount)
	assert.True(t, m.Data().IsCircuitOpen)
	assert.NotZero(t, m.Data().CircuitOpenedSince)

	m.UpdateStatus(true)
	m.UpdateStatus(true)
	m.UpdateStatus(true)
	m.UpdateStatus(true)
	m.UpdateStatus(true)
	assert.False(t, m.Data().IsCircuitOpen)
	assert.True(t, callBackCalledClosed)

	// Test case 6: Update status with success=false and failure percentage greater than threshold (ThresholdType=ThresholdPercentage)
	// Expected output: Failure count incremented, circuit opened
	monitorOptions.ThresholdType = ThresholdCount
	monitorOptions.Threshold = 2
	monitorOptions.MinimumCount = 3
	m, err = ConfigureCircuit(monitorOptions)
	assert.NoError(t, err)

	m.UpdateStatus(false)
	assert.Equal(t, int64(0), m.Data().SuccessCount)
	assert.Equal(t, int64(1), m.Data().FailureCount)
	assert.False(t, m.Data().IsCircuitOpen)

	m.UpdateStatus(false)
	m.UpdateStatus(false)
	m.UpdateStatus(true)
	assert.Equal(t, int64(1), m.Data().SuccessCount)
	assert.Equal(t, int64(3), m.Data().FailureCount)
	assert.True(t, m.Data().IsCircuitOpen)
	assert.NotZero(t, m.Data().CircuitOpenedSince)

	monitorOptions.Threshold = 20
	monitorOptions.MinimumCount = 21
	m, err = ConfigureCircuit(monitorOptions)
	assert.NoError(t, err)
	m.UpdateStatus(true)
	assert.False(t, m.Data().IsCircuitOpen)
	assert.True(t, callBackCalledClosed)
}

// func TestInterval(t *testing.T) {
// 	callBackCalledClosed := false
// 	monitorOptionsX := CircuitOptions{
// 		Name:              generateRandomString(10),
// 		Threshold:         2,
// 		MinimumCount:      4,
// 		IntervalInSeconds: 5,
// 		ThresholdType:     ThresholdCount,
// 		OnCircuitClosed: func(x CallbackEvent) {
// 			callBackCalledClosed = true
// 		},
// 	}
// 	mx, err := ConfigureCircuit(monitorOptionsX)
// 	assert.NoError(t, err)

// 	// Simulate failures and a success
// 	mx.UpdateStatus(false)
// 	mx.UpdateStatus(false)
// 	mx.UpdateStatus(true)
// 	mx.UpdateStatus(false)

// 	// Assert initial state
// 	assert.Equal(t, int64(1), mx.Data().SuccessCount)
// 	assert.Equal(t, int64(3), mx.Data().FailureCount)
// 	assert.True(t, mx.Data().IsCircuitOpen)
// 	assert.NotZero(t, mx.Data().CircuitOpenedSince)

// 	futureTime := time.Now().Unix() + int64(monitorOptionsX.IntervalInSeconds+1)
// 	for time.Now().Unix() < futureTime {
// 		// Simulate internal logic of the monitor on interval tick
// 		//mx.UpdateStatus(true) // Simulate any internal housekeeping
// 	}

//		// Assert circuit closed state
//		assert.Equal(t, int64(0), mx.Data().SuccessCount)
//		assert.Equal(t, int64(0), mx.Data().FailureCount)
//		assert.False(t, mx.Data().IsCircuitOpen)
//		assert.True(t, callBackCalledClosed)
//		assert.Zero(t, mx.Data().CircuitOpenedSince)
//	}
func generateRandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)

}
func TestUpdateStatusRaceSingle(t *testing.T) {
	monitorOptionsX := CircuitOptions{
		Name:              generateRandomString(10),
		Threshold:         2,
		MinimumCount:      4,
		IntervalInSeconds: 60,
		ThresholdType:     ThresholdCount,
	}
	mx, err := ConfigureCircuit(monitorOptionsX)
	assert.NoError(t, err)
	numGoroutines := 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			rand.Seed(time.Now().UnixNano())
			randId := rand.Intn(100-1) + 1
			mx.UpdateStatus(randId%2 == 0)
			// Decrement the wait group counter
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Equal(t, int64(100), mx.Data().SuccessCount+mx.Data().FailureCount)
}

func TestUpdateStatusRaceTwo(t *testing.T) {
	monitorOptionsX := CircuitOptions{
		Name:              "x",
		Threshold:         2,
		MinimumCount:      4,
		IntervalInSeconds: 60,
		ThresholdType:     ThresholdCount,
	}
	mx, err := ConfigureCircuit(monitorOptionsX)
	assert.NoError(t, err)
	monitorOptionsX1 := CircuitOptions{
		Name:              "x1",
		Threshold:         2,
		MinimumCount:      4,
		IntervalInSeconds: 60,
		ThresholdType:     ThresholdCount,
	}

	mx1, errx1 := ConfigureCircuit(monitorOptionsX1)
	assert.NoError(t, errx1)

	numGoroutines := 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {

			rand.Seed(time.Now().UnixNano())
			randId := rand.Intn(100-1) + 1
			mx.UpdateStatus(randId%2 == 0)
			mx1.UpdateStatus(randId%2 == 0)
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Equal(t, int64(200), mx.Data().SuccessCount+mx.Data().FailureCount+mx1.Data().SuccessCount+mx1.Data().FailureCount)

}

// func TestGetMonitor(t *testing.T) {
// 	tripperOpts := TripperOptions{}
// 	tripper := Configure(tripperOpts)

// 	// Test case 1: Get an existing monitor
// 	// Expected output: Monitor and no error
// 	monitorOptions := CircuitOptions{
// 		Name:              "test",
// 		Threshold:         65,
// 		MinimumCount:      20,
// 		IntervalInSeconds: 120,
// 		ThresholdType:     ThresholdPercentage,
// 	}
// 	_, err := tripper.AddMonitor(monitorOptions)
// 	assert.NoError(t, err)

// 	m, err := tripper.GetMonitor("test")
// 	assert.NoError(t, err)
// 	assert.NotNil(t, m)

// 	// Test case 2: Get a non-existing monitor
// 	// Expected output: Error
// 	_, err = tripper.GetMonitor("non-existing")
// 	assert.Error(t, err)
// 	assert.EqualError(t, err, "Monitor with name non-existing does not exist")
// }
// func TestIsCircuitOpen(t *testing.T) {
// 	// Test case 1: Circuit is closed
// 	// Expected output: false
// 	monitorOptions := CircuitOptions{
// 		Name:              "test",
// 		Threshold:         65,
// 		MinimumCount:      20,
// 		IntervalInSeconds: 120,
// 		ThresholdType:     ThresholdPercentage,
// 	}
// 	tripperOpts := TripperOptions{}
// 	tripper := Configure(tripperOpts)
// 	m, err := tripper.AddMonitor(monitorOptions)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, m)

// 	result := m.IsCircuitOpen()
// 	assert.False(t, result)

// }
func TestIsCircuitOpen(t *testing.T) {
	// Test case 1: Circuit is open
	// Expected output: true
	monitorOptions := CircuitOptions{
		Name:              "test",
		Threshold:         50,
		MinimumCount:      2,
		IntervalInSeconds: 120,
		ThresholdType:     ThresholdPercentage,
	}
	m, err := ConfigureCircuit(monitorOptions)
	assert.NoError(t, err)
	m.UpdateStatus(false)
	m.UpdateStatus(false)
	m.UpdateStatus(false)
	assert.True(t, m.IsCircuitOpen())

	// Test case 2: Circuit is closed
	// Expected output: false
	m.UpdateStatus(true)
	m.UpdateStatus(true)
	m.UpdateStatus(true)
	m.UpdateStatus(true)
	m.UpdateStatus(true)
	m.UpdateStatus(true)
	m.UpdateStatus(true)
	assert.False(t, m.IsCircuitOpen())
}
