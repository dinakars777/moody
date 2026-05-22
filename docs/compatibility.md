# Compatibility

Moody is developed for macOS on Apple Silicon MacBooks. Sensor support varies by
hardware model, macOS version, permissions, and connected devices.

Run this before filing hardware reports:

```bash
moody doctor --json
```

## Current Matrix

| Area | Expected Support | Notes |
|------|------------------|-------|
| OS | macOS | Linux and Windows are unsupported. |
| CPU | Apple Silicon arm64 | Accelerometer support requires Apple Silicon. |
| Accelerometer | MacBook-class Apple Silicon with sudo | Run `sudo moody`, or use `--no-accel`. |
| Power/Battery | MacBooks with a battery | Desktop Macs may not report battery state. |
| USB | macOS IOKit USB registry | Counts USB device changes. |
| Lid | MacBooks exposing clamshell state | External-display setups may not expose this. |
| WiFi | macOS `networksetup` WiFi hardware port | Reports connected/disconnected state only. |
| Headphones | macOS CoreAudio default output route | Tracks default output route changes. |
| Display | macOS CoreGraphics active display list | Tracks display count changes. |
| AI IDE | Kiro hooks at `~/.kiro/hooks` | Cursor and Windsurf are not implemented yet. |

## Reporting A New Machine

Please include:

- Mac model and chip
- macOS version
- Whether Moody was run with `sudo`
- `moody doctor --json`
- Which sensor did not behave as expected
