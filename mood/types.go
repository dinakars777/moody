package mood

import "fmt"

// MoodLabel represents the derived emotional state
type MoodLabel string

const (
	MoodHappy      MoodLabel = "happy"
	MoodGrumpy     MoodLabel = "grumpy"
	MoodAnxious    MoodLabel = "anxious"
	MoodDramatic   MoodLabel = "dramatic"
	MoodDeadInside MoodLabel = "dead_inside"
)

// Mood represents the MacBook's emotional state across 3 axes
type Mood struct {
	Happiness float64 `json:"happiness"` // -1.0 (miserable) to 1.0 (ecstatic)
	Energy    float64 `json:"energy"`    // -1.0 (exhausted) to 1.0 (hyper)
	Trust     float64 `json:"trust"`     // -1.0 (betrayed) to 1.0 (loyal)
}

// Label derives a mood label from the current state
func (m Mood) Label() MoodLabel {
	avg := (m.Happiness + m.Energy + m.Trust) / 3.0

	switch {
	case avg > 0.3:
		return MoodHappy
	case avg > 0.0:
		if m.Energy > 0.3 {
			return MoodHappy
		}
		if m.Trust < -0.2 {
			return MoodGrumpy
		}
		return MoodHappy
	case avg > -0.3:
		if m.Happiness < -0.3 {
			return MoodGrumpy
		}
		if m.Energy < -0.3 {
			return MoodAnxious
		}
		return MoodGrumpy
	case avg > -0.6:
		return MoodDramatic
	default:
		return MoodDeadInside
	}
}

// Emoji returns the emoji for the current mood
func (m Mood) Emoji() string {
	switch m.Label() {
	case MoodHappy:
		return "😊"
	case MoodGrumpy:
		return "😤"
	case MoodAnxious:
		return "😰"
	case MoodDramatic:
		return "🎭"
	case MoodDeadInside:
		return "💀"
	default:
		return "🫠"
	}
}

func (m Mood) String() string {
	return fmt.Sprintf("%s %s (H:%.2f E:%.2f T:%.2f)",
		m.Emoji(), m.Label(), m.Happiness, m.Energy, m.Trust)
}

// clamp keeps a value between -1.0 and 1.0
func clamp(v float64) float64 {
	if v > 1.0 {
		return 1.0
	}
	if v < -1.0 {
		return -1.0
	}
	return v
}
