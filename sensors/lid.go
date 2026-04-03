package sensors

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation

#include <IOKit/IOKitLib.h>
#include <CoreFoundation/CoreFoundation.h>

// Check clamshell (lid) state: 1 = closed, 0 = open, -1 = error
static int get_clamshell_state(void) {
    io_registry_entry_t rootDomain;
    CFBooleanRef clamshellState;
    int result = -1;
    
    rootDomain = IORegistryEntryFromPath(kIOMainPortDefault,
        "IOService:/IOResources/IODisplayWrangler");
    if (!rootDomain) {
        // Try alternative path for newer macOS
        rootDomain = IOServiceGetMatchingService(kIOMainPortDefault,
            IOServiceMatching("IOPMrootDomain"));
    }
    if (!rootDomain) return -1;
    
    clamshellState = IORegistryEntryCreateCFProperty(rootDomain,
        CFSTR("AppleClamshellState"), kCFAllocatorDefault, 0);
    
    if (clamshellState) {
        result = CFBooleanGetValue(clamshellState) ? 1 : 0;
        CFRelease(clamshellState);
    }
    
    IOObjectRelease(rootDomain);
    return result;
}
*/
import "C"
import (
	"sync"
	"time"

	"github.com/dinakars777/moody/mood"
)

// Lid monitors MacBook lid open/close state
type Lid struct {
	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
}

func NewLid() *Lid {
	return &Lid{
		stopCh: make(chan struct{}),
	}
}

func (l *Lid) Name() string { return "Lid" }

func (l *Lid) Available() bool {
	state := int(C.get_clamshell_state())
	return state >= 0
}

func (l *Lid) Start(events chan<- mood.HardwareEvent) error {
	l.mu.Lock()
	if l.running {
		l.mu.Unlock()
		return nil
	}
	l.running = true
	l.mu.Unlock()

	go l.pollLoop(events)
	return nil
}

func (l *Lid) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.running {
		close(l.stopCh)
		l.running = false
	}
}

func (l *Lid) pollLoop(events chan<- mood.HardwareEvent) {
	prevState := int(C.get_clamshell_state())

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-l.stopCh:
			return
		case <-ticker.C:
			state := int(C.get_clamshell_state())
			if state < 0 || state == prevState {
				continue
			}

			if state == 1 {
				events <- mood.HardwareEvent{
					Type:      mood.EventLidClose,
					Timestamp: time.Now(),
					Meta:      "lid closed",
				}
			} else {
				events <- mood.HardwareEvent{
					Type:      mood.EventLidOpen,
					Timestamp: time.Now(),
					Meta:      "lid opened",
				}
			}
			prevState = state
		}
	}
}
