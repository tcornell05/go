# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

An isometric world decoration and social game built with Ebitengine in Go, inspired by Habbo Hotel. Players can walk around in 500x500 world tiles, each with customizable walkable areas and furniture placement. The game features a comprehensive world tile editor for room design and boundary editing.

**Art Style:** 64x64 pixel art with isometric perspective using real asset tiles
**Architecture:** World Tile system - each tile (A1, B2, etc.) is a 500x500 area with 11x11 default walkable space
**Platform:** Desktop (with WebAssembly support planned)
**Screen Resolution:** 1024x768 pixels

## Development Commands

### Quick Commands (Use these during development)

```bash
# Run the game immediately
make run

# Start development server with hot reload (automatic rebuilds)
make dev

# Quick compilation check (no build)
make check

# Watch mode (rebuilds on changes)
make watch

# Run all tests
make test

# Debug information and status
./scripts/debug.sh
```

### Build Commands

```bash
# Build desktop executable
make build

# Build for WebAssembly
make wasm

# Clean build artifacts
make clean
```

### Validation Commands

```bash
# Comprehensive code validation (formatting, vet, security)
make validate

# Validate all game assets
make assets

# Run full test suite with coverage
./scripts/test.sh
```

### Database Commands

```bash
# Reset database (deletes data/game.db)
make db-reset

# Update dependencies
make deps
```

### Advanced Development

```bash
# Load development environment variables
source .dev.env

# All available commands
make help
```

## Architecture

### Package Structure

- `main.go` - Entry point, embeds assets, initializes game
- `internal/game/` - Main game loop and state management  
- `internal/entity/` - Player entity with movement and animation
- `internal/world/` - **World Tile system** - 500x500 tiles with walkability and object placement
- `internal/room/` - Legacy room structures (still used for editor compatibility)
- `internal/render/` - Isometric rendering system with F2 boundary editing visuals
- `internal/input/` - Input handling for player, editor, and boundary editing
- `internal/editor/` - World tile editor UI and inventory system  
- `internal/database/` - SQLite persistence for world tiles, objects, and assets
- `pkg/assets/` - Asset loading and management
- `pkg/math/` - Isometric coordinate transformations

### Core Systems

#### Isometric Rendering

- Coordinate transform: `ix = (x - y) * 32; iy = (x + y) * 16`
- Tile size: 64x64 pixels
- Z-ordering for proper depth (floor → walls → furniture → player)
- Camera fixed at room center with mathematical positioning

#### Movement System

- **Grid Mode** (default off): Tile-by-tile movement with perfect alignment
- **Continuous Mode** (default on): Smooth movement with 0.05 increments
- Player bounds: Grid coordinates (1,1) to (11,11)
- 8-directional movement with diagonal support

#### World Tile System

- **500x500 World Tiles**: Each tile ID (A1, B2, etc.) represents a large area  
- **11x11 Default Grid**: Central walkable area, expandable via boundary editing
- **F2 Boundary Editor**: Toggle grid overlay, left-click to allow, right-click to block
- **Visual Feedback**: Grey borders + yellow + for walkable, blue borders for blocked
- **Auto-save as WorldTile**: Tile IDs automatically save in WorldTile JSON format

#### Editor Mode

- **Editor Mode** (E key): Toggle world tile decoration mode
- **Asset Placement**: Click to place, middle-click to inspect objects
- **Asset Properties**: Position offsets, rotation states, depth adjustment
- **Boundary Editing**: F2 key + mouse clicks to modify walkable areas
- **Persistence**: WorldTile JSON files for tiles, SQLite for objects and assets
- **Inventory System**: Categorized tiles (floors, walls, furniture, etc.)

### Key Controls

- **Movement**: WASD/Arrow keys + QEZX for diagonals
- **Editor**: E (toggle), Click (place), Middle-click (inspect), Delete (remove)
- **Boundary Editing**: F2 (toggle grid), Left-click (allow), Right-click (block)
- **Inspection**: I key or Middle-click, Inspect button for magnifying glass cursor
- **Camera**: I/K (zoom), J/L (rotate - planned)
- **Mode Toggle**: Backtick (`) switches movement modes
- **Rotation**: R key (rotates player or selected asset)

## Database Schema

The game uses SQLite with three main tables:

- `rooms`: Stores room metadata and layouts
- `objects`: Placed furniture and tiles with positions
- `entity_properties`: Custom asset definitions with multiple rotation states

## Asset System

Assets are embedded at compile time using `//go:embed assets/*`. The system supports:

- Multiple rotation states per asset (stored in database)
- Dynamic asset loading from filesystem
- Categorized inventory (Floors, Walls, Furniture, etc.)
- Real-time preview in editor mode

## Constants

- Room size: 15x15 tiles (configurable)
- Default player spawn: (5, 5)
- Tile dimensions: 64x64 pixels
- Screen: 1024x768 pixels
- Database path: `data/game.db`

## Test-Driven Development Philosophy

### Testing Strategy (CRITICAL FOR CLAUDE CODE)

**ALL development must follow this testing hierarchy:**

1. **Smoke Test 1: Unit & Integration Tests** - The primary driver of development
   - Write tests FIRST, before implementing features
   - Every new function/system needs unit tests in `/test/`
   - Integration tests verify component interactions
   - Tests must pass before moving to smoke test 2

2. **Smoke Test 2: Game Client Runtime** - Secondary validation
   - Run the actual game client (`go run . editor --tile A1`)
   - Manual verification of features working end-to-end
   - Only after tests prove the code works

### Testing Commands (Use These First!)

```bash
# ALWAYS RUN THESE BEFORE IMPLEMENTING
make test                    # Unit tests
go test ./test/             # Integration tests
go test -v ./test/          # Verbose test output

# After tests pass, then run game
go run . editor --tile A1   # Test editor functionality
go run . play --start A1    # Test gameplay
```

### Test Structure

- `/test/` - All test files
- `/test/world_tile_test.go` - World tile system tests
- `/test/integration_test.go` - Cross-component tests
- Test files end with `_test.go`
- Test functions start with `Test`

### Development Workflow (MANDATORY)

```bash
# 1. Write tests first (TDD)
# 2. Run tests to see failures
make test

# 3. Implement minimum code to pass tests
# 4. Re-run tests until they pass
make test

# 5. Only then test with game client
go run . editor --tile A1

# 6. Refactor if needed, keeping tests passing
```

## Development Tooling

### Iterative Development Workflow

This project is designed for rapid development through Claude Code interface with comprehensive tooling:

1. **Test-First Development**: Write unit tests before implementation
2. **Hot Reload Development**: `make dev` - Automatic rebuilds on file changes
3. **Quick Error Detection**: `make check` - Instant compilation feedback
4. **Continuous Testing**: `make test` - Full test suite with coverage
5. **Asset Validation**: `make assets` - Validates all 601+ PNG files
6. **Code Quality**: `make validate` - Formatting, linting, security checks

### Development Scripts

- `scripts/dev.sh` - Hot reload server with automatic rebuilds/restarts
- `scripts/watch.sh` - File monitoring with build error reporting
- `scripts/test.sh` - Comprehensive test runner with coverage reports
- `scripts/validate.sh` - Code quality validation (formatting, vet, security)
- `scripts/check-assets.sh` - Asset validation and inventory reporting
- `scripts/debug.sh` - System status, build info, and debugging helper

### Development Environment

- `.dev.env` - Environment variables for development mode
- `Makefile` - Centralized command interface with help system
- **Database**: SQLite with 6 tables (entities, rooms, world_tiles, etc.)
- **Assets**: 601 PNG files (47MB total) with automated validation
- **System**: 32-core development environment with 31GB RAM

### Recommended Development Cycle

```bash
# Start development session
make dev              # Hot reload server

# In separate terminal
./scripts/debug.sh    # Check system status
make test            # Run tests
make validate        # Check code quality
```

### Asset Management

- **Location**: `assets/` directory with organized subdirectories
- **Types**: Character sprites, floor/wall tiles, furniture, UI elements
- **Validation**: Automated checking of image files and inventory
- **Integration**: Embedded at compile time with runtime loading support

## Extending the Game

When adding features:

- Game state belongs in the `Game` struct
- Input handling in `Controller.HandleInput()`
- Rendering in `Renderer.Draw()` with proper z-ordering
- New assets: Add to `/assets` and register in `TileInventory`
- Database changes: Update schema in `database.NewDatabase()`
- Maintain isometric perspective using `math.CartesianToIso()`
- **Testing**: Use `make dev` for immediate feedback on changes, any new feature gets added to test suite
- **Validation**: Run `make check` before committing changes
- references to "room": it used to be it's own object/model (db table and all) it was deprecated and replaced with world tiles

