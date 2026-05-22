# Releasing Moody

Moody releases are tag-driven through GoReleaser.

## One-Time Setup

Add a repository secret named `TAP_GITHUB_TOKEN` in `dinakars777/moody`.
The token needs contents write access to `dinakars777/homebrew-tap` so
GoReleaser can publish `Formula/moody.rb`.

## Release

```bash
git checkout main
git pull --ff-only origin main
git tag vX.Y.Z
git push origin vX.Y.Z
```

The release workflow builds a `darwin_arm64` archive, uploads
`checksums.txt`, and updates the Homebrew formula.

## Local Snapshot Check

```bash
go run github.com/goreleaser/goreleaser/v2@latest release --snapshot --clean
```

Snapshot builds write artifacts to `dist/` and do not publish anything.
