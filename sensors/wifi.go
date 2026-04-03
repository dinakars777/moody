package sensors

import (
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/dinakars777/moody/mood"
)

// WiFi monitors WiFi connection state by polling networksetup
type WiFi struct {
	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
	iface   string // network interface name (usually en0)
}

func NewWiFi() *WiFi {
	return &WiFi{
		stopCh: make(chan struct{}),
		iface:  detectWiFiInterface(),
	}
}

func (w *WiFi) Name() string { return "WiFi" }

func (w *WiFi) Available() bool {
	return w.iface != ""
}

func (w *WiFi) Start(events chan<- mood.HardwareEvent) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return nil
	}
	w.running = true
	w.mu.Unlock()

	go w.pollLoop(events)
	return nil
}

func (w *WiFi) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.running {
		close(w.stopCh)
		w.running = false
	}
}

func (w *WiFi) pollLoop(events chan<- mood.HardwareEvent) {
	prevSSID := getSSID(w.iface)
	prevConnected := prevSSID != ""

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			ssid := getSSID(w.iface)
			connected := ssid != ""

			if prevConnected && !connected {
				events <- mood.HardwareEvent{
					Type:      mood.EventWiFiLost,
					Timestamp: time.Now(),
					Meta:      prevSSID,
				}
			} else if !prevConnected && connected {
				events <- mood.HardwareEvent{
					Type:      mood.EventWiFiBack,
					Timestamp: time.Now(),
					Meta:      ssid,
				}
			}

			prevSSID = ssid
			prevConnected = connected
		}
	}
}

// getSSID returns the current WiFi SSID, or "" if not connected
func getSSID(iface string) string {
	out, err := exec.Command("networksetup", "-getairportnetwork", iface).Output()
	if err != nil {
		return ""
	}
	line := strings.TrimSpace(string(out))

	// Output: "Current Wi-Fi Network: MyNetwork"
	// Or:     "You are not associated with an AirPort network."
	if strings.Contains(line, "not associated") || strings.Contains(line, "Error") {
		return ""
	}

	parts := strings.SplitN(line, ": ", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

// detectWiFiInterface finds the WiFi interface name (usually en0 on MacBooks)
func detectWiFiInterface() string {
	out, err := exec.Command("networksetup", "-listallhardwareports").Output()
	if err != nil {
		return "en0" // Fallback
	}

	lines := strings.Split(string(out), "\n")
	foundWiFi := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Wi-Fi") || strings.Contains(line, "AirPort") {
			foundWiFi = true
			continue
		}
		if foundWiFi && strings.HasPrefix(line, "Device:") {
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
		if foundWiFi && line == "" {
			foundWiFi = false
		}
	}

	return "en0" // Fallback
}
