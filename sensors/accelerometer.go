package sensors

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/dinakars777/moody/mood"
	"github.com/taigrr/apple-silicon-accelerometer/detector"
	"github.com/taigrr/apple-silicon-accelerometer/sensor"
	"github.com/taigrr/apple-silicon-accelerometer/shm"
)

// Accelerometer detects physical impacts on an Apple Silicon MacBook
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

func (a *Accelerometer) Name() string { return "Accelerometer (Apple Silicon)" }

func (a *Accelerometer) Available() bool {
	return runtime.GOOS == "darwin" && runtime.GOARCH == "arm64"
}

func (a *Accelerometer) Start(events chan<- mood.HardwareEvent) error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return nil
	}
	a.running = true
	a.mu.Unlock()

	// Create shared memory for accelerometer data
	accelRing, err := shm.CreateRing("accel_moody")
	if err != nil {
		a.mu.Lock()
		a.running = false
		a.mu.Unlock()
		return fmt.Errorf("creating accel shm: %w", err)
	}

	// Wait briefly for immediate launch failures from the background sensor worker.
	sensorErr := make(chan error, 1)

	// Fire up the sensor.Run() function which needs CFRunLoop.
	// It blocks forever, so we run it in a goroutine.
	go func() {
		err := sensor.Run(sensor.Config{
			AccelRing: accelRing,
		})
		if err != nil {
			sensorErr <- err
		}
	}()

	select {
	case err := <-sensorErr:
		a.mu.Lock()
		a.running = false
		a.mu.Unlock()
		accelRing.Close()
		accelRing.Unlink()
		return fmt.Errorf("sensor launch failed: %w", err)
	case <-time.After(500 * time.Millisecond):
	}

	go a.pollLoop(events, accelRing)
	return nil
}

func (a *Accelerometer) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.running {
		close(a.stopCh)
		a.running = false
		// Note: sensor.Run() loops forever, but our program is typically exiting anyway.
		// In a long-running robust daemon we'd signal the CFRunLoop to stop.
	}
}

func (a *Accelerometer) pollLoop(events chan<- mood.HardwareEvent, accelRing *shm.RingBuffer) {
	defer accelRing.Close()
	defer accelRing.Unlink()

	det := detector.New()
	var lastAccelTotal uint64
	var lastEventTime time.Time

	pollInterval := 10 * time.Millisecond
	if a.fastMode {
		pollInterval = 4 * time.Millisecond
	}
	maxBatch := 200
	if a.fastMode {
		maxBatch = 320
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	var lastTrigger time.Time

	for {
		select {
		case <-a.stopCh:
			return
		case <-ticker.C:
			now := time.Now()
			tNow := float64(now.UnixNano()) / 1e9

			samples, newTotal := accelRing.ReadNew(lastAccelTotal, shm.AccelScale)
			lastAccelTotal = newTotal

			if len(samples) > maxBatch {
				samples = samples[len(samples)-maxBatch:]
			}

			nSamples := len(samples)
			for idx, sample := range samples {
				tSample := tNow - float64(nSamples-idx-1)/float64(det.FS)
				det.Process(sample.X, sample.Y, sample.Z, tSample)
			}

			if len(det.Events) == 0 {
				continue
			}

			// Get latest event
			ev := det.Events[len(det.Events)-1]
			if ev.Time.Equal(lastEventTime) {
				continue
			}
			lastEventTime = ev.Time

			// Cooldown & Amplitude check
			if ev.Amplitude < a.minAmplitude {
				continue
			}
			if time.Since(lastTrigger) < time.Duration(a.cooldownMs)*time.Millisecond {
				continue
			}
			lastTrigger = now

			// Normalize intensity to 0.0-1.0 range
			intensity := math.Min(ev.Amplitude/0.8, 1.0)

			events <- mood.HardwareEvent{
				Type:      mood.EventSlap,
				Intensity: intensity,
				Timestamp: now,
				Meta:      fmt.Sprintf("%g g", ev.Amplitude),
			}
		}
	}
}
