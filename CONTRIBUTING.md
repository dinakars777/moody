# Contributing To Moody

Thanks for helping improve Moody. Keep changes small, tested, and easy to
review.

## Local Checks

Run these before opening a PR:

```bash
test -z "$(gofmt -l .)"
go test ./...
go vet ./...
go build ./...
```

For release config changes, also run:

```bash
go run github.com/goreleaser/goreleaser/v2@latest check
```

## Voice Packs

Validate local packs before submitting them:

```bash
go run . pack validate ./my-pack
```

Voice pack authors should follow [docs/voice-pack-authoring.md](docs/voice-pack-authoring.md).

When adding a built-in personality, update the interactive demo in `docs/` in
the same PR. This keeps the README, CLI, and demo aligned.

## Sensor Changes

Sensor changes should preserve partial startup. If one sensor is unavailable,
Moody should still run with the remaining available sensors.

Add or update structured diagnostics for any changed sensor:

```bash
go run . doctor
go run . doctor --json
```

## Pull Requests

Include:

- What changed
- How you verified it
- Any hardware/macOS assumptions
- `doctor --json` output for sensor compatibility changes
