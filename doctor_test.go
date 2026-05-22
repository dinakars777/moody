package main

import (
	"runtime"
	"testing"

	"github.com/dinakars777/moody/mood"
	"github.com/dinakars777/moody/sensors"
)

type fakeSensor struct {
	status sensors.Status
}

func (f fakeSensor) Name() string { return f.status.Name }
func (f fakeSensor) Start(chan<- mood.HardwareEvent) error {
	return nil
}
func (f fakeSensor) Stop() {}
func (f fakeSensor) Available() bool {
	return f.status.Available
}
func (f fakeSensor) Status() sensors.Status {
	return f.status
}

func TestSensorStatusesUsesStructuredStatus(t *testing.T) {
	statuses := sensorStatuses([]sensors.Sensor{
		fakeSensor{status: sensors.Status{
			ID:        "fake",
			Name:      "Fake",
			Supported: true,
			Available: false,
			Reason:    "not real",
		}},
	})

	if len(statuses) != 1 {
		t.Fatalf("len(statuses) = %d, want 1", len(statuses))
	}
	if statuses[0].Reason != "not real" {
		t.Fatalf("reason = %q, want %q", statuses[0].Reason, "not real")
	}
}

func TestNewDoctorReportIncludesRuntime(t *testing.T) {
	report := newDoctorReport(nil)

	if report.Version != version {
		t.Fatalf("version = %q, want %q", report.Version, version)
	}
	if report.OS != runtime.GOOS {
		t.Fatalf("os = %q, want %q", report.OS, runtime.GOOS)
	}
	if report.Arch != runtime.GOARCH {
		t.Fatalf("arch = %q, want %q", report.Arch, runtime.GOARCH)
	}
}

func TestStatusSummary(t *testing.T) {
	tests := []struct {
		name   string
		status sensors.Status
		want   string
	}{
		{
			name:   "unsupported",
			status: sensors.Status{Supported: false},
			want:   "✗ unsupported",
		},
		{
			name:   "available",
			status: sensors.Status{Supported: true, Available: true},
			want:   "✓ available",
		},
		{
			name:   "optional unavailable",
			status: sensors.Status{Supported: true, Optional: true},
			want:   "- optional unavailable",
		},
		{
			name:   "unavailable",
			status: sensors.Status{Supported: true},
			want:   "✗ unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statusSummary(tt.status); got != tt.want {
				t.Fatalf("statusSummary() = %q, want %q", got, tt.want)
			}
		})
	}
}
