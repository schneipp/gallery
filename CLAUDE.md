# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

- **Run the application**: `go run main.go`
- **Build the application**: `go build -o gallery`
- **Run tests**: `go test ./...`
- **Run a specific test**: `go test -v -run TestName`

## Architecture & Structure

The project is a self-contained Go web application that serves as a photo gallery with integration for Capture One and YouTube audio.

### Core Architecture
- **Single Binary**: The application is primarily contained within `main.go`, handling routing, logic, and data management.
- **Data Persistence**: Uses a local JSON file (`gallery_data.json`) as a flat-file database. A `sync.RWMutex` is used to manage concurrent access to the shared `appData` state.
- **Templates**: HTML templates are embedded as strings in separate Go files (e.g., `tmpl_gallery.go`, `tmpl_admin_edit.go`) to keep the main logic cleaner.
- **Media Management**: 
    - Images are proxied through `/proxy/image` from Capture One's cloud.
    - Audio is downloaded from YouTube using `yt-dlp` and stored in a local `media/` directory.

### Key Components
- **Capture One Integration**: Syncs galleries by establishing a session with the Capture One API and fetching photo variants.
- **Admin Panel**: A protected area (`/admin`) for managing site configuration, galleries, and audio media.
- **Auth System**: Simple session-based authentication using a cookie (`gallery_session`) and SHA-256 password hashing.
- **Public Gallery**: Supports both slug-based public access (`/slug`) and secret token-based private access (`/s/{token}`).

### External Dependencies
- `yt-dlp`: Required for downloading audio from YouTube.
- `ffprobe`: Used as a fallback for determining audio duration.
