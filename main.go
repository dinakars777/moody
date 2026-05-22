package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dinakars777/moody/mood"
	"github.com/dinakars777/moody/sensors"
	"github.com/dinakars777/moody/tui"
	"github.com/dinakars777/moody/voice"
)

var (
	version = "dev"
)

func main() {
	// Subcommand handling
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		switch os.Args[1] {
		case "doctor":
			runDoctor(os.Args[2:])
			return
		case "pack":
			runPackCommand(os.Args[2:])
			return
		case "install":
			if len(os.Args) < 3 {
				fmt.Println("Usage: moody install <git-url>")
				os.Exit(1)
			}
			if err := voice.InstallPack(os.Args[2]); err != nil {
				log.Fatal(err)
			}
			return
		case "packs":
			voiceMgr := voice.NewManager()
			fmt.Println("Installed voice packs:")
			for _, name := range voiceMgr.ListPacks() {
				info := voiceMgr.GetPackInfo(name)
				nsfw := ""
				if info.NSFW {
					nsfw = " 🔞"
				}
				fmt.Printf("  %-20s %s%s\n", name, info.Description, nsfw)
			}
			return
		}
	}

	// CLI flags
	spicy := flag.Bool("spicy", false, "Enable NSFW voice pack 😏")
	pack := flag.String("pack", "", "Use a specific voice pack (e.g., en_spicy)")
	dashboard := flag.Bool("dashboard", false, "Show live mood dashboard in terminal")
	mute := flag.Bool("mute", false, "Track mood without playing audio/text")
	fast := flag.Bool("fast", false, "Faster polling, shorter cooldown")
	minAmplitude := flag.Float64("min-amplitude", 0.05, "Accelerometer sensitivity (lower = more sensitive)")
	cooldown := flag.Int("cooldown", 750, "Minimum ms between responses")
	noAccel := flag.Bool("no-accel", false, "Disable accelerometer sensor")
	noUSB := flag.Bool("no-usb", false, "Disable USB sensor")
	noPower := flag.Bool("no-power", false, "Disable power/battery sensor")
	noLid := flag.Bool("no-lid", false, "Disable lid sensor")
	noWiFi := flag.Bool("no-wifi", false, "Disable WiFi sensor")
	noHeadphones := flag.Bool("no-headphones", false, "Disable headphone sensor")
	noDisplay := flag.Bool("no-display", false, "Disable display sensor")
	noAI := flag.Bool("no-ai", false, "Disable AI IDE monitoring")
	silent := flag.Bool("silent", false, "Disable TTS audio (text output only)")
	verbose := flag.Bool("verbose", false, "Log all sensor events to stderr")
	listSensors := flag.Bool("list-sensors", false, "List detected sensors and exit")
	showPacks := flag.Bool("packs", false, "List installed voice packs and exit")
	showVersion := flag.Bool("version", false, "Print version and exit")
	jsonOutput := flag.Bool("json", false, "Print machine-readable JSON for diagnostics")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `moody 🫠 — Your MacBook has feelings.

Usage: sudo moody [flags]

Every hardware event triggers a personality response.
Your MacBook's mood evolves based on how you treat it.

`)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  sudo moody                    # Start with default SFW pack
  sudo moody --spicy            # NSFW mode 😏
  sudo moody --dashboard        # Show live mood dashboard
  sudo moody --fast             # More responsive detection
  sudo moody --min-amplitude 0.1  # More sensitive slap detection

`)
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("moody v%s\n", version)
		return
	}

	// Initialize voice manager
	voiceMgr := voice.NewManager()

	if *showPacks {
		fmt.Println("Installed voice packs:")
		for _, name := range voiceMgr.ListPacks() {
			info := voiceMgr.GetPackInfo(name)
			nsfw := ""
			if info.NSFW {
				nsfw = " 🔞"
			}
			fmt.Printf("  %-15s %s%s\n", name, info.Description, nsfw)
		}
		return
	}

	// Select voice pack
	if *pack != "" {
		if err := voiceMgr.SetActive(*pack); err != nil {
			log.Fatal(err)
		}
	} else if *spicy {
		if err := voiceMgr.SetActive("en_spicy"); err != nil {
			log.Fatal(err)
		}
	}

	activePack := voiceMgr.ActivePack()
	packInfo := voiceMgr.GetPackInfo(activePack)

	allSensors := buildSensors(sensorOptions{
		minAmplitude: *minAmplitude,
		cooldownMs:   *cooldown,
		fastMode:     *fast,
		verbose:      *verbose,
		noAccel:      *noAccel,
		noUSB:        *noUSB,
		noPower:      *noPower,
		noLid:        *noLid,
		noWiFi:       *noWiFi,
		noHeadphones: *noHeadphones,
		noDisplay:    *noDisplay,
		noAI:         *noAI,
	})

	if *listSensors {
		report := newDoctorReport(sensorStatuses(allSensors))
		if *jsonOutput {
			writeJSON(report)
		} else {
			printSensorList(report.Sensors)
		}
		return
	}

	// Initialize mood engine
	engine := mood.NewEngine()
	defer engine.Shutdown()

	// Initialize audio player (TTS)
	player := voice.NewPlayer(!*mute && !*silent)
	defer player.Stop()

	// Set the language for TTS voice selection
	if packInfo != nil {
		player.SetLanguage(packInfo.Language)
	}

	// Event channel
	events := make(chan mood.HardwareEvent, 32)

	// Start sensors
	activeSensorCount := 0
	skippedSensorCount := 0
	for _, s := range allSensors {
		status := s.Status()
		if !status.Available {
			skippedSensorCount++
			if *verbose {
				log.Printf("[sensor] %s: not available, skipping: %s", s.Name(), status.Reason)
			}
			continue
		}
		if err := s.Start(events); err != nil {
			skippedSensorCount++
			if *verbose {
				log.Printf("[sensor] %s: failed to start: %v", s.Name(), err)
			}
			continue
		}
		activeSensorCount++
		if *verbose {
			log.Printf("[sensor] %s: started", s.Name())
		}
	}

	if activeSensorCount == 0 {
		log.Fatal("No sensors started. Run `moody doctor` for details.")
	}

	printStartupBanner(packInfo, activePack, activeSensorCount, len(allSensors), skippedSensorCount)

	// Dashboard (optional)
	var dash *tui.Dashboard
	if *dashboard {
		dash = tui.NewDashboard(engine, activePack, *verbose)
	}

	// Signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Dashboard refresh ticker
	var dashTicker *time.Ticker
	if *dashboard {
		dashTicker = time.NewTicker(500 * time.Millisecond)
		defer dashTicker.Stop()
	}

	// Main event loop
	for {
		if *dashboard && dashTicker != nil {
			select {
			case evt := <-events:
				processEvent(engine, voiceMgr, player, dash, evt, *mute, *verbose)
			case <-dashTicker.C:
				fmt.Print(dash.Render())
			case <-sigCh:
				shutdown(engine, allSensors)
				return
			}
		} else {
			select {
			case evt := <-events:
				processEvent(engine, voiceMgr, player, nil, evt, *mute, *verbose)
			case <-sigCh:
				shutdown(engine, allSensors)
				return
			}
		}
	}
}

func printStartupBanner(packInfo *voice.Manifest, activePack string, activeSensorCount, configuredSensorCount, skippedSensorCount int) {
	if packInfo != nil && packInfo.NSFW {
		fmt.Println("🫠 moody v" + version + " — " + packInfo.Description + " 🔞")
	} else {
		fmt.Println("🫠 moody v" + version + " — Your MacBook has feelings.")
	}
	fmt.Printf("   Pack: %s | Sensors: %d active", activePack, activeSensorCount)
	if skippedSensorCount > 0 {
		fmt.Printf(", %d skipped", skippedSensorCount)
	}
	fmt.Printf(" of %d configured\n", configuredSensorCount)
	if skippedSensorCount > 0 {
		fmt.Println("   Run `moody doctor` for sensor details.")
	}
	fmt.Println("   Press Ctrl+C to quit")
	fmt.Println()
}

func processEvent(engine *mood.Engine, voiceMgr *voice.Manager, player *voice.Player, dash *tui.Dashboard, evt mood.HardwareEvent, mute, verbose bool) {
	// Update mood
	moodLabel := engine.ProcessEvent(evt)

	// Get voice line
	eventName := mood.EventName(evt.Type)
	line := voiceMgr.GetLine(eventName, moodLabel)
	audioPath := voiceMgr.GetAudioPath(eventName)

	if verbose {
		log.Printf("[event] %s | mood: %s | eventName: %s | activePack: %s | response: %s",
			mood.EventLabel(evt.Type), moodLabel, eventName, voiceMgr.ActivePack(), line)
	}

	// Display/play response
	if !mute && line != "" {
		m := engine.CurrentMood()
		// Color-code the output based on mood
		color := moodColor(moodLabel)
		fmt.Printf("%s %s%s: %s%s\n", m.Emoji(), color,
			strings.ToUpper(string(moodLabel)), line, "\033[0m")

		// Play it via MP3 or TTS
		if player != nil {
			if audioPath != "" {
				player.PlayFile(audioPath)
			} else {
				player.Speak(line, moodLabel)
			}
		}
	}

	// Update dashboard
	if dash != nil && line != "" {
		dash.SetLastLine(fmt.Sprintf("%s → %s", mood.EventLabel(evt.Type), line))
	}
}

func moodColor(label mood.MoodLabel) string {
	switch label {
	case mood.MoodHappy:
		return "\033[32m" // Green
	case mood.MoodGrumpy:
		return "\033[31m" // Red
	case mood.MoodAnxious:
		return "\033[33m" // Yellow
	case mood.MoodDramatic:
		return "\033[35m" // Magenta
	case mood.MoodDeadInside:
		return "\033[90m" // Dark gray
	default:
		return "\033[0m"
	}
}

func shutdown(engine *mood.Engine, allSensors []sensors.Sensor) {
	fmt.Println("\n\n🫠 moody shutting down. Your MacBook will remember this.")
	for _, s := range allSensors {
		s.Stop()
	}
	engine.Shutdown()
}
