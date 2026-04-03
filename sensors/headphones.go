package sensors

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreAudio -framework AudioToolbox -framework Foundation

#include <CoreAudio/CoreAudio.h>
#include <AudioToolbox/AudioToolbox.h>
#include <string.h>

// Returns the transport type of the default output device
// 0 = unknown, 1 = built-in speaker, 2 = headphones/USB/bluetooth
static int get_audio_output_type(void) {
    AudioDeviceID deviceID;
    UInt32 size = sizeof(deviceID);
    AudioObjectPropertyAddress addr = {
        kAudioHardwarePropertyDefaultOutputDevice,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    OSStatus status = AudioObjectGetPropertyData(
        kAudioObjectSystemObject, &addr, 0, NULL, &size, &deviceID);
    if (status != noErr) return 0;

    // Get transport type
    UInt32 transportType;
    size = sizeof(transportType);
    AudioObjectPropertyAddress transportAddr = {
        kAudioDevicePropertyTransportType,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    status = AudioObjectGetPropertyData(deviceID, &transportAddr, 0, NULL, &size, &transportType);
    if (status != noErr) return 0;

    // kAudioDeviceTransportTypeBuiltIn = 'bltn'
    // kAudioDeviceTransportTypeUSB = 'usb '
    // kAudioDeviceTransportTypeBluetooth = 'blue'
    // kAudioDeviceTransportTypeBluetoothLE = 'blea'
    // kAudioDeviceTransportTypeHDMI = 'hdmi'
    // kAudioDeviceTransportTypeAirPlay = 'airp'
    if (transportType == kAudioDeviceTransportTypeBuiltIn) {
        return 1; // Built-in speakers
    }
    return 2; // External audio (headphones, USB, bluetooth, etc.)
}

// Get the name of the default output device
static void get_audio_output_name(char *buf, int bufLen) {
    buf[0] = '\0';
    AudioDeviceID deviceID;
    UInt32 size = sizeof(deviceID);
    AudioObjectPropertyAddress addr = {
        kAudioHardwarePropertyDefaultOutputDevice,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    OSStatus status = AudioObjectGetPropertyData(
        kAudioObjectSystemObject, &addr, 0, NULL, &size, &deviceID);
    if (status != noErr) return;

    CFStringRef name = NULL;
    size = sizeof(name);
    AudioObjectPropertyAddress nameAddr = {
        kAudioObjectPropertyName,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    status = AudioObjectGetPropertyData(deviceID, &nameAddr, 0, NULL, &size, &name);
    if (status != noErr || !name) return;

    CFStringGetCString(name, buf, bufLen, kCFStringEncodingUTF8);
    CFRelease(name);
}

// Check data source for headphone detection on built-in audio
// Returns 1 if headphones detected, 0 if speakers, -1 if unknown
static int check_headphone_jack(void) {
    AudioDeviceID deviceID;
    UInt32 size = sizeof(deviceID);
    AudioObjectPropertyAddress addr = {
        kAudioHardwarePropertyDefaultOutputDevice,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    OSStatus status = AudioObjectGetPropertyData(
        kAudioObjectSystemObject, &addr, 0, NULL, &size, &deviceID);
    if (status != noErr) return -1;

    // Get the data source
    UInt32 dataSource;
    size = sizeof(dataSource);
    AudioObjectPropertyAddress sourceAddr = {
        kAudioDevicePropertyDataSource,
        kAudioDevicePropertyScopeOutput,
        kAudioObjectPropertyElementMain
    };

    status = AudioObjectGetPropertyData(deviceID, &sourceAddr, 0, NULL, &size, &dataSource);
    if (status != noErr) return -1;

    // 'hdpn' = headphones, 'ispk' = internal speakers
    if (dataSource == 'hdpn') return 1;
    if (dataSource == 'ispk') return 0;
    return -1;
}
*/
import "C"
import (
	"sync"
	"time"
	"unsafe"

	"github.com/dinakars777/moody/mood"
)

// Headphones monitors audio output route changes
type Headphones struct {
	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
}

func NewHeadphones() *Headphones {
	return &Headphones{
		stopCh: make(chan struct{}),
	}
}

func (h *Headphones) Name() string    { return "Headphones" }
func (h *Headphones) Available() bool { return true }

func (h *Headphones) Start(events chan<- mood.HardwareEvent) error {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return nil
	}
	h.running = true
	h.mu.Unlock()

	go h.pollLoop(events)
	return nil
}

func (h *Headphones) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.running {
		close(h.stopCh)
		h.running = false
	}
}

func (h *Headphones) pollLoop(events chan<- mood.HardwareEvent) {
	prevType := int(C.get_audio_output_type())
	prevJack := int(C.check_headphone_jack())

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-h.stopCh:
			return
		case <-ticker.C:
			outputType := int(C.get_audio_output_type())
			jackState := int(C.check_headphone_jack())

			// Get device name for meta info
			var nameBuf [256]C.char
			C.get_audio_output_name(&nameBuf[0], 256)
			name := C.GoString((*C.char)(unsafe.Pointer(&nameBuf[0])))
			if name == "" {
				name = "audio device"
			}

			changed := false
			headphonesIn := false

			// Check transport type change (e.g., Bluetooth headphones)
			if outputType != prevType {
				changed = true
				headphonesIn = (outputType == 2)
			}

			// Check headphone jack change (wired headphones on built-in)
			if jackState != prevJack && jackState >= 0 && prevJack >= 0 {
				changed = true
				headphonesIn = (jackState == 1)
			}

			if changed {
				if headphonesIn {
					events <- mood.HardwareEvent{
						Type:      mood.EventHeadphonesIn,
						Timestamp: time.Now(),
						Meta:      name,
					}
				} else {
					events <- mood.HardwareEvent{
						Type:      mood.EventHeadphonesOut,
						Timestamp: time.Now(),
						Meta:      name,
					}
				}
			}

			prevType = outputType
			prevJack = jackState
		}
	}
}
