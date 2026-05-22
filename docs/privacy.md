# Privacy

Moody runs locally. It reads local hardware and system state so it can react to
events such as power changes, USB changes, WiFi state, display count, audio
route, lid state, accelerometer events, and local Kiro hook files.

Moody does not send telemetry or hardware data to a server.

Voice packs are local files under `~/.moody/packs`. Community pack installs use
`git clone` for the URL you provide, validate the cloned pack, and then install
it locally.

Diagnostic output from `moody doctor --json` can include local state such as OS,
architecture, root status, sensor availability, battery percentage, display
count, and audio output name. Review diagnostic output before posting it in a
public issue.
