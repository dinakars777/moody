# Moody Incremental Improvement Plan

Created: 2026-05-22

## Current Baseline

Moody is now in a cleaner state than the initial audit:

- PR #9 added Go CI, removed committed local artifacts, fixed README drift, and formatted the Go tree.
- PR #10 made shutdown safer, stopped killing unrelated `say`/`afplay` processes, and added the first focused tests.
- PR #8 merged the community `en_pirate` voice pack.
- Issues #2, #4, and #7 are closed as implemented.

Remaining open work is mostly product maturity: distribution, diagnostics, compatibility reporting, tests around OS-specific code, voice-pack safety, and community contribution flow.

## Research Summary

Comparable projects and docs point to 5 themes:

- Release automation should be tag-driven. GoReleaser documents GitHub Actions release workflows that run on tags and need `contents: write` permissions.
- Homebrew distribution should start with a personal tap. Homebrew core has acceptability and notability expectations that are too heavyweight for a niche macOS-only sensor CLI at this stage.
- Sensor CLIs need explicit diagnostics. Projects like `mactop` and Apple Silicon accelerometer tools make hardware support, unsupported models, permissions, and raw probe data visible instead of hiding them behind a boolean.
- Voice packs should be treated as data packages. `xbar` and its plugin ecosystem are a useful model: clear folder conventions, validation, contribution docs, and a low-friction plugin path.
- OS-specific code needs layered tests. Go build constraints let the project separate pure logic from macOS/cgo probes, while GitHub-hosted macOS runners can verify Darwin builds.

## Incremental Roadmap

### 1. Release And Install Path

Goal: make installation reliable without committed binaries or manual copying.

Deliverables:

- Add `.goreleaser.yaml` for a `darwin_arm64` CLI artifact.
- Add a tag-triggered `.github/workflows/release.yml`.
- Inject the version with `-ldflags "-X main.version={{.Version}}"`.
- Publish checksums with every release.
- Create or reuse `dinakars777/homebrew-tap`.
- Publish a Homebrew formula first. Move to a cask later if signing, notarization, or app-bundle packaging becomes necessary.
- Update README install docs to prefer `brew install dinakars777/tap/moody`, with `go install` as the developer fallback.

Acceptance checks:

- A `vX.Y.Z` tag creates a GitHub release with a `darwin_arm64` binary and checksum.
- Fresh install via Homebrew can run `moody --version`.
- README no longer points users at empty release assets.

### 2. Runtime Diagnostics

Goal: users can understand why a sensor is unavailable without reading source code.

Deliverables:

- Replace `Available() bool` with structured sensor status: `supported`, `available`, `reason`, `suggestedFix`, and optional raw probe fields.
- Add `moody doctor`.
- Add `moody --list-sensors --json`.
- Probe at least: OS, architecture, root/euid, Mac model, accelerometer probe status, power source access, WiFi interface, audio route, display count, and Kiro hooks directory.
- Make startup degrade gracefully when accelerometer needs root but other sensors can run.

Acceptance checks:

- Running without sudo clearly says accelerometer needs elevated access, while non-root sensors still start where possible.
- `moody doctor --json` produces a compatibility report suitable for GitHub issues.
- `--list-sensors` no longer claims accelerometer availability solely from compile-time platform support.

### 3. Test And CI Layers

Goal: keep macOS integration moving without making every test require privileged hardware.

Deliverables:

- Add tests for mood labels, persistence, event names, and response selection.
- Add tests for voice pack registration and manifest parsing.
- Add fake command/filesystem seams for voice playback and pack installation.
- Split macOS/cgo probes behind build constraints or small interfaces.
- Expand CI into a cheap pure-test job and a macOS build job.

Acceptance checks:

- `go test ./...` covers pure logic without requiring sudo or real sensors.
- macOS CI still builds cgo sensor packages.
- New sensors can be tested with fake probe data.

### 4. Voice Pack Platform

Goal: make community packs safe to install and easy to author.

Deliverables:

- Define `voice/pack.schema.json`.
- Add `moody pack validate <path>`.
- Add `moody pack init <name>`.
- Extend manifests with `minMoodyVersion`, `license`, `homepage`, `nsfw`, supported events, supported audio formats, and optional checksums.
- Install packs atomically only after validation.
- Prefer release zip plus checksum for remote installs. Keep raw `git clone` as an explicit advanced path.
- Add `docs/voice-pack-authoring.md`.

Acceptance checks:

- Invalid manifests fail before files are copied into `~/.moody/packs`.
- A generated starter pack validates.
- README points pack authors to the authoring guide.

### 5. Community And Compatibility

Goal: make the project easier to trust, debug, and contribute to.

Deliverables:

- Add `CONTRIBUTING.md`.
- Add GitHub issue templates for sensor compatibility, voice pack proposal, and bug report.
- Add a compatibility matrix covering tested Mac models and macOS versions.
- Add a troubleshooting section based on `moody doctor`.
- Add a short privacy note: Moody reads local hardware state and local voice-pack files; it should not send telemetry unless a future feature explicitly says so.
- Replace or supplement the web demo with a small GIF or video if issue #6 remains desirable.

Acceptance checks:

- New hardware reports include `doctor --json`.
- Pack contributions have a documented validation command.
- README sets expectations for sudo, Apple Silicon support, and unsupported configurations.

### 6. Later Hardening

Goal: add operational polish after the project has a stable install path and diagnostic loop.

Candidates:

- Signed and notarized binary or cask.
- Optional LaunchDaemon installer for accelerometer access.
- Release attestations.
- SBOM generation.
- Menu bar companion only if CLI adoption shows real demand.

## Suggested Order

1. Release automation and Homebrew tap.
2. `doctor` and structured sensor status.
3. Pure tests plus OS-specific CI split.
4. Voice pack validation and authoring docs.
5. Community templates and compatibility reporting.

This order front-loads distribution and diagnostics because they reduce support cost for every later feature.

## Sources

- GoReleaser GitHub Actions: https://goreleaser.com/customization/ci/actions/
- GoReleaser Homebrew casks: https://goreleaser.com/customization/publish/homebrew_casks/
- Homebrew acceptable formulae: https://docs.brew.sh/Acceptable-Formulae
- Go build constraints: https://go.dev/pkg/go/build/?m=old#hdr-Build_Constraints
- GitHub-hosted macOS runners: https://docs.github.com/en/actions/reference/runners/github-hosted-runners
- GitHub README guidance: https://docs.github.com/articles/about-readmes
- taigrr/spank: https://git.taigrr.com/taigrr/spank
- apple-silicon-accelerometer: https://git.taigrr.com/taigrr/apple-silicon-accelerometer
- mactop: https://github.com/metaspartan/mactop
- xbar plugin model: https://github.com/matryer/xbar
