# Voice Pack Authoring

Community packs are directories with a `manifest.json` at the root. They can
provide text lines in the manifest, audio files under `audio/<event>/`, or both.

## Validate A Pack

```bash
moody pack validate ./my-pack
moody pack validate --json ./my-pack
```

Validation checks required manifest metadata, safe language/personality slugs,
known event names, known mood names, supported audio file extensions, and
whether the pack contains at least one usable line or audio file.

## Manifest

```json
{
  "name": "English Pirate",
  "language": "en",
  "personality": "pirate",
  "version": "1.0.0",
  "author": "you",
  "nsfw": false,
  "description": "Pirate speak pack",
  "minMoodyVersion": "1.4.0",
  "license": "MIT",
  "homepage": "https://example.com/moody-pirate",
  "supportedEvents": ["slap"],
  "audioFormats": [".mp3"],
  "lines": {
    "slap": {
      "happy": ["Easy there, matey."]
    }
  }
}
```

The installed pack name is `<language>_<personality>`, so the example installs
as `en_pirate`.

Supported audio formats are `.mp3`, `.wav`, `.mp4`, `.m4a`, and `.aiff`.
