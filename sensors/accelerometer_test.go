package sensors

import (
	"os"
	"runtime"
	"testing"
)

func TestAccelerometerStatusMatchesSupportedPlatform(t *testing.T) {
	accel := NewAccelerometer(0, 0, false)
	status := accel.Status()
	want := runtime.GOOS == "darwin" && runtime.GOARCH == "arm64"

	if status.Supported != want {
		t.Fatalf("supported = %t, want %t", status.Supported, want)
	}
	if status.Available != accel.Available() {
		t.Fatalf("available = %t, Available() = %t", status.Available, accel.Available())
	}
	if want && os.Geteuid() != 0 && status.Available {
		t.Fatal("accelerometer should report unavailable without elevated privileges")
	}
}
