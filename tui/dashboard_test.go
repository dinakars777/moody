package tui

import (
	"strings"
	"testing"
)

func TestPercentClampsMoodValue(t *testing.T) {
	tests := []struct {
		value float64
		want  int
	}{
		{value: -2, want: 0},
		{value: -1, want: 0},
		{value: 0, want: 50},
		{value: 1, want: 100},
		{value: 2, want: 100},
	}

	for _, tt := range tests {
		if got := percent(tt.value); got != tt.want {
			t.Fatalf("percent(%v) = %d, want %d", tt.value, got, tt.want)
		}
	}
}

func TestVisibleLenIgnoresANSI(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")

	colored := colorize(green+bold, "Mood")
	if got := visibleLen(colored); got != 4 {
		t.Fatalf("visibleLen(%q) = %d, want 4", colored, got)
	}
}

func TestFitTextUsesVisibleWidth(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")

	text := colorize(cyan, "Happiness overload")
	got := fitText(text, 10)
	if got != "Happiness…" {
		t.Fatalf("fitText() = %q, want %q", got, "Happiness…")
	}
}

func TestAnimatedProgressBarHasStableVisibleWidth(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")

	bar := animatedProgressBar(0.25, 12, green, 0)
	if got := visibleLen(bar); got != 12 {
		t.Fatalf("visibleLen(bar) = %d, want 12", got)
	}
}

func TestRowKeepsFixedVisibleWidth(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	line := row("Short")
	if !strings.HasPrefix(line, "│ ") || !strings.HasSuffix(line, " │\n") {
		t.Fatalf("row has unexpected frame: %q", line)
	}
	want := dashboardWidth + 4
	if got := visibleLen(strings.TrimSuffix(line, "\n")); got != want {
		t.Fatalf("visible row width = %d, want %d", got, want)
	}
}
