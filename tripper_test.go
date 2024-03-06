package tripper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateStatusByMonitor(t *testing.T) {
	monitor := &Monitor{
		SuccessCount:      5,
		FailureCount:      3,
		CircuitOpen:       false,
		MinimumCount:      10,
		ThresholdType:     ThresholdCount,
		Threshold:         3,
		Name:              "test",
		IntervalInSeconds: 60,
	}

	monitor.CircuitOpenedSince = getStartOfIntervalTimestamp(monitor) - 30
	monitor.LastCapturedAt = getStartOfIntervalTimestamp(monitor)

	// Test case 1: Success
	updateStatusByMonitor(monitor, true)
	assert.Equal(t, int64(6), monitor.SuccessCount)
	assert.Equal(t, int64(3), monitor.FailureCount)
	assert.False(t, monitor.CircuitOpen)

	// Test case 2: Failure
	updateStatusByMonitor(monitor, false)
	assert.Equal(t, int64(6), monitor.SuccessCount)
	assert.Equal(t, int64(4), monitor.FailureCount)
	assert.True(t, monitor.CircuitOpen)

	// Test case 3: Circuit opened since more than a minute, reset counts
	monitor.LastCapturedAt = getStartOfIntervalTimestamp(monitor) - 120
	updateStatusByMonitor(monitor, true)
	assert.Equal(t, int64(1), monitor.SuccessCount)
	assert.Equal(t, int64(0), monitor.FailureCount)
	assert.False(t, monitor.CircuitOpen)
	assert.Equal(t, int64(0), monitor.CircuitOpenedSince)

	// Test case 4: Circuit opened since less than a minute, do not reset counts
	monitor.CircuitOpenedSince = monitor.LastCapturedAt - 30
	updateStatusByMonitor(monitor, false)
	assert.Equal(t, int64(1), monitor.SuccessCount)
	assert.Equal(t, int64(1), monitor.FailureCount)
	assert.False(t, monitor.CircuitOpen)
	assert.NotEqual(t, int64(0), monitor.CircuitOpenedSince)

	// Test case 5: Circuit open threshold count reached
	monitor.SuccessCount = 5
	monitor.FailureCount = 5
	monitor.ThresholdType = ThresholdCount
	monitor.Threshold = 5
	updateStatusByMonitor(monitor, false)
	assert.Equal(t, int64(5), monitor.SuccessCount)
	assert.Equal(t, int64(6), monitor.FailureCount)
	assert.True(t, monitor.CircuitOpen)
	assert.Equal(t, monitor.LastCapturedAt, monitor.CircuitOpenedSince)

	// Test case 6: Circuit open threshold percentage reached
	monitor.SuccessCount = 10
	monitor.FailureCount = 90
	monitor.ThresholdType = ThresholdPercentage
	monitor.Threshold = 80
	updateStatusByMonitor(monitor, false)
	assert.Equal(t, int64(10), monitor.SuccessCount)
	assert.Equal(t, int64(91), monitor.FailureCount)
	assert.True(t, monitor.CircuitOpen)
	assert.Equal(t, monitor.LastCapturedAt, monitor.CircuitOpenedSince)
}
func TestAddMonitor(t *testing.T) {
	// Test case 1: Add a new monitor successfully
	monitor := Monitor{
		SuccessCount:  5,
		FailureCount:  3,
		CircuitOpen:   false,
		MinimumCount:  10,
		ThresholdType: ThresholdCount,
		Threshold:     3,
		Name:          "test",
	}
	monitor.IntervalInSeconds = 60
	monitor.CircuitOpenedSince = getStartOfIntervalTimestamp(&monitor) - 30
	monitor.LastCapturedAt = getStartOfIntervalTimestamp(&monitor)

	err := AddMonitor(monitor)
	assert.NoError(t, err)
	assert.NotNil(t, circuits["test"])
	assert.Equal(t, &monitor, circuits["test"])

	// Test case 2: Add a monitor with an existing name
	err = AddMonitor(monitor)
	assert.Error(t, err)
	assert.EqualError(t, err, "Monitor with name test already exists")
	assert.NotNil(t, circuits["test"])
	assert.Equal(t, &monitor, circuits["test"])

	// Test case 3: Add a monitor with an invalid threshold type
	monitor.Name = "test1"
	monitor.ThresholdType = "invalid"
	err = AddMonitor(monitor)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid threshold type invalid")
	assert.Nil(t, circuits["test1"])

	// Test case 4: Add a monitor with an invalid threshold value for percentage type
	monitor.ThresholdType = ThresholdPercentage
	monitor.Threshold = -10
	err = AddMonitor(monitor)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid threshold value -10.000000 for percentage type")
	assert.Nil(t, circuits["test1"])

	// Test case 5: Add a monitor with an invalid threshold value for count type
	monitor.ThresholdType = ThresholdCount
	monitor.Threshold = 0
	err = AddMonitor(monitor)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid threshold value 0.000000 for count type")
	assert.Nil(t, circuits["test1"])

	// Test case 6: Add a monitor with an invalid minimum count
	monitor.MinimumCount = 0
	monitor.Threshold = 1
	err = AddMonitor(monitor)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid minimum count 0")
	assert.Nil(t, circuits["test1"])

	// Test case 7: Add a monitor with an invalid interval
	monitor.MinimumCount = 10
	monitor.IntervalInSeconds = 59
	err = AddMonitor(monitor)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid interval 59")
	assert.Nil(t, circuits["test1"])

	// Test case 8: Add a monitor with an interval that is not a multiple of 60
	monitor.IntervalInSeconds = 61
	err = AddMonitor(monitor)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid interval 61")
	assert.Nil(t, circuits["test1"])
}
func TestUpdateMonitorStatusByName(t *testing.T) {
	// Test case 1: Update monitor status successfully
	monitor := &Monitor{
		SuccessCount:  5,
		FailureCount:  3,
		CircuitOpen:   false,
		MinimumCount:  10,
		ThresholdType: ThresholdCount,
		Threshold:     3,
		Name:          "test",
	}

	monitor.IntervalInSeconds = 60
	monitor.CircuitOpenedSince = getStartOfIntervalTimestamp(monitor) - 30
	monitor.LastCapturedAt = getStartOfIntervalTimestamp(monitor)
	circuits["test"] = monitor
	updatedMonitor, err := UpdateMonitorStatusByName("test", true)
	assert.NoError(t, err)
	assert.Equal(t, int64(6), updatedMonitor.SuccessCount)
	assert.Equal(t, int64(3), updatedMonitor.FailureCount)
	assert.False(t, updatedMonitor.CircuitOpen)

	// Test case 2: Update monitor status with non-existent name
	_, err = UpdateMonitorStatusByName("nonexistent", true)
	assert.Error(t, err)
	assert.EqualError(t, err, "Monitor with name nonexistent does not exist")
}
func TestIsCircuitOpen(t *testing.T) {
	// Test case 1: Circuit is not open
	monitor := &Monitor{
		CircuitOpen: false,
		Name:        "test",
	}

	monitor.IntervalInSeconds = 60
	monitor.CircuitOpenedSince = getStartOfIntervalTimestamp(monitor) - 30
	monitor.LastCapturedAt = getStartOfIntervalTimestamp(monitor)
	circuits["test"] = monitor
	isOpen, err := IsCircuitOpen("test")
	assert.NoError(t, err)
	assert.False(t, isOpen)

	// Test case 2: Circuit is open, but less than a minute has passed
	monitor.CircuitOpen = true
	isOpen, err = IsCircuitOpen("test")
	assert.NoError(t, err)
	assert.True(t, isOpen)

	// Test case 3: Circuit is open, and more than a minute has passed
	monitor.CircuitOpenedSince = getStartOfIntervalTimestamp(monitor) - 70
	isOpen, err = IsCircuitOpen("test")
	assert.NoError(t, err)
	assert.False(t, isOpen)

	// Test case 4: Circuit does not exist
	isOpen, err = IsCircuitOpen("nonexistent")
	assert.Error(t, err)
	assert.EqualError(t, err, "Monitor with name nonexistent does not exist")
	assert.False(t, isOpen)
}
func TestGetMonitorByName(t *testing.T) {
	// Test case 1: Get existing monitor by name
	monitor := &Monitor{
		SuccessCount:  5,
		FailureCount:  3,
		CircuitOpen:   false,
		MinimumCount:  10,
		ThresholdType: ThresholdCount,
		Threshold:     3,
		Name:          "test",
	}
	monitor.IntervalInSeconds = 60
	monitor.CircuitOpenedSince = getStartOfIntervalTimestamp(monitor) - 30
	monitor.LastCapturedAt = getStartOfIntervalTimestamp(monitor)
	circuits["test"] = monitor

	result, err := GetMonitorByName("test")
	assert.NoError(t, err)
	assert.Equal(t, monitor, result)

	// Test case 2: Get non-existent monitor by name
	result, err = GetMonitorByName("nonexistent")
	assert.Error(t, err)
	assert.EqualError(t, err, "Monitor with name nonexistent does not exist")
	assert.Nil(t, result)
}
func TestGetAllMonitors(t *testing.T) {
	// Test case 1: Get all monitors successfully
	monitor1 := &Monitor{
		SuccessCount:      5,
		FailureCount:      3,
		CircuitOpen:       false,
		MinimumCount:      10,
		ThresholdType:     ThresholdCount,
		Threshold:         3,
		Name:              "test1",
		IntervalInSeconds: 60,
	}
	monitor2 := &Monitor{
		SuccessCount:      10,
		FailureCount:      5,
		CircuitOpen:       true,
		MinimumCount:      20,
		ThresholdType:     ThresholdPercentage,
		Threshold:         80,
		Name:              "test2",
		IntervalInSeconds: 60,
	}

	//reset circuits to empty
	circuits = make(map[string]*Monitor)

	circuits["test1"] = monitor1
	circuits["test2"] = monitor2

	result := GetAllMonitors()
	assert.Equal(t, 2, len(result))
	assert.Equal(t, monitor1, result["test1"])
	assert.Equal(t, monitor2, result["test2"])
}
