package voice

import (
	"fmt"
	"os/exec"
	"sync"

	"github.com/dinakars777/moody/mood"
)

// macOS voice mapping per mood — each mood gets a distinct system voice
var moodVoices = map[mood.MoodLabel]string{
	mood.MoodHappy:      "Samantha",   // Cheerful female
	mood.MoodGrumpy:     "Alex",       // Deep male
	mood.MoodAnxious:    "Samantha",   // Same voice, faster rate
	mood.MoodDramatic:   "Daniel",     // British accent for theatrical flair
	mood.MoodDeadInside: "Fred",       // Robotic monotone
}

// Speech rate per mood
var moodRates = map[mood.MoodLabel]int{
	mood.MoodHappy:      200, // Normal-ish
	mood.MoodGrumpy:     170, // Slower, more deliberate
	mood.MoodAnxious:    260, // Fast, panicky
	mood.MoodDramatic:   150, // Slow for dramatic effect
	mood.MoodDeadInside: 140, // Slow, lifeless
}

// Player handles text-to-speech audio playback using macOS `say`
type Player struct {
	mu       sync.Mutex
	speaking bool
	enabled  bool
}

// NewPlayer creates an audio player
func NewPlayer(enabled bool) *Player {
	return &Player{enabled: enabled}
}

// Speak plays a voice line using macOS TTS with mood-appropriate voice
func (p *Player) Speak(text string, moodLabel mood.MoodLabel) {
	if !p.enabled || text == "" {
		return
	}

	p.mu.Lock()
	if p.speaking {
		p.mu.Unlock()
		return // Don't overlap speech
	}
	p.speaking = true
	p.mu.Unlock()

	go func() {
		defer func() {
			p.mu.Lock()
			p.speaking = false
			p.mu.Unlock()
		}()

		voice := moodVoices[moodLabel]
		if voice == "" {
			voice = "Samantha"
		}
		rate := moodRates[moodLabel]
		if rate == 0 {
			rate = 200
		}

		cmd := exec.Command("say",
			"-v", voice,
			"-r", fmt.Sprintf("%d", rate),
			text,
		)
		cmd.Run() // Blocks until speech finishes
	}()
}

// SpeakSync plays a voice line and waits for it to finish
func (p *Player) SpeakSync(text string, moodLabel mood.MoodLabel) {
	if !p.enabled || text == "" {
		return
	}

	voice := moodVoices[moodLabel]
	if voice == "" {
		voice = "Samantha"
	}
	rate := moodRates[moodLabel]
	if rate == 0 {
		rate = 200
	}

	cmd := exec.Command("say",
		"-v", voice,
		"-r", fmt.Sprintf("%d", rate),
		text,
	)
	cmd.Run()
}

// Stop interrupts any current speech
func (p *Player) Stop() {
	exec.Command("killall", "say").Run()
	p.mu.Lock()
	p.speaking = false
	p.mu.Unlock()
}

// SetEnabled enables or disables audio playback
func (p *Player) SetEnabled(enabled bool) {
	p.mu.Lock()
	p.enabled = enabled
	p.mu.Unlock()
}

// IsSpeaking returns true if currently speaking
func (p *Player) IsSpeaking() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.speaking
}

// ListVoices returns available macOS system voices
func ListVoices() ([]string, error) {
	out, err := exec.Command("say", "-v", "?").Output()
	if err != nil {
		return nil, err
	}
	// Just return raw output lines for now
	return []string{string(out)}, nil
}
