package sensors

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation -framework Foundation

#include <IOKit/ps/IOPowerSources.h>
#include <IOKit/ps/IOPSKeys.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>

typedef struct {
    int is_charging;
    int battery_pct;
    int ac_connected;
} PowerState;

static PowerState get_power_state(void) {
    PowerState state = {0, -1, 0};
    
    CFTypeRef info = IOPSCopyPowerSourcesInfo();
    if (!info) return state;
    
    CFArrayRef sources = IOPSCopyPowerSourcesList(info);
    if (!sources) {
        CFRelease(info);
        return state;
    }
    
    if (CFArrayGetCount(sources) > 0) {
        CFDictionaryRef ps = IOPSGetPowerSourceDescription(info, CFArrayGetValueAtIndex(sources, 0));
        if (ps) {
            // Battery percentage
            CFNumberRef cap = CFDictionaryGetValue(ps, CFSTR(kIOPSCurrentCapacityKey));
            if (cap) CFNumberGetValue(cap, kCFNumberIntType, &state.battery_pct);
            
            // Power source type
            CFStringRef source = CFDictionaryGetValue(ps, CFSTR(kIOPSPowerSourceStateKey));
            if (source) {
                state.ac_connected = CFStringCompare(source, CFSTR(kIOPSACPowerValue), 0) == kCFCompareEqualTo;
            }
            
            // Charging state
            CFBooleanRef charging = CFDictionaryGetValue(ps, CFSTR(kIOPSIsChargingKey));
            if (charging) {
                state.is_charging = CFBooleanGetValue(charging);
            }
        }
    }
    
    CFRelease(sources);
    CFRelease(info);
    return state;
}
*/
import "C"
import (
	"fmt"
	"sync"
	"time"

	"github.com/dinakars777/moody/mood"
)

// Power monitors charger and battery state
type Power struct {
	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
}

func NewPower() *Power {
	return &Power{
		stopCh: make(chan struct{}),
	}
}

func (p *Power) Name() string    { return "Power/Battery" }
func (p *Power) Available() bool { return true } // Always available on macOS

func (p *Power) Start(events chan<- mood.HardwareEvent) error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return nil
	}
	p.running = true
	p.mu.Unlock()

	go p.pollLoop(events)
	return nil
}

func (p *Power) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.running {
		close(p.stopCh)
		p.running = false
	}
}

func (p *Power) pollLoop(events chan<- mood.HardwareEvent) {
	var prevAC int = -1 // -1 = unknown
	var sentLow, sentCrit bool
	
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopCh:
			return
		case <-ticker.C:
			state := C.get_power_state()
			ac := int(state.ac_connected)
			pct := int(state.battery_pct)

			// Charger events
			if prevAC >= 0 && ac != prevAC {
				if ac == 1 {
					events <- mood.HardwareEvent{
						Type:      mood.EventChargerIn,
						Timestamp: time.Now(),
						Meta:      fmt.Sprintf("Battery at %d%%", pct),
					}
					sentLow = false
					sentCrit = false
				} else {
					events <- mood.HardwareEvent{
						Type:      mood.EventChargerOut,
						Timestamp: time.Now(),
						Meta:      fmt.Sprintf("Battery at %d%%", pct),
					}
				}
			}
			prevAC = ac

			// Battery level events (only on battery power)
			if ac == 0 && pct >= 0 {
				if pct <= 5 && !sentCrit {
					events <- mood.HardwareEvent{
						Type:      mood.EventBatteryCrit,
						Timestamp: time.Now(),
						Meta:      fmt.Sprintf("%d%%", pct),
					}
					sentCrit = true
				} else if pct <= 20 && !sentLow {
					events <- mood.HardwareEvent{
						Type:      mood.EventBatteryLow,
						Timestamp: time.Now(),
						Meta:      fmt.Sprintf("%d%%", pct),
					}
					sentLow = true
				}
			}

			// Reset flags when battery is charging and level rises
			if ac == 1 {
				if pct > 25 {
					sentLow = false
				}
				if pct > 10 {
					sentCrit = false
				}
			}
		}
	}
}
