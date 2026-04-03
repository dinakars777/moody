package sensors

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation -framework Foundation

#include <IOKit/hid/IOHIDManager.h>
#include <CoreFoundation/CoreFoundation.h>
#include <math.h>
#include <stdlib.h>

// Accelerometer state
static IOHIDManagerRef hidManager = NULL;
static double lastX = 0, lastY = 0, lastZ = 0;
static int dataReady = 0;

static void inputCallback(void *context, IOReturn result, void *sender, IOHIDValueRef value) {
    IOHIDElementRef element = IOHIDValueGetElement(value);
    uint32_t usage = IOHIDElementGetUsage(element);
    CFIndex raw = IOHIDValueGetIntegerValue(value);
    double scaled = (double)raw / 10000.0; // Scale to approximate g-force

    switch (usage) {
        case 0x30: lastX = scaled; break; // X axis
        case 0x31: lastY = scaled; break; // Y axis
        case 0x32: lastZ = scaled; dataReady = 1; break; // Z axis (last to update)
    }
}

static int accel_init(void) {
    hidManager = IOHIDManagerCreate(kCFAllocatorDefault, kIOHIDOptionsTypeNone);
    if (!hidManager) return -1;

    // Match Apple SPU accelerometer
    CFMutableDictionaryRef match = CFDictionaryCreateMutable(
        kCFAllocatorDefault, 0,
        &kCFTypeDictionaryKeyCallBacks,
        &kCFTypeDictionaryValueCallBacks
    );

    int usagePage = 0x20; // kHIDPage_Sensor
    int usage = 0x73;     // kHIDUsage_Snsr_Motion_Accelerometer3D
    CFNumberRef pageRef = CFNumberCreate(kCFAllocatorDefault, kCFNumberIntType, &usagePage);
    CFNumberRef usageRef = CFNumberCreate(kCFAllocatorDefault, kCFNumberIntType, &usage);

    CFDictionarySetValue(match, CFSTR(kIOHIDDeviceUsagePageKey), pageRef);
    CFDictionarySetValue(match, CFSTR(kIOHIDDeviceUsageKey), usageRef);

    CFRelease(pageRef);
    CFRelease(usageRef);

    IOHIDManagerSetDeviceMatching(hidManager, match);
    CFRelease(match);

    IOHIDManagerRegisterInputValueCallback(hidManager, inputCallback, NULL);
    IOHIDManagerScheduleWithRunLoop(hidManager, CFRunLoopGetCurrent(), kCFRunLoopDefaultMode);

    IOReturn ret = IOHIDManagerOpen(hidManager, kIOHIDOptionsTypeNone);
    if (ret != kIOReturnSuccess) {
        CFRelease(hidManager);
        hidManager = NULL;
        return -2;
    }

    return 0;
}

static void accel_close(void) {
    if (hidManager) {
        IOHIDManagerClose(hidManager, kIOHIDOptionsTypeNone);
        CFRelease(hidManager);
        hidManager = NULL;
    }
}

// Read one sample (non-blocking). Returns 1 if data available.
static int accel_read(double *x, double *y, double *z) {
    // Run the runloop briefly to process HID events
    CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0.004, false);
    if (dataReady) {
        *x = lastX;
        *y = lastY;
        *z = lastZ;
        dataReady = 0;
        return 1;
    }
    return 0;
}
*/
import "C"
import (
	"math"
	"sync"
	"time"

	"github.com/dinakars777/moody/mood"
)

// Accelerometer detects physical impacts on the MacBook
type Accelerometer struct {
	mu           sync.Mutex
	running      bool
	stopCh       chan struct{}
	minAmplitude float64
	cooldownMs   int
	fastMode     bool
}

// NewAccelerometer creates an accelerometer sensor
func NewAccelerometer(minAmplitude float64, cooldownMs int, fast bool) *Accelerometer {
	if minAmplitude <= 0 {
		minAmplitude = 0.05
	}
	if cooldownMs <= 0 {
		cooldownMs = 750
	}
	if fast {
		if minAmplitude == 0.05 {
			minAmplitude = 0.18
		}
		if cooldownMs == 750 {
			cooldownMs = 350
		}
	}
	return &Accelerometer{
		stopCh:       make(chan struct{}),
		minAmplitude: minAmplitude,
		cooldownMs:   cooldownMs,
		fastMode:     fast,
	}
}

func (a *Accelerometer) Name() string { return "Accelerometer" }

func (a *Accelerometer) Available() bool {
	// Try to init and immediately close to test availability
	ret := C.accel_init()
	if ret == 0 {
		C.accel_close()
		return true
	}
	return false
}

func (a *Accelerometer) Start(events chan<- mood.HardwareEvent) error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return nil
	}
	a.running = true
	a.mu.Unlock()

	ret := C.accel_init()
	if ret != 0 {
		return errSensorInit("accelerometer", int(ret))
	}

	go a.pollLoop(events)
	return nil
}

func (a *Accelerometer) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.running {
		close(a.stopCh)
		a.running = false
		C.accel_close()
	}
}

func (a *Accelerometer) pollLoop(events chan<- mood.HardwareEvent) {
	// Ring buffer for STA/LTA detection
	const bufSize = 200
	var buf [bufSize]float64
	var bufIdx int
	var bufFull bool
	var lastTrigger time.Time

	pollInterval := 10 * time.Millisecond
	if a.fastMode {
		pollInterval = 4 * time.Millisecond
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopCh:
			return
		case <-ticker.C:
			var x, y, z C.double
			if C.accel_read(&x, &y, &z) != 1 {
				continue
			}

			// Compute acceleration magnitude (removing gravity ~1g on z)
			magnitude := math.Sqrt(float64(x*x) + float64(y*y) + float64((z-1.0)*(z-1.0)))

			// Store in ring buffer
			buf[bufIdx] = magnitude
			bufIdx = (bufIdx + 1) % bufSize
			if bufIdx == 0 {
				bufFull = true
			}

			if !bufFull {
				continue
			}

			// Calculate mean and check for spike
			var sum float64
			for _, v := range buf {
				sum += v
			}
			mean := sum / float64(bufSize)

			// Peak detection: is current value significantly above the mean?
			peak := magnitude - mean
			if peak > a.minAmplitude {
				// Cooldown check
				if time.Since(lastTrigger) < time.Duration(a.cooldownMs)*time.Millisecond {
					continue
				}
				lastTrigger = time.Now()

				// Normalize intensity to 0.0-1.0 range
				intensity := math.Min(peak/0.5, 1.0)

				events <- mood.HardwareEvent{
					Type:      mood.EventSlap,
					Intensity: intensity,
					Timestamp: time.Now(),
					Meta:      "physical impact",
				}
			}
		}
	}
}

type sensorError struct {
	sensor string
	code   int
}

func (e *sensorError) Error() string {
	return e.sensor + ": init failed with code " + string(rune('0'+e.code))
}

func errSensorInit(name string, code int) error {
	return &sensorError{sensor: name, code: code}
}
