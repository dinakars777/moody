package sensors

import "github.com/dinakars777/moody/mood"

// Status describes whether a sensor is expected to work on this machine.
type Status struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Supported    bool              `json:"supported"`
	Available    bool              `json:"available"`
	Optional     bool              `json:"optional,omitempty"`
	Reason       string            `json:"reason,omitempty"`
	SuggestedFix string            `json:"suggestedFix,omitempty"`
	Details      map[string]string `json:"details,omitempty"`
}

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
	// Status returns structured diagnostics for this sensor
	Status() Status
}
