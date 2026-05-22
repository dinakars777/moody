package mood

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProcessEventScalesSlapIntensity(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	engine := NewEngine()
	defer engine.Shutdown()

	label := engine.ProcessEvent(HardwareEvent{
		Type:      EventSlap,
		Intensity: 0.5,
		Timestamp: time.Now(),
	})

	if label != MoodHappy {
		t.Fatalf("label = %q, want %q", label, MoodHappy)
	}

	got := engine.CurrentMood()
	assertFloat(t, got.Happiness, 0.425)
	assertFloat(t, got.Energy, 0.525)
	assertFloat(t, got.Trust, 0.45)

	if engine.EventCount() != 1 {
		t.Fatalf("event count = %d, want 1", engine.EventCount())
	}
	if engine.LastEvent() == nil || engine.LastEvent().Type != EventSlap {
		t.Fatalf("last event was not recorded")
	}
}

func TestShutdownPersistsStateAndIsIdempotent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	engine := NewEngine()
	engine.ProcessEvent(HardwareEvent{
		Type:      EventBatteryCrit,
		Timestamp: time.Now(),
	})

	engine.Shutdown()
	engine.Shutdown()

	data, err := os.ReadFile(filepath.Join(home, ".moody", "state.json"))
	if err != nil {
		t.Fatalf("read state: %v", err)
	}

	var state persistedState
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("decode state: %v", err)
	}

	if state.EventCount != 1 {
		t.Fatalf("event count = %d, want 1", state.EventCount)
	}
	assertFloat(t, state.Mood.Happiness, 0.3)
	assertFloat(t, state.Mood.Energy, 0.25)
	assertFloat(t, state.Mood.Trust, 0.4)
}

func assertFloat(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.0001 {
		t.Fatalf("got %f, want %f", got, want)
	}
}
