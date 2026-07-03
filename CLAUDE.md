# CLAUDE.md — AI Assistant Guide for mediamtx-path-viewer

## Project Overview

**mediamtx-path-viewer** is a lightweight Go web dashboard that monitors active stream paths from a [MediaMTX](https://github.com/bluenviron/mediamtx) server via its REST API. It displays real-time stream information with video playback (HLS/WebRTC), viewer counts, and protocol-specific stream links.

## Tech Stack

- **Backend:** Go 1.26 (standard library `net/http` router)
- **Frontend:** Go `html/template`, Bootstrap 5.3.3 (dark theme), HTMX 1.9.2, HLS.js
- **Build:** Docker (Alpine), GoReleaser, npm scripts (dev tooling only)
- **Dependencies:** Minimal — `godotenv` for env loading, `env_logger`/`logrus` for logging

## Project Structure

```
├── main.go                 # Entry point, HTTP routing, env loading
├── main_types.go           # Data structures (HTMLdata, MediaMTX, Path, Session)
├── mediamtx.go             # MediaMTX REST API client and data formatting
├── version/
│   └── version.go          # Version/build info management
├── static/
│   ├── css/                # Bootstrap CSS (minified)
│   ├── js/                 # HTMX (minified), HLS.js
│   ├── html_templates/     # Go HTML templates (index, path_list)
│   └── pictures/           # Favicon
├── dockerfile              # Docker image definition (golang:1.26-alpine)
├── .goreleaser.yml         # Cross-platform release builds
├── .github/workflows/      # CI: release_build.yml (tag-triggered)
├── injectGitVars.sh        # Generates gitinfo.go with git metadata
├── package.json            # npm dev scripts (wgo, air, docker builds)
└── go.mod                  # Go module definition
```

## Key Architecture Decisions

- **Embedded static assets** — All CSS, JS, HTML templates, and images are embedded into the binary via `//go:embed static`, producing a single self-contained executable.
- **HTMX for dynamic content** — Stream list and viewer counts update via HTMX partial page replacement, avoiding a JavaScript framework.
- **Code generation** — `go generate` runs `injectGitVars.sh` to produce `gitinfo.go` (git-ignored) with build metadata. This must run before building.
- **No test suite** — There are currently no `*_test.go` files or automated tests.

## Development Workflow

### Prerequisites

- Go 1.26+
- Node.js/npm (for dev scripts only)
- A running MediaMTX server (for functional testing)

### Running Locally

1. Create a `.env` file with required environment variables (see below)
2. Run code generation + start the app:
   ```sh
   npm run test          # go generate && go run .
   ```
3. Or use watch mode for live reload:
   ```sh
   npm run wgo           # go generate && wgo (requires wgo installed)
   npm run air           # go generate && air (requires air installed)
   ```

### Building

```sh
go generate && go build .                   # Local binary
npm run build_docker                        # Docker image (stable tag)
npm run build_docker_beta                   # Docker image (beta tag)
```

### Important: Always run `go generate` before building

The build requires `gitinfo.go` to be generated. All npm scripts handle this automatically. If building manually, run `go generate` first (or `sh injectGitVars.sh`).

### Releasing

Releases are automated via GitHub Actions. Push a git tag to trigger GoReleaser, which builds binaries for Linux (amd64) and Windows (amd64).

## Environment Variables

**Required:**
| Variable | Description |
|---|---|
| `MEDIAMTX_API_URL` | MediaMTX API server address (forced to `http://`) |
| `MEDIAMTX_WEBRTC_URL` | WebRTC stream base URL |
| `MEDIAMTX_HLS_URL` | HLS stream base URL |

**Optional:**
| Variable | Default | Description |
|---|---|---|
| `MEDIAMTX_API_PORT` | `9997` | MediaMTX API port |
| `MEDIAMTX_USERNAME` | _(empty)_ | Basic auth username |
| `MEDIAMTX_PASSWORD` | _(empty)_ | Basic auth password |
| `MEDIAMTX_RTMP_URL` | _(empty)_ | RTMP stream base URL |
| `MEDIAMTX_RTSP_URL` | _(empty)_ | RTSP stream base URL |
| `MEDIAMTX_PLAYBACK_URL` | _(empty)_ | MediaMTX playback server base URL (enables recording segment playback) |
| `APP_PORT` | `8080` | Web server listen port |
| `APP_PATH` | _(empty)_ | URL base path prefix (e.g., `/monitor`) |

## MediaMTX Control API Reference

**Docs:** https://mediamtx.org/docs/references/control-api
**OpenAPI spec:** https://github.com/bluenviron/mediamtx/blob/main/api/openapi.yaml
**Current version at time of writing:** v1.19.2

> **v1.19.2 notes:** `Path` fields `ready`/`readyTime`/`tracks` and session `bytesReceived`/`bytesSent`/`rtpPackets*` are **deprecated** in favor of `available`/`online`/`availableTime`, `tracks2` (objects with `codec`/`codecProps`), and `inbound*`/`outbound*`. The deprecated aliases still populate, but the code reads the new `Path` fields (see `formatPathData`) and keeps the deprecated ones only as a fallback for older servers.

All endpoints are paginated via `?page=N&itemsPerPage=N` where applicable.

### Currently Used Endpoints
- `GET /v3/paths/list` — All active paths (paginated)
- `GET /v3/paths/get/{name}` — Single path details (source, available/online, tracks2, readers, inboundFramesInError)
- `GET /v3/info` — Server version + startup time (uptime)
- `GET /v3/hlsmuxers/list` — HLS muxers (per-path bytesSent)
- `GET /v3/hlssessions/list` — HLS reader sessions (remoteAddr, user, isCDN, outboundBytes) — per-viewer HLS + geo
- `GET /v3/rtspsessions/list` + `GET /v3/rtspssessions/list` — RTSP + RTSPS (TLS) sessions
- `GET /v3/rtmpconns/list` + `GET /v3/rtmpsconns/list` — RTMP + RTMPS (TLS) connections
- `GET /v3/srtconns/list` — SRT connections (RTT, send/receive rate, link capacity, loss)
- `GET /v3/webrtcsessions/list` — WebRTC sessions (candidates, RTP metrics)
- `GET /v3/moqsessions/list` — Media-over-QUIC sessions (new in v1.19.0)
- `GET /v3/recordings/list` — Recording segments (cross-referenced to mark recording paths)

> **Note:** Kick endpoints (`POST .../kick/{id}`) are deliberately NOT exposed — the dashboard has no login, so proxying kick actions would let anyone with page access disconnect sessions. Revisit once the app has its own authentication.

### Available but Unused Endpoints

| Group | Endpoints | Key Data Available |
|-------|-----------|-------------------|
| **Global Config** | `GET /v3/config/global/get`, `PATCH .../patch` | Full server configuration |
| **Path Config** | `GET/POST/PATCH/DELETE /v3/config/paths/...` | Path configuration CRUD |
| **Path Defaults** | `GET /v3/config/pathdefaults/get`, `PATCH .../patch` | Default path config |
| **RTSP Connections** | `GET /v3/rtspconns/list`, `GET .../get/{id}` | remoteAddr, session, tunnel, bytes (raw conns; sessions are used instead) |
| **Single-item gets** | `GET /v3/{type}/get/{id}` for all session/conn types | Same data as list, one item |
| **Kick** | `POST .../kick/{id}` for all session/conn types | Disconnect a session (see note above) |
| **Recording detail** | `GET /v3/recordings/get/{name}`, `DELETE .../deletesegment` | Per-recording segments, segment deletion |

## Source Code Guide

### main.go
- `main()` — Loads env, parses flags, creates HTTP client, sets up routes, starts server
- `setupRoutes()` — Registers endpoints:
  - `GET /` — Serves the index page (navbar, modal, toast containers)
  - `GET /connect-to-server/` — HTMX endpoint returning grouped stream grid
  - `GET /viewCount/{id}` — HTMX endpoint returning updated viewer count
  - `GET /stream-detail/{id}` — HTMX endpoint returning modal content for stream detail
  - `GET /monitoring-list/` — HTMX endpoint aggregating info + paths + all session/conn types (incl. HLS sessions, MoQ, RTSPS/RTMPS), geo, and per-stream summaries
- `serveStatic()` — Serves embedded CSS, JS, and icon files
- `getEnv()` — Loads `.env` file and reads all environment variables with defaults/validation

### main_types.go
- `HTMLdata` — Template rendering context (includes `Groups`, `Version`)
- `PathGroup` — Group of paths sharing the same PathName prefix
- `MediaMTX` — API response wrapper (paginated)
- `Path` — Stream path with source info, tracks, readers, and generated fields (URLs, display names)
- `Session` — Reader/source session (type + id)

### mediamtx.go
- `getMediamtxPaths()` — Fetches paginated path list from `/v3/paths/list`
- `getMediamtxPath()` — Fetches single path from `/v3/paths/get/{path}`
- `formatPathData()` — Enriches path with stream URLs, PrettyName, PathName, HTML-safe ID; derives `Ready`/`Tracks`/`ReadyTimeStr` from the v1.19.2 `available`/`tracks2`/`availableTime` fields (fallback to deprecated fields)
- `sortPaths()` / `groupPaths()` — Sort/group paths hierarchically by PathName
- `mediamtxAPIRequest()` — Shared GET helper (auth-aware)
- `getMediamtxInfo/HLSMuxers/HLSSessions/RTSPSessions/RTSPSSessions/RTMPConns/RTMPSConns/SRTConns/WebRTCSessions/MoQSessions()` — Monitoring fetchers (list endpoints), each formatting derived display strings
- `collectRemoteIPs()` / `lookupGeoIP()` / `applyGeoData()` — Batch GeoIP enrichment via ip-api.com (includes HLS + MoQ sessions)
- `buildStreamSummaries()` — Per-path viewer/bandwidth aggregation across all protocols (HLS viewers from sessions, MoQ included)

### HTML Templates (static/html_templates/)
- `index.html` — Page shell with navbar, Bootstrap modal/toast containers, footer
- `path_list.html` — Grouped responsive grid of stream cards with video players, badges, protocol links, empty state, and auto-updating viewer counts (5s HTMX polling)
- `stream_detail.html` — Modal content with large video player, metadata, and protocol links

## Code Conventions

- **Naming:** CamelCase for Go identifiers; `UPPER_SNAKE_CASE` for environment variable references
- **Error handling:** `log.Should(err)` for non-critical errors (log and continue); `log.Fatalf()` for fatal startup errors; standard `(value, error)` return tuples for API functions
- **Logging:** Uses `env_logger` (wraps logrus) — log level controlled via `LOG` env var
- **Templates:** Go `html/template` with `{{.Field}}` syntax; templates parsed fresh per request
- **Frontend interactivity:** HTMX attributes (`hx-get`, `hx-trigger`, `hx-swap`) for all dynamic behavior; no custom JavaScript framework
- **Path ID encoding:** Forward slashes in stream paths are replaced with dashes for use as HTML element IDs

## Files to Never Commit

- `.env` — Contains server credentials and URLs
- `gitinfo.go` — Auto-generated by `go generate`; listed in `.gitignore`
- `node_modules/` — Dev dependencies
- `config.json` — Local configuration

## Common Tasks for AI Assistants

- **Adding a new route:** Add handler in `setupRoutes()` in `main.go`, create HTML template in `static/html_templates/`
- **Adding a new environment variable:** Add to `getEnv()` in `main.go`, document in `dockerfile` ENV block, and add to the table above
- **Modifying stream display:** Edit `static/html_templates/path_list.html` (Go template + HTMX)
- **Adding a new static file:** Place in `static/` directory (it's auto-embedded), then add a handler in `serveStatic()` in `main.go`
- **Changing API integration:** Modify functions in `mediamtx.go`; add new fields to `Path` struct in `main_types.go`
