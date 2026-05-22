package sensors

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dinakars777/moody/mood"
	"github.com/fsnotify/fsnotify"
)

// AI monitors AI IDE activity (Kiro, Cursor, Windsurf)
type AI struct {
	watcher *fsnotify.Watcher
	events  chan<- mood.HardwareEvent
	done    chan struct{}
	verbose bool
}

func NewAI(verbose bool) *AI {
	return &AI{
		done:    make(chan struct{}),
		verbose: verbose,
	}
}

func (a *AI) Name() string {
	return "AI IDE Monitor"
}

func (a *AI) Available() bool {
	// Check if Kiro hooks directory exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	kiroHooksDir := filepath.Join(homeDir, ".kiro", "hooks")
	if _, err := os.Stat(kiroHooksDir); err == nil {
		return true
	}

	// Could also check for Cursor/Windsurf directories
	return false
}

func (a *AI) Start(events chan<- mood.HardwareEvent) error {
	a.events = events

	var err error
	a.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// Watch Kiro hooks directory
	homeDir, _ := os.UserHomeDir()
	kiroHooksDir := filepath.Join(homeDir, ".kiro", "hooks")

	if err := a.watcher.Add(kiroHooksDir); err != nil {
		return err
	}

	if a.verbose {
		log.Printf("[ai] Watching %s for AI activity", kiroHooksDir)
	}

	go a.watch()

	return nil
}

func (a *AI) Stop() {
	close(a.done)
	if a.watcher != nil {
		a.watcher.Close()
	}
}

func (a *AI) watch() {
	for {
		select {
		case event, ok := <-a.watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				a.handleHookEvent(event.Name)
			}

		case err, ok := <-a.watcher.Errors:
			if !ok {
				return
			}
			if a.verbose {
				log.Printf("[ai] watcher error: %v", err)
			}

		case <-a.done:
			return
		}
	}
}

func (a *AI) handleHookEvent(filename string) {
	// Read the hook file to determine event type
	data, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	var hook map[string]interface{}
	if err := json.Unmarshal(data, &hook); err != nil {
		return
	}

	// Check if this is an agentStop event
	if when, ok := hook["when"].(map[string]interface{}); ok {
		if eventType, ok := when["type"].(string); ok {
			if eventType == "agentStop" {
				// AI finished generating code!
				a.events <- mood.HardwareEvent{
					Type:      mood.EventAIDone,
					Timestamp: time.Now(),
				}

				if a.verbose {
					log.Printf("[ai] AI finished generating code")
				}
			}
		}
	}
}
