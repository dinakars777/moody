package mood

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// EventType represents the type of hardware event
type EventType int

const (
	EventSlap         EventType = iota // Accelerometer impact
	EventUSBIn                         // USB device connected
	EventUSBOut                        // USB device disconnected
	EventChargerIn                     // Power adapter connected
	EventChargerOut                    // Power adapter disconnected
	EventBatteryLow                    // Battery below 20%
	EventBatteryCrit                   // Battery below 5%
	EventLidClose                      // Lid closing
	EventLidOpen                       // Lid opening
	EventHeadphonesIn                  // Audio jack connected
	EventHeadphonesOut                 // Audio jack disconnected
	EventWiFiLost                      // WiFi disconnected
	EventWiFiBack                      // WiFi reconnected
	EventDisplayIn                     // External display connected
	EventDisplayOut                    // External display disconnected
)

// EventName returns a string identifier for voice pack lookups
func EventName(e EventType) string {
	names := map[EventType]string{
		EventSlap:         "slap",
		EventUSBIn:        "usb_in",
		EventUSBOut:       "usb_out",
		EventChargerIn:    "charger_in",
		EventChargerOut:   "charger_out",
		EventBatteryLow:   "battery_low",
		EventBatteryCrit:  "battery_crit",
		EventLidClose:     "lid_close",
		EventLidOpen:      "lid_open",
		EventHeadphonesIn: "headphones_in",
		EventHeadphonesOut:"headphones_out",
		EventWiFiLost:     "wifi_lost",
		EventWiFiBack:     "wifi_back",
		EventDisplayIn:    "display_in",
		EventDisplayOut:   "display_out",
	}
	if n, ok := names[e]; ok {
		return n
	}
	return "unknown"
}

// EventLabel returns a human-readable label
func EventLabel(e EventType) string {
	labels := map[EventType]string{
		EventSlap:         "Slap detected",
		EventUSBIn:        "USB device connected",
		EventUSBOut:       "USB device disconnected",
		EventChargerIn:    "Charger connected",
		EventChargerOut:   "Charger disconnected",
		EventBatteryLow:   "Battery low",
		EventBatteryCrit:  "Battery critical",
		EventLidClose:     "Lid closed",
		EventLidOpen:      "Lid opened",
		EventHeadphonesIn: "Headphones connected",
		EventHeadphonesOut:"Headphones disconnected",
		EventWiFiLost:     "WiFi disconnected",
		EventWiFiBack:     "WiFi reconnected",
		EventDisplayIn:    "Display connected",
		EventDisplayOut:   "Display disconnected",
	}
	if l, ok := labels[e]; ok {
		return l
	}
	return "Unknown event"
}

// HardwareEvent represents a detected hardware change
type HardwareEvent struct {
	Type      EventType
	Intensity float64   // 0.0-1.0, for accelerometer force
	Timestamp time.Time
	Meta      string    // e.g., USB device name, WiFi SSID
}

// moodImpact defines how each event shifts the mood
type moodImpact struct {
	Happiness float64
	Energy    float64
	Trust     float64
}

var impacts = map[EventType]moodImpact{
	EventSlap:         {-0.15, +0.05, -0.10},
	EventUSBIn:        {+0.05, +0.05, +0.02},
	EventUSBOut:       {-0.05, -0.02, -0.05},
	EventChargerIn:    {+0.15, +0.10, +0.05},
	EventChargerOut:   {-0.10, -0.05, -0.05},
	EventBatteryLow:   {-0.10, -0.15, 0},
	EventBatteryCrit:  {-0.20, -0.25, -0.10},
	EventLidClose:     {0, -0.05, 0},
	EventLidOpen:      {+0.05, +0.05, 0},
	EventHeadphonesIn: {+0.05, 0, +0.05},
	EventHeadphonesOut:{-0.03, +0.02, -0.02},
	EventWiFiLost:     {-0.15, -0.05, -0.10},
	EventWiFiBack:     {+0.10, +0.05, +0.05},
	EventDisplayIn:    {+0.05, +0.05, +0.02},
	EventDisplayOut:   {-0.03, -0.02, 0},
}

// Engine manages mood state and processes events
type Engine struct {
	mu          sync.RWMutex
	current     Mood
	eventCount  int
	lastEvent   *HardwareEvent
	stateFile   string
	decayTicker *time.Ticker
}

// NewEngine creates a mood engine and loads persisted state
func NewEngine() *Engine {
	homeDir, _ := os.UserHomeDir()
	stateDir := filepath.Join(homeDir, ".moody")
	os.MkdirAll(stateDir, 0755)

	e := &Engine{
		current:   Mood{Happiness: 0.5, Energy: 0.5, Trust: 0.5}, // Start slightly positive
		stateFile: filepath.Join(stateDir, "state.json"),
	}

	// Try to load persisted state
	e.loadState()

	// Start mood decay (drift toward neutral over time)
	e.decayTicker = time.NewTicker(30 * time.Second)
	go e.decayLoop()

	// Persist state periodically
	go e.persistLoop()

	return e
}

// ProcessEvent updates the mood based on a hardware event
func (e *Engine) ProcessEvent(evt HardwareEvent) MoodLabel {
	e.mu.Lock()
	defer e.mu.Unlock()

	impact, ok := impacts[evt.Type]
	if !ok {
		return e.current.Label()
	}

	// Apply impact, scaled by intensity for accelerometer events
	scale := 1.0
	if evt.Type == EventSlap && evt.Intensity > 0 {
		scale = evt.Intensity
	}

	e.current.Happiness = clamp(e.current.Happiness + impact.Happiness*scale)
	e.current.Energy = clamp(e.current.Energy + impact.Energy*scale)
	e.current.Trust = clamp(e.current.Trust + impact.Trust*scale)

	e.eventCount++
	e.lastEvent = &evt

	return e.current.Label()
}

// CurrentMood returns the current mood state
func (e *Engine) CurrentMood() Mood {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.current
}

// EventCount returns total events processed
func (e *Engine) EventCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.eventCount
}

// LastEvent returns the most recent event
func (e *Engine) LastEvent() *HardwareEvent {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.lastEvent
}

// decayLoop slowly drifts mood back toward neutral
func (e *Engine) decayLoop() {
	for range e.decayTicker.C {
		e.mu.Lock()
		decay := 0.005 // Per 30 seconds
		if e.current.Happiness > 0 {
			e.current.Happiness -= decay
		} else if e.current.Happiness < 0 {
			e.current.Happiness += decay
		}
		if e.current.Energy > 0 {
			e.current.Energy -= decay
		} else if e.current.Energy < 0 {
			e.current.Energy += decay
		}
		if e.current.Trust > 0 {
			e.current.Trust -= decay
		} else if e.current.Trust < 0 {
			e.current.Trust += decay
		}
		e.mu.Unlock()
	}
}

// persistLoop saves mood state to disk every 30 seconds
func (e *Engine) persistLoop() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		e.saveState()
	}
}

type persistedState struct {
	Mood       Mood      `json:"mood"`
	EventCount int       `json:"event_count"`
	SavedAt    time.Time `json:"saved_at"`
}

func (e *Engine) saveState() {
	e.mu.RLock()
	state := persistedState{
		Mood:       e.current,
		EventCount: e.eventCount,
		SavedAt:    time.Now(),
	}
	e.mu.RUnlock()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(e.stateFile, data, 0644)
}

func (e *Engine) loadState() {
	data, err := os.ReadFile(e.stateFile)
	if err != nil {
		return
	}

	var state persistedState
	if err := json.Unmarshal(data, &state); err != nil {
		return
	}

	// Only restore if saved less than 24 hours ago
	if time.Since(state.SavedAt) < 24*time.Hour {
		e.current = state.Mood
		e.eventCount = state.EventCount
	}
}

// Shutdown saves state and cleans up
func (e *Engine) Shutdown() {
	e.decayTicker.Stop()
	e.saveState()
}
