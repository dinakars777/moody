package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/dinakars777/moody/mood"
)

// Dashboard renders a live mood status in the terminal
type Dashboard struct {
	engine    *mood.Engine
	startTime time.Time
	lastLine   string
	packName   string
	verbose    bool
}

// NewDashboard creates a new TUI dashboard
func NewDashboard(engine *mood.Engine, packName string, verbose bool) *Dashboard {
	return &Dashboard{
		engine:    engine,
		startTime: time.Now(),
		packName:  packName,
		verbose:   verbose,
	}
}

// Render returns the dashboard as a string
func (d *Dashboard) Render() string {
	m := d.engine.CurrentMood()
	label := m.Label()
	emoji := m.Emoji()
	uptime := time.Since(d.startTime).Round(time.Second)

	// Progress bars for each axis
	hBar := progressBar(m.Happiness, 20)
	eBar := progressBar(m.Energy, 20)
	tBar := progressBar(m.Trust, 20)

	// Happiness percentage (mapped from -1..1 to 0..100)
	hPct := int((m.Happiness + 1) / 2 * 100)
	ePct := int((m.Energy + 1) / 2 * 100)
	tPct := int((m.Trust + 1) / 2 * 100)

	var lastEvent string
	if evt := d.engine.LastEvent(); evt != nil {
		lastEvent = fmt.Sprintf("%s (%s)", mood.EventLabel(evt.Type), evt.Meta)
	} else {
		lastEvent = "None yet"
	}

	if d.lastLine != "" {
		lastEvent = d.lastLine
	}

	var sb strings.Builder
	sb.WriteString("\033[2J\033[H") // Clear screen, cursor to top

	sb.WriteString("┌─────────────────────────────────────────────┐\n")
	sb.WriteString(fmt.Sprintf("│          %s  %-12s                   │\n", emoji, strings.ToUpper(string(label))))
	sb.WriteString("│                                             │\n")
	sb.WriteString(fmt.Sprintf("│  Happiness: %s %3d%%          │\n", hBar, hPct))
	sb.WriteString(fmt.Sprintf("│  Energy:    %s %3d%%          │\n", eBar, ePct))
	sb.WriteString(fmt.Sprintf("│  Trust:     %s %3d%%          │\n", tBar, tPct))
	sb.WriteString("│                                             │\n")
	sb.WriteString(fmt.Sprintf("│  Events: %-5d  │  Uptime: %-14s  │\n", d.engine.EventCount(), uptime))
	sb.WriteString(fmt.Sprintf("│  Pack: %-12s                        │\n", d.packName))
	sb.WriteString("│                                             │\n")
	sb.WriteString(fmt.Sprintf("│  Last: %-36s │\n", truncate(lastEvent, 36)))
	sb.WriteString("└─────────────────────────────────────────────┘\n")
	sb.WriteString("\n  Press Ctrl+C to quit\n")

	return sb.String()
}

// SetLastLine sets the last voice response for display
func (d *Dashboard) SetLastLine(line string) {
	d.lastLine = line
}

// progressBar renders a bar from -1.0 to 1.0 as a fixed-width string
func progressBar(value float64, width int) string {
	// Map -1..1 to 0..width
	pos := int((value + 1) / 2 * float64(width))
	if pos < 0 {
		pos = 0
	}
	if pos > width {
		pos = width
	}

	filled := strings.Repeat("█", pos)
	empty := strings.Repeat("░", width-pos)
	return filled + empty
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
