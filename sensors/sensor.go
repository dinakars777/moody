package sensors

import "github.com/dinakars777/moody/mood"

// Sensor is implemented by all hardware event sources
type Sensor interface {
	// Name returns a human-readable sensor name
	Name() string
	// Start begins monitoring and sends events to the channel
	Start(events chan<- mood.HardwareEvent) error
	// Stop cleanly shuts down the sensor
	Stop()
	// Available returns true if this sensor can run on the current hardware
	Available() bool
}
