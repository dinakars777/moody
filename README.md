# moody 🫠

Your MacBook has feelings. And it's not afraid to express them.

![Demo](demo.gif)

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

Download from [releases](https://github.com/dinakars777/moody/releases/latest), or build from source:

```bash
go install github.com/dinakars777/moody@latest
sudo cp "$(go env GOPATH)/bin/moody" /usr/local/bin/moody
```

## Usage

```bash
# Start moody (SFW mode)
sudo moody

# NSFW mode 😏
sudo moody --spicy

# Show live mood dashboard
sudo moody --dashboard

# List available sensors
sudo moody --list-sensors

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
- Go 1.22+ (if building from source)

## How It Works

1. Reads accelerometer data via IOKit HID to detect physical impacts
2. Monitors USB, power, battery, and lid state via IOKit
3. Monitors WiFi and Headphone connections using `networksetup` and `CoreAudio`
4. Monitors AI IDE activity (Kiro, Cursor, Windsurf) for code generation completion
5. Maintains a 3-axis mood engine (happiness, energy, trust)
6. Mood persists to `~/.moody/state.json` — your MacBook remembers
7. Selects personality-appropriate responses based on current mood
8. Speaks the response aloud using macOS Text-to-Speech (TTS) with mood-specific voices

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

# Use NSFW pack
sudo moody --spicy
```

**Built-in packs:**
- `en_default` — Passive-aggressive office coworker (SFW)
- `en_spicy` — Your MacBook is... very friendly (NSFW 🔞)

## Options

| Flag | Description |
|------|-------------|
| `--spicy` | Enable NSFW voice pack |
| `--pack <NAME>` | Use specific voice pack |
| `--dashboard` | Show live TUI mood dashboard |
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
| `--no-ai` | Disable AI IDE monitoring |
| `--verbose` | Log all events |
| `--list-sensors` | Show available sensors |
| `--packs` | List voice packs |

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

## License

MIT
