package tui

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dinakars777/moody/mood"
)

const dashboardWidth = 72

const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	dim     = "\033[2m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	white   = "\033[37m"
)

var spinnerFrames = []string{"◐", "◓", "◑", "◒"}
var pulseFrames = []string{"·", "✦", "✧", "✦"}

type moodTheme struct {
	color   string
	accent  string
	tagline string
	faces   []string
}

var themes = map[mood.MoodLabel]moodTheme{
	mood.MoodHappy: {
		color:   green,
		accent:  cyan,
		tagline: "cooperative, charged, weirdly optimistic",
		faces:   []string{"(^‿^)", "(^‿~)", "(•‿•)", "(^‿^)"},
	},
	mood.MoodGrumpy: {
		color:   red,
		accent:  yellow,
		tagline: "sarcasm rising, trust reserves low",
		faces:   []string{"(ಠ_ಠ)", "(¬_¬)", "(ಠ_ಠ)", "(>_>)"},
	},
	mood.MoodAnxious: {
		color:   yellow,
		accent:  red,
		tagline: "fans metaphorically spinning",
		faces:   []string{"(O_O)", "(o_O)", "(O_O)", "(O_o)"},
	},
	mood.MoodDramatic: {
		color:   magenta,
		accent:  cyan,
		tagline: "performing emotional damage in three acts",
		faces:   []string{"(T_T)", "(T∩T)", "(T_T)", "(ಥ_ಥ)"},
	},
	mood.MoodDeadInside: {
		color:   white,
		accent:  dim,
		tagline: "technically alive, spiritually buffering",
		faces:   []string{"(-_-)", "(-_-) z", "(x_x)", "(-_-)"},
	},
}

// Dashboard renders a live mood status in the terminal
type Dashboard struct {
	engine    *mood.Engine
	startTime time.Time
	lastLine  string
	packName  string
	verbose   bool
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
	theme := themes[label]
	frame := int(time.Since(d.startTime) / (250 * time.Millisecond))
	uptime := time.Since(d.startTime).Round(time.Second)
	face := theme.faces[frame%len(theme.faces)]
	spinner := spinnerFrames[frame%len(spinnerFrames)]
	pulse := pulseFrames[frame%len(pulseFrames)]

	var lastEvent string
	if evt := d.engine.LastEvent(); evt != nil {
		lastEvent = fmt.Sprintf("%s (%s)", mood.EventLabel(evt.Type), evt.Meta)
	} else {
		lastEvent = "Waiting for hardware drama"
	}
	if d.lastLine != "" {
		lastEvent = d.lastLine
	}

	var sb strings.Builder
	sb.WriteString("\033[2J\033[H") // Clear screen, cursor to top
	sb.WriteString(topBorder())
	sb.WriteString(row(fmt.Sprintf("%s %s MOODY LIVE %s %s %s %s",
		colorize(bold, "◢"), colorize(theme.color, spinner), colorize(theme.color, m.Emoji()),
		colorize(theme.color+bold, strings.ToUpper(string(label))), colorize(theme.accent, pulse), colorize(dim, theme.tagline))))
	sb.WriteString(separator())
	sb.WriteString(row(fmt.Sprintf("%s  Pack %s  Uptime %s  Events %s",
		colorize(theme.color+bold, face),
		colorize(cyan, d.packName),
		colorize(yellow, uptime.String()),
		colorize(magenta, fmt.Sprintf("%d", d.engine.EventCount())))))
	sb.WriteString(row(""))
	sb.WriteString(row(axisLine("Happiness", "♥", m.Happiness, green, frame)))
	sb.WriteString(row(axisLine("Energy", "⚡", m.Energy, blue, frame+1)))
	sb.WriteString(row(axisLine("Trust", "◆", m.Trust, magenta, frame+2)))
	sb.WriteString(separator())
	sb.WriteString(row(colorize(dim, "Latest signal")))
	for _, line := range wrapText(lastEvent, dashboardWidth) {
		sb.WriteString(row(colorize(theme.color, line)))
	}
	sb.WriteString(separator())
	sb.WriteString(row(fmt.Sprintf("%s %s    %s",
		colorize(dim, "Ctrl+C quits"),
		colorize(theme.accent, "•"),
		colorize(dim, "Mood drifts while events arrive"))))
	sb.WriteString(bottomBorder())

	return sb.String()
}

// SetLastLine sets the last voice response for display
func (d *Dashboard) SetLastLine(line string) {
	d.lastLine = line
}

func axisLine(name, icon string, value float64, color string, frame int) string {
	pct := percent(value)
	bar := animatedProgressBar(value, 28, color, frame)
	delta := moodDelta(value)
	return fmt.Sprintf("%-10s %s  %s  %3d%%  %s", name, colorize(color, icon), bar, pct, delta)
}

func animatedProgressBar(value float64, width int, color string, frame int) string {
	pos := int((value + 1) / 2 * float64(width))
	if pos < 0 {
		pos = 0
	}
	if pos > width {
		pos = width
	}

	var sb strings.Builder
	for i := 0; i < width; i++ {
		switch {
		case i < pos:
			sb.WriteString(colorize(color, "█"))
		case i == pos && pos > 0 && pos < width && frame%2 == 0:
			sb.WriteString(colorize(color+bold, "▓"))
		default:
			sb.WriteString(colorize(dim, "░"))
		}
	}
	return sb.String()
}

func moodDelta(value float64) string {
	switch {
	case value > 0.45:
		return colorize(green, "glowing")
	case value > 0.05:
		return colorize(cyan, "steady")
	case value > -0.35:
		return colorize(yellow, "touchy")
	default:
		return colorize(red, "critical")
	}
}

func percent(value float64) int {
	pct := int((value + 1) / 2 * 100)
	if pct < 0 {
		return 0
	}
	if pct > 100 {
		return 100
	}
	return pct
}

func row(content string) string {
	return "│ " + padRight(fitText(content, dashboardWidth), dashboardWidth) + " │\n"
}

func topBorder() string {
	return "╭" + strings.Repeat("─", dashboardWidth+2) + "╮\n"
}

func separator() string {
	return "├" + strings.Repeat("─", dashboardWidth+2) + "┤\n"
}

func bottomBorder() string {
	return "╰" + strings.Repeat("─", dashboardWidth+2) + "╯\n"
}

func wrapText(s string, width int) []string {
	if visibleLen(s) <= width {
		return []string{s}
	}

	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	var current string
	for _, word := range words {
		next := word
		if current != "" {
			next = current + " " + word
		}
		if visibleLen(next) <= width {
			current = next
			continue
		}
		if current != "" {
			lines = append(lines, current)
		}
		current = word
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func fitText(s string, width int) string {
	if visibleLen(s) <= width {
		return s
	}

	plain := stripANSI(s)
	runes := []rune(plain)
	if width <= 1 {
		return string(runes[:width])
	}
	return string(runes[:width-1]) + "…"
}

func padRight(s string, width int) string {
	padding := width - visibleLen(s)
	if padding <= 0 {
		return s
	}
	return s + strings.Repeat(" ", padding)
}

func colorize(code, s string) string {
	if !colorsEnabled() || code == "" {
		return s
	}
	return code + s + reset
}

func colorsEnabled() bool {
	return os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb"
}

func visibleLen(s string) int {
	return utf8.RuneCountInString(stripANSI(s))
}

func stripANSI(s string) string {
	var sb strings.Builder
	inEscape := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if inEscape {
			if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
				inEscape = false
			}
			continue
		}
		if ch == 0x1b {
			inEscape = true
			continue
		}
		sb.WriteByte(ch)
	}
	return sb.String()
}
