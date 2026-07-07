# Video2IPTVnDLNA

Serve videos from a local folder as **IPTV channels** (M3U playlist) and as a **DLNA/UPnP media server** for smart TVs and media players — all from a single lightweight Go binary on one HTTP port.

## Features

- **IPTV (M3U playlist)** — each video file becomes a channel for VLC, TiviMate, and other IPTV players
- **DLNA/UPnP media server** — auto-discovered on your local network by smart TVs, consoles, and DLNA apps
- **Shared video library** — one folder, one HTTP server, two protocols
- **Configurable folder** — point at any directory with `-folder`
- **Recursive scan** — optionally include videos in subdirectories with `-r`
- **Configurable port** — default `8080`, change with `-port`
- **Custom DLNA name** — set the name shown on TVs with `-name`
- **Cross-platform** — pre-built binaries for Windows, Linux (amd64, 386, arm64, armv7)

### Supported video formats

`.mp4`, `.mkv`, `.avi`, `.mov`, `.webm`, `.m4v`, `.ts`, `.flv`, `.wmv`

## Architecture

```mermaid
flowchart LR
  subgraph shared [Shared]
    Folder[Video folder]
    Scanner[listVideos]
    HTTP[/videos/ HTTP server]
  end

  subgraph iptv [IPTV]
    M3U[/playlist.m3u]
  end

  subgraph dlna [DLNA]
    SSDP[SSDP discovery]
    UPnP[UPnP SOAP + DIDL-Lite]
  end

  Folder --> Scanner
  Scanner --> M3U
  Scanner --> UPnP
  M3U --> HTTP
  UPnP --> HTTP
```

## Quick start

### Download a release

Grab the latest binary for your platform from the [Releases](https://github.com/lord-aali/Video2IPTVnDLNA/releases) page.

### Windows

```powershell
.\video2iptvndlna.exe -folder "C:\Users\You\Videos" -port 8080
```

### Linux

```bash
chmod +x video2iptvndlna-linux-amd64
./video2iptvndlna-linux-amd64 -folder /media/videos -port 8080
```

### Build from source

Requirements: [Go 1.21+](https://go.dev/dl/)

```bash
git clone https://github.com/lord-aali/Video2IPTVnDLNA.git
cd Video2IPTVnDLNA
go build -o video2iptvndlna .
```

## Usage

```text
video2iptvndlna -folder <path> [-port <port>] [-r] [-name <dlna-name>]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-folder` | `.` | Path to the folder containing video files |
| `-port` | `8080` | HTTP listening port |
| `-r` | `false` | Search for videos recursively in subdirectories (IPTV playlist only) |
| `-name` | `Video2IPTVnDLNA` | DLNA server name shown to TVs and media players on the network |

### Examples

Serve the current directory on port 8080:

```bash
video2iptvndlna
```

Serve a specific folder with recursive IPTV listing:

```bash
video2iptvndlna -folder /home/user/Movies -r
```

Custom port and DLNA name:

```bash
video2iptvndlna -folder "D:\Videos" -port 9090 -name "My Home Videos"
```

## Endpoints

| Endpoint | Protocol | Description |
|----------|----------|-------------|
| `http://<host>:<port>/playlist.m3u` | IPTV | M3U playlist for IPTV players |
| `http://<host>:<port>/videos/<file>` | IPTV | Direct HTTP video stream |
| `http://<host>:<port>/rootDesc.xml` | DLNA | UPnP device description |
| SSDP multicast `239.255.255.250:1900` | DLNA | Network discovery |

## Client setup

### IPTV (VLC, TiviMate, etc.)

1. Start the server.
2. Open the playlist URL in your player:

   ```text
   http://<server-ip>:8080/playlist.m3u
   ```

   In VLC: **Media → Open Network Stream** and paste the URL.

### DLNA (Smart TV, BubbleUPnP, etc.)

1. Start the server on the same LAN as your TV or player.
2. Open the media source / DLNA / network browser on your device.
3. Select **Video2IPTVnDLNA** (or the name you set with `-name`).
4. Browse folders and play videos.

## Notes

- **IPTV vs DLNA browsing** — the `-r` flag only affects the flat M3U IPTV playlist. DLNA always exposes the full folder tree for browsing.
- **Firewall** — allow inbound TCP on your chosen port and UDP port `1900` for DLNA discovery.
- **Special characters** — filenames with spaces and `&` are handled correctly in playlist URLs.
- **No transcoding** — videos are served as-is. Use widely supported formats (e.g. H.264 MP4) for best DLNA compatibility.

## License

BSD-3-Clause (via [anacrolix/dms](https://github.com/anacrolix/dms) dependency)
