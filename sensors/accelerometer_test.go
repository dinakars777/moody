package sensors

import (
	"runtime"
	"testing"
)

func TestAccelerometerAvailabilityMatchesSupportedPlatform(t *testing.T) {
	got := NewAccelerometer(0, 0, false).Available()
	want := runtime.GOOS == "darwin" && runtime.GOARCH == "arm64"

	if got != want {
		t.Fatalf("available = %t, want %t", got, want)
	}
}
