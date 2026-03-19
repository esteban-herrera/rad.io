# rad.io

A minimal terminal radio player. Browse internet radio stations grouped by tag, pick one, and let it run. Designed for narrow quake-style terminal windows.

```
  rad.io
  ────────────────────────────
  ▶ Ambient
  ▼ Jazz
      WBGO Jazz 88.3
    ▶ NTS Radio 1
  ▶ Untagged
  ────────────────────────────
  ▂▄▆▄▂▅▇▅▃▂▄▆▄▂▃▅▇▄▂▄▆▂▄
  ♪ NTS Radio 1
  vol:████░░ 80%
  ♬ Floating Points - LesAlpx
  h:help
```

## Features

- Stations grouped into collapsible tag sections
- Two visualizer modes: spectrum bars and ASCII radio with flying notes (`v`)
- Scrolling marquee for long station names and track metadata
- Hide the list while listening; press `l` to bring it back
- Random station (`r`)
- Volume, mute, and pause controls
- Stations stored as plain JSON at `~/.config/rad.io/stations.json`
- Responsive to terminal width — works comfortably from ~30 columns up

## Requirements

- [mpv](https://mpv.io) — handles all audio playback
- Go 1.22 or newer — only needed to build from source

## Install

### Quick install (macOS / Linux)

```sh
curl -fsSL https://raw.githubusercontent.com/estie/rad.io/main/install.sh | sh
```

The script checks for mpv and Go, installs mpv via Homebrew or your system package manager if missing, then builds and installs `rad.io` to `~/.local/bin`.

### go install

If you already have Go 1.22+ and mpv:

```sh
go install github.com/esteban-herrera/rad.io@latest
```

Make sure `$(go env GOPATH)/bin` is on your `PATH`.

### Build from source

```sh
git clone https://github.com/esteban-herrera/rad.io
cd rad.io
go build -o rad.io .
```

## Usage

Run `rad.io` in any terminal. On first launch the station list is empty — press `a` to add your first station (name + stream URL).

### Keys

| Key | Action |
|-----|--------|
| `↑` / `↓` or `k` / `j` | Navigate |
| `enter` | Expand/collapse section · Play station |
| `l` | Show / hide station list |
| `r` | Play a random station |
| `s` | Stop |
| `space` | Pause / resume |
| `+` / `-` | Volume up / down |
| `m` | Toggle mute |
| `v` | Switch visualizer mode |
| `a` | Add station |
| `d` | Delete selected station |
| `t` | Edit tags for selected station |
| `f` | Cycle tag filter |
| `h` or `?` | Show / hide key hints |
| `q` | Quit |

### Adding stations

Press `a`, enter a name, then paste a direct stream URL (the kind that ends in `.mp3`, `.aac`, `.m3u8`, or a bare Icecast/Shoutcast address). Most public radio stations publish these on their websites or in their M3U playlists.

### Tagging

Press `t` on any station to edit its tags as a comma-separated list (e.g. `jazz, mellow`). Stations with multiple tags appear under each of their sections.

## Station file

Stations are stored in `~/.config/rad.io/stations.json` as a plain JSON array:

```json
[
  {
    "name": "NTS Radio 1",
    "url": "https://stream-relay-geo.ntslive.net/stream",
    "tags": ["nts", "eclectic"]
  }
]
```

You can edit this file directly — changes are picked up on the next launch.
