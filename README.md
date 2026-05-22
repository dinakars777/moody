# moody 🫠

Your MacBook has feelings. And it's not afraid to express them.

[Interactive demo](https://dinakars777.github.io/moody/)

## What Is This?

Every hardware event triggers a personality response:

- 👋 **Slap it** → it complains (and remembers)
- 🔌 **Plug in USB** → it gets curious
- ⚡ **Connect charger** → it sighs with relief
- 🪫 **Battery dying** → it begs for its life
- 📶 **WiFi drops** → existential crisis
- 🎧 **Plug in headphones** → "just the two of us now"
- 🤖 **AI finishes code** → celebrates your generated code

Your MacBook's **mood evolves** based on how you treat it.
Slap it too much? It gets grumpy. Charge it? It forgives you. Maybe.

## Install

Using Homebrew:

```bash
brew install dinakars777/tap/moody
```

Developer build from source:

```bash
go install github.com/dinakars777/moody@latest
sudo cp "$(go env GOPATH)/bin/moody" /usr/local/bin/moody
```

Release builds report the tagged version. Source builds may report `dev` unless
you build with `-ldflags "-X main.version=<version>"`.

Maintainers: see [docs/releasing.md](docs/releasing.md) for the tag release flow.

## Usage

```bash
# Start moody (SFW mode)
sudo moody

# NSFW mode 😏
sudo moody --spicy

# Show live animated mood dashboard
sudo moody --dashboard

# List available sensors
sudo moody --list-sensors

# Explain sensor compatibility and fixes
moody doctor
moody doctor --json

# Adjust slap sensitivity
sudo moody --min-amplitude 0.15

# Fast mode (quicker detection, shorter cooldown)
sudo moody --fast

# Silent mode (disables TTS voice, text only)
sudo moody --silent

# Verbose logging
sudo moody --verbose
```

## Requirements

- macOS on Apple Silicon (M2+ or M1 Pro)
- `sudo` (for accelerometer access)
- Go 1.26.1+ (if building from source)

See [docs/compatibility.md](docs/compatibility.md) for the sensor support matrix.

## How It Works

1. Reads accelerometer data via IOKit HID to detect physical impacts
2. Monitors USB, power, battery, and lid state via IOKit
3. Monitors WiFi and Headphone connections using `networksetup` and `CoreAudio`
4. Monitors external display changes
5. Monitors AI IDE activity (Kiro) for code generation completion
6. Maintains a 3-axis mood engine (happiness, energy, trust)
7. Mood persists to `~/.moody/state.json` — your MacBook remembers
8. Selects personality-appropriate responses based on current mood
9. Speaks the response aloud using macOS Text-to-Speech (TTS) with mood-specific voices

## The Mood System

Your MacBook's mood shifts with every event:

| Mood | Trigger | Personality |
|------|---------|-------------|
| 😊 Happy | Charged, USB in | Cheerful, friendly |
| 😤 Grumpy | Slapped, charger removed | Sarcastic, snippy |
| 😰 Anxious | Battery low, WiFi lost | Panicky, desperate |
| 🎭 Dramatic | Multiple negative events | Over-the-top theatrical |
| 💀 Dead Inside | Sustained abuse | Nihilistic, apathetic |

## Voice Packs

```bash
# List installed packs
moody --packs

# Validate a local community pack
moody pack validate ./my-pack

# Use NSFW pack
sudo moody --spicy
```

**Built-in packs:**
- `en_default` — Passive-aggressive office coworker (SFW)
- `en_spicy` — Your MacBook is... very friendly (NSFW 🔞)
- `ja_spicy` — Anime-inspired Japanese pack (NSFW 🔞)
- `hi_default` — Hindi default pack (SFW)
- `hi_spicy` — Hindi spicy pack (NSFW 🔞)
- `en_pirate` — Pirate speak pack (SFW)

## Options

| Flag | Description |
|------|-------------|
| `--spicy` | Enable NSFW voice pack |
| `--pack <NAME>` | Use specific voice pack |
| `--dashboard` | Show live animated TUI mood dashboard |
| `--mute` | Track mood without responses |
| `--silent` | Disable TTS audio (text output only) |
| `--fast` | Faster polling, shorter cooldown |
| `--min-amplitude <F>` | Accelerometer sensitivity (default: 0.05) |
| `--cooldown <MS>` | Min ms between responses (default: 750) |
| `--no-accel` | Disable accelerometer |
| `--no-usb` | Disable USB sensor |
| `--no-power` | Disable power sensor |
| `--no-lid` | Disable lid sensor |
| `--no-wifi` | Disable WiFi sensor |
| `--no-headphones` | Disable headphone sensor |
| `--no-display` | Disable external display sensor |
| `--no-ai` | Disable AI IDE monitoring |
| `--verbose` | Log all events |
| `--list-sensors` | Show available sensors |
| `--json` | Print JSON with `--list-sensors` |
| `--packs` | List voice packs |
| `--version` | Print version |

## Diagnostics

```bash
moody doctor
moody doctor --json
```

`doctor` reports OS, architecture, root status, and per-sensor support details.
If a sensor cannot run, the report includes a reason and suggested fix.

## AI IDE Integration

Moody can notify you when your AI coding assistant finishes generating code!

**Supported IDEs:**
- [Kiro](https://kiro.ai) - Automatically detected
- Cursor - Coming soon
- Windsurf - Coming soon

When AI finishes generating code, your Mac celebrates (or complains, depending on its mood).

**Related Projects:**
- [ai-done-hooks](https://github.com/dinakars777/ai-done-hooks) - Simple notification configs
- [ai-done](https://github.com/dinakars777/ai-done) - Standalone menu bar app

## Contributing

Contributions welcome! Especially:
- [ ] More voice packs (languages, personalities)
- [ ] More AI IDE integrations (Cursor, Windsurf)
- [ ] Gordon Ramsay voice pack
- [ ] HAL 9000 voice pack

When adding a built-in personality, update the interactive demo in `docs/` at the same time.

Voice pack authors can start with [docs/voice-pack-authoring.md](docs/voice-pack-authoring.md).
Please read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a PR.

## Privacy

Moody runs locally and does not send telemetry. See [docs/privacy.md](docs/privacy.md).

## License

MIT
