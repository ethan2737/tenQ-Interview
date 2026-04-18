# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

TenQ Interview is a desktop review workstation application built with Wails (Go + Vanilla JavaScript). It compresses long-form interview preparation materials into concise flashcards suitable for oral answers in real interviews.

## Build & Development Commands

```bash
# Run tests
go test ./...

# Build application
wails build

# Development mode
wails dev
```

## Architecture

### Technology Stack
- **Backend**: Go 1.25.0 with Wails v2.12.0
- **Frontend**: Vanilla JavaScript (no framework), native CSS
- **Key Dependencies**: `golang.org/x/text` (encoding), `github.com/samber/lo` (utilities)

### Project Structure

```
.
├── main.go              # Wails entry point
├── app.go               # App logic exposed to frontend
├── wails.json           # Wails configuration
├── DESIGN.md            # Design system (REQUIRED reading for UI changes)
├── TODOS.md             # Feature roadmap
├── frontend/
│   ├── src/
│   │   ├── app.js              # Frontend state management
│   │   ├── style.css           # Styles following DESIGN.md
│   │   ├── index.html          # HTML template
│   │   ├── markdown-render.js  # Custom Markdown renderer
│   │   └── import-session.js   # Import session management
│   └── wailsjs/         # Auto-generated Go-JS bridge
└── internal/
    ├── cache/           # File-based cache with versioning
    ├── card/            # Flashcard generation (250 char limit)
    ├── importer/        # Encoding detection (UTF-8/GB18030)
    ├── library/         # Markdown file scanning
    ├── parser/          # Markdown parsing
    ├── pipeline/        # Processing pipeline orchestration
    ├── segment/         # Paragraph selection by keyword scoring
    └── workbench/       # Service layer for frontend API
```

### Data Flow

```
File Import → Encoding Normalization → Markdown Parse → Segment Selection → Card Generation → Cache
    ↓              ↓                        ↓              ↓                  ↓              ↓
.md/.txt      UTF-8/GB18030           Title/Body     Top 3 paragraphs   Q+A + Source   index.json
```

### Key Design Decisions

1. **Encoding Auto-Detection**: Supports UTF-8 and GB18030 for Chinese text
2. **Versioned Cache**: Cache keys include rule versions to avoid stale data after upgrades
3. **Zero Framework Frontend**: Native JS for instant startup, no build step
4. **Left-Tree Right-Read Layout**: Mobile collapses to drawer sidebar

## Important Conventions

- **Design System**: Always read `DESIGN.md` before making any visual/UI changes
- **No Frontend Build**: Frontend is served directly from `frontend/` directory
- **Go-JS Bridge**: Frontend calls Go methods via `window.goApp` (auto-generated)
- **Cache Location**: System cache dir or `.cache/tenq-interview/index.json`

## Testing

All `internal/*` packages have corresponding `_test.go` files. Run with:
```bash
go test ./...
```

## Feature Roadmap (TODOS.md)

1. PDF/docs import with multi-card splitting
2. Topic/random review modes
3. Mock interview podcast (TTS) generation
4. Document search capability
