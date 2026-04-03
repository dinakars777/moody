package sensors

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation -framework Foundation

#include <IOKit/IOKitLib.h>
#include <IOKit/usb/IOUSBLib.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>
#include <string.h>

// Simple USB device count tracking
static int count_usb_devices(void) {
    int count = 0;
    CFMutableDictionaryRef match = IOServiceMatching(kIOUSBDeviceClassName);
    if (!match) return 0;
    
    io_iterator_t iter;
    kern_return_t kr = IOServiceGetMatchingServices(kIOMainPortDefault, match, &iter);
    if (kr != KERN_SUCCESS) return 0;
    
    io_service_t device;
    while ((device = IOIteratorNext(iter)) != 0) {
        count++;
        IOObjectRelease(device);
    }
    IOObjectRelease(iter);
    return count;
}

// Get the name of the most recently added USB device
static void get_last_usb_name(char *buf, int bufLen) {
    buf[0] = '\0';
    CFMutableDictionaryRef match = IOServiceMatching(kIOUSBDeviceClassName);
    if (!match) return;
    
    io_iterator_t iter;
    kern_return_t kr = IOServiceGetMatchingServices(kIOMainPortDefault, match, &iter);
    if (kr != KERN_SUCCESS) return;
    
    io_service_t device;
    io_service_t lastDevice = 0;
    while ((device = IOIteratorNext(iter)) != 0) {
        if (lastDevice) IOObjectRelease(lastDevice);
        lastDevice = device;
    }
    
    if (lastDevice) {
        io_name_t name;
        if (IORegistryEntryGetName(lastDevice, name) == KERN_SUCCESS) {
            strncpy(buf, name, bufLen - 1);
            buf[bufLen - 1] = '\0';
        }
        IOObjectRelease(lastDevice);
    }
    IOObjectRelease(iter);
}
*/
import "C"
import (
	"sync"
	"time"
	"unsafe"

	"github.com/dinakars777/moody/mood"
)

// USB monitors USB device connections and disconnections
type USB struct {
	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
}

func NewUSB() *USB {
	return &USB{
		stopCh: make(chan struct{}),
	}
}

func (u *USB) Name() string    { return "USB" }
func (u *USB) Available() bool { return true }

func (u *USB) Start(events chan<- mood.HardwareEvent) error {
	u.mu.Lock()
	if u.running {
		u.mu.Unlock()
		return nil
	}
	u.running = true
	u.mu.Unlock()

	go u.pollLoop(events)
	return nil
}

func (u *USB) Stop() {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.running {
		close(u.stopCh)
		u.running = false
	}
}

func (u *USB) pollLoop(events chan<- mood.HardwareEvent) {
	prevCount := int(C.count_usb_devices())

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-u.stopCh:
			return
		case <-ticker.C:
			count := int(C.count_usb_devices())
			if count == prevCount {
				continue
			}

			if count > prevCount {
				// Device connected
				var nameBuf [256]C.char
				C.get_last_usb_name(&nameBuf[0], 256)
				name := C.GoString((*C.char)(unsafe.Pointer(&nameBuf[0])))
				if name == "" {
					name = "USB device"
				}

				events <- mood.HardwareEvent{
					Type:      mood.EventUSBIn,
					Timestamp: time.Now(),
					Meta:      name,
				}
			} else {
				// Device disconnected
				events <- mood.HardwareEvent{
					Type:      mood.EventUSBOut,
					Timestamp: time.Now(),
					Meta:      "USB device",
				}
			}
			prevCount = count
		}
	}
}
