package voice

import (
	"fmt"
	"os/exec"
	"sync"

	"github.com/dinakars777/moody/mood"
)

// macOS voice mapping per mood — default to female (Samantha) for all moods
// to maintain a consistent female personality across the app
var moodVoices = map[mood.MoodLabel]string{
	mood.MoodHappy:      "Samantha",
	mood.MoodGrumpy:     "Samantha",
	mood.MoodAnxious:    "Samantha",
	mood.MoodDramatic:   "Samantha",
	mood.MoodDeadInside: "Samantha",
}

// Speech rate per mood
var moodRates = map[mood.MoodLabel]int{
	mood.MoodHappy:      200, // Normal-ish
	mood.MoodGrumpy:     140, // Slower, more deliberate and annoyed
	mood.MoodAnxious:    260, // Fast, panicky
	mood.MoodDramatic:   150, // Slow for dramatic effect
	mood.MoodDeadInside: 120, // Very slow, lifeless
}

// Player handles text-to-speech audio playback using macOS `say`
type Player struct {
	mu       sync.Mutex
	speaking bool
	enabled  bool
	language string // Language code for voice selection
	current  *exec.Cmd
}

// NewPlayer creates an audio player
func NewPlayer(enabled bool) *Player {
	return &Player{enabled: enabled, language: "en"}
}

// SetLanguage sets the language for voice selection
func (p *Player) SetLanguage(lang string) {
	p.mu.Lock()
	p.language = lang
	p.mu.Unlock()
}

// Speak plays a voice line using macOS TTS with mood-appropriate voice
func (p *Player) Speak(text string, moodLabel mood.MoodLabel) {
	if text == "" {
		return
	}

	voice := p.getVoiceForLanguage(moodLabel)
	rate := moodRates[moodLabel]
	if rate == 0 {
		rate = 200
	}

	cmd := exec.Command("say",
		"-v", voice,
		"-r", fmt.Sprintf("%d", rate),
		text,
	)
	p.start(cmd)
}

// SpeakSync plays a voice line and waits for it to finish
func (p *Player) SpeakSync(text string, moodLabel mood.MoodLabel) {
	if !p.enabled || text == "" {
		return
	}

	voice := p.getVoiceForLanguage(moodLabel)
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

// getVoiceForLanguage returns the appropriate voice based on language
func (p *Player) getVoiceForLanguage(moodLabel mood.MoodLabel) string {
	p.mu.Lock()
	lang := p.language
	p.mu.Unlock()

	// Language-specific voices
	switch lang {
	case "hi": // Hindi
		return "Lekha"
	case "ja": // Japanese
		return "Kyoko"
	case "en": // English
		fallthrough
	default:
		voice := moodVoices[moodLabel]
		if voice == "" {
			voice = "Samantha"
		}
		return voice
	}
}

// PlayFile plays an audio file using macOS `afplay`
func (p *Player) PlayFile(path string) {
	if path == "" {
		return
	}

	cmd := exec.Command("afplay", path)
	p.start(cmd)
}

// Stop interrupts any current speech
func (p *Player) Stop() {
	p.mu.Lock()
	cmd := p.current
	p.current = nil
	p.speaking = false
	p.mu.Unlock()

	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
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

func (p *Player) start(cmd *exec.Cmd) {
	p.mu.Lock()
	if !p.enabled || p.speaking {
		p.mu.Unlock()
		return
	}
	if err := cmd.Start(); err != nil {
		p.mu.Unlock()
		return
	}
	p.current = cmd
	p.speaking = true
	p.mu.Unlock()

	go p.wait(cmd)
}

func (p *Player) wait(cmd *exec.Cmd) {
	defer func() {
		p.mu.Lock()
		if p.current == cmd {
			p.current = nil
			p.speaking = false
		}
		p.mu.Unlock()
	}()

	_ = cmd.Wait()
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
