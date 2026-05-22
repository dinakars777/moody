package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sort"

	"github.com/dinakars777/moody/sensors"
)

type sensorOptions struct {
	minAmplitude float64
	cooldownMs   int
	fastMode     bool
	verbose      bool
	noAccel      bool
	noUSB        bool
	noPower      bool
	noLid        bool
	noWiFi       bool
	noHeadphones bool
	noDisplay    bool
	noAI         bool
}

type doctorReport struct {
	Version       string           `json:"version"`
	OS            string           `json:"os"`
	Arch          string           `json:"arch"`
	RunningAsRoot bool             `json:"runningAsRoot"`
	Sensors       []sensors.Status `json:"sensors"`
}

func runDoctor(args []string) {
	doctorFlags := flag.NewFlagSet("doctor", flag.ExitOnError)
	jsonOutput := doctorFlags.Bool("json", false, "Print machine-readable JSON")
	if err := doctorFlags.Parse(args); err != nil {
		os.Exit(2)
	}

	report := newDoctorReport(sensorStatuses(buildSensors(sensorOptions{
		minAmplitude: 0.05,
		cooldownMs:   750,
	})))
	if *jsonOutput {
		writeJSON(report)
		return
	}
	printDoctorReport(report)
}

func buildSensors(opts sensorOptions) []sensors.Sensor {
	allSensors := []sensors.Sensor{}
	if !opts.noAccel {
		allSensors = append(allSensors, sensors.NewAccelerometer(opts.minAmplitude, opts.cooldownMs, opts.fastMode))
	}
	if !opts.noPower {
		allSensors = append(allSensors, sensors.NewPower())
	}
	if !opts.noUSB {
		allSensors = append(allSensors, sensors.NewUSB())
	}
	if !opts.noLid {
		allSensors = append(allSensors, sensors.NewLid())
	}
	if !opts.noWiFi {
		allSensors = append(allSensors, sensors.NewWiFi())
	}
	if !opts.noHeadphones {
		allSensors = append(allSensors, sensors.NewHeadphones())
	}
	if !opts.noDisplay {
		allSensors = append(allSensors, sensors.NewDisplay())
	}
	if !opts.noAI {
		allSensors = append(allSensors, sensors.NewAI(opts.verbose))
	}
	return allSensors
}

func sensorStatuses(allSensors []sensors.Sensor) []sensors.Status {
	statuses := make([]sensors.Status, 0, len(allSensors))
	for _, s := range allSensors {
		statuses = append(statuses, s.Status())
	}
	return statuses
}

func newDoctorReport(statuses []sensors.Status) doctorReport {
	return doctorReport{
		Version:       version,
		OS:            runtime.GOOS,
		Arch:          runtime.GOARCH,
		RunningAsRoot: os.Geteuid() == 0,
		Sensors:       statuses,
	}
}

func writeJSON(report doctorReport) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		log.Fatal(err)
	}
}

func printSensorList(statuses []sensors.Status) {
	fmt.Println("Available sensors:")
	for _, status := range statuses {
		fmt.Printf("  %-30s %s\n", status.Name, statusSummary(status))
	}
}

func printDoctorReport(report doctorReport) {
	fmt.Printf("moody doctor v%s\n", report.Version)
	fmt.Printf("System: %s/%s\n", report.OS, report.Arch)
	fmt.Printf("Running as root: %t\n", report.RunningAsRoot)
	fmt.Println()
	printSensorList(report.Sensors)
	fmt.Println()

	for _, status := range report.Sensors {
		if status.Available && status.Reason == "" && len(status.Details) == 0 {
			continue
		}
		fmt.Printf("%s\n", status.Name)
		if status.Reason != "" {
			fmt.Printf("  reason: %s\n", status.Reason)
		}
		if status.SuggestedFix != "" {
			fmt.Printf("  fix: %s\n", status.SuggestedFix)
		}
		if len(status.Details) > 0 {
			keys := make([]string, 0, len(status.Details))
			for key := range status.Details {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				fmt.Printf("  %s: %s\n", key, status.Details[key])
			}
		}
	}
}

func statusSummary(status sensors.Status) string {
	if !status.Supported {
		return "✗ unsupported"
	}
	if status.Available {
		return "✓ available"
	}
	if status.Optional {
		return "- optional unavailable"
	}
	return "✗ unavailable"
}
