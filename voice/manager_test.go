package voice

import (
	"strings"
	"testing"

	"github.com/dinakars777/moody/mood"
)

func TestBuiltinDramaticPackIsRegistered(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	manager := NewManager()
	info := manager.GetPackInfo("en_dramatic")
	if info == nil {
		t.Fatal("en_dramatic pack was not registered")
	}
	if info.Name != "Overly Dramatic" {
		t.Fatalf("name = %q, want %q", info.Name, "Overly Dramatic")
	}
	if info.NSFW {
		t.Fatal("en_dramatic should be SFW")
	}
	if err := manager.SetActive("en_dramatic"); err != nil {
		t.Fatalf("SetActive(en_dramatic) error = %v", err)
	}
}

func TestBuiltinDramaticPackCoversEventsAndMoods(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	manager := NewManager()
	if err := manager.SetActive("en_dramatic"); err != nil {
		t.Fatalf("SetActive(en_dramatic) error = %v", err)
	}

	for eventName := range validEventNames {
		for _, label := range []mood.MoodLabel{
			mood.MoodHappy,
			mood.MoodGrumpy,
			mood.MoodAnxious,
			mood.MoodDramatic,
			mood.MoodDeadInside,
		} {
			line := manager.GetLine(eventName, label)
			if strings.TrimSpace(line) == "" {
				t.Fatalf("line for %s/%s is empty", eventName, label)
			}
		}
	}
}
