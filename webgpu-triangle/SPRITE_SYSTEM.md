# Sprite System Implementation

This document describes the n x m sprite sheet system with animation support.

## Overview

The sprite system is built in 5 stages, all now complete and ready for testing:

## Stage 1: Sprite Interface and Types ✓

**File:** `internal/types/sprite.go`

Created the core `Sprite` interface that all sprite types must implement:

```go
type Sprite interface {
    GetSpriteRenderData() SpriteRenderData  // Returns render data for pipeline
    Update(deltaTime float64)                // Updates animation state
    SetPosition(pos Vector2)                 // Sets sprite position
    GetPosition() Vector2                    // Gets sprite position
    SetVisible(visible bool)                 // Shows/hides sprite
    IsVisible() bool                         // Checks visibility
}
```

**SpriteRenderData** structure contains everything needed to render:
- `TexturePath` - Path to the sprite sheet texture
- `Position` - World position
- `Size` - Display size
- `UV` - UV coordinates for current frame
- `Visible` - Whether to render

## Stage 2: Sprite Sheet with n x m Grid ✓

**File:** `internal/sprite/sprite.go`

Implemented `SpriteSheet` struct supporting any n x m grid:

```go
type SpriteSheet struct {
    texturePath string
    position    Vector2
    size        Vector2
    columns     int  // n columns
    rows        int  // m rows
    // ... animation fields
}
```

**Key Features:**
- Constructor: `NewSpriteSheet(texturePath, position, size, columns, rows)`
- Automatically calculates total frames = columns × rows
- Calculates UV coordinates for current frame based on grid position
- Frame layout: left-to-right, top-to-bottom

**Example:**
```go
// Create a 2x2 sprite sheet (4 frames total)
sprite := sprite.NewSpriteSheet(
    "llama.png",
    sprite.Vector2{X: 100, Y: 100},  // Position
    sprite.Vector2{X: 256, Y: 256},  // Size
    2,  // 2 columns
    2,  // 2 rows
)
```

## Stage 3: Frame Animation ✓

**Animation System:**

The sprite sheet automatically animates through frames:

- `frameTime` - Time per frame (default 0.1s = 10 FPS)
- `currentFrame` - Current frame index (0-based)
- `elapsed` - Time elapsed in current frame

**Update Logic:**
```
elapsed += deltaTime
if elapsed >= frameTime:
    currentFrame = (currentFrame + 1) % totalFrames
    elapsed -= frameTime
```

**Control Methods:**
- `SetFrameTime(seconds)` - Set animation speed
- `SetCurrentFrame(index)` - Jump to specific frame
- `GetCurrentFrame()` - Get current frame index
- `GetTotalFrames()` - Get total number of frames

**UV Calculation:**
```go
frameWidth := 1.0 / float64(columns)
frameHeight := 1.0 / float64(rows)

frameX := currentFrame % columns
frameY := currentFrame / columns

UV = {
    U: frameX * frameWidth,
    V: frameY * frameHeight,
    W: frameWidth,
    H: frameHeight,
}
```

## Stage 4: Game State Sprite Management ✓

**File:** `internal/engine/engine.go`

Engine now manages sprites per game state:

```go
type Engine struct {
    gameStatePipelines map[types.GameState][]types.PipelineType
    gameStateSprites   map[types.GameState][]types.Sprite  // NEW!
    stateLock          sync.Mutex
}
```

**Initialization:**
```go
func (e *Engine) initializeGameStates() {
    // SPRITE state - textured pipeline + sprite array
    e.gameStatePipelines[types.SPRITE] = []types.PipelineType{
        types.TexturedPipeline,
    }
    
    llamaSprite := sprite.NewSpriteSheet(
        "llama.png",
        sprite.Vector2{X: 100, Y: 100},
        sprite.Vector2{X: 256, Y: 256},
        2, 2,  // 2x2 grid = 4 frames
    )
    llamaSprite.SetFrameTime(0.5)  // 2 FPS for testing
    
    e.gameStateSprites[types.SPRITE] = []types.Sprite{llamaSprite}
    
    // TRIANGLE state - triangle pipeline + no sprites
    e.gameStatePipelines[types.TRIANGLE] = []types.PipelineType{
        types.TrianglePipeline,
    }
    e.gameStateSprites[types.TRIANGLE] = []types.Sprite{}
}
```

## Stage 5: Render Integration ✓

**Update Loop:**
```go
func (e *Engine) Update(deltaTime float64) {
    // Get current state's sprites (thread-safe)
    e.stateLock.Lock()
    sprites := e.gameStateSprites[e.currentGameState]
    e.stateLock.Unlock()
    
    // Update each sprite's animation
    for _, sprite := range sprites {
        sprite.Update(deltaTime)
    }
    
    // Load textures if needed
    e.loadSpriteTextures()
}
```

**Render Loop:**
```go
func (e *Engine) Render() {
    // Get current state's sprites
    e.stateLock.Lock()
    sprites := e.gameStateSprites[e.currentGameState]
    e.stateLock.Unlock()
    
    // Render each visible sprite
    for _, sprite := range sprites {
        renderData := sprite.GetSpriteRenderData()
        
        if !renderData.Visible {
            continue
        }
        
        e.canvasManager.DrawTexturedRect(
            renderData.TexturePath,
            renderData.Position,
            renderData.Size,
            renderData.UV,  // UV changes per frame!
        )
    }
    
    e.canvasManager.Render()
}
```

## Testing Instructions

### Test 1: Single Frame (1x1 grid)
```go
sprite := sprite.NewSpriteSheet("llama.png", pos, size, 1, 1)
// Should display static image (no animation)
```

### Test 2: Horizontal Animation (4x1 grid)
```go
sprite := sprite.NewSpriteSheet("sprite_sheet.png", pos, size, 4, 1)
sprite.SetFrameTime(0.2)  // 5 FPS
// Should animate left-to-right across 4 frames
```

### Test 3: Grid Animation (2x2 grid)
```go
sprite := sprite.NewSpriteSheet("sprite_sheet.png", pos, size, 2, 2)
sprite.SetFrameTime(0.5)  // 2 FPS
// Should animate: top-left → top-right → bottom-left → bottom-right
```

### Test 4: State Switching
- Press `1` - Switch to SPRITE state (shows animated sprite)
- Press `2` - Switch to TRIANGLE state (shows red triangle)

### Test 5: Multiple Sprites
```go
e.gameStateSprites[types.SPRITE] = []types.Sprite{
    sprite1,  // Llama at (100, 100)
    sprite2,  // Another sprite at (400, 100)
}
// Both should animate independently
```

## Current Configuration

**Default setup (in engine):**
- Sprite sheet: `llama.png` (2x2 grid = 4 frames)
- Position: (100, 100)
- Size: 256x256
- Frame time: 0.5 seconds (2 FPS for easy observation)
- State: SPRITE (press '1' to activate)

## How to Modify

### Change grid size:
```go
llamaSprite := sprite.NewSpriteSheet(
    "llama.png",
    sprite.Vector2{X: 100, Y: 100},
    sprite.Vector2{X: 256, Y: 256},
    4,  // n columns
    3,  // m rows (total 12 frames)
)
```

### Change animation speed:
```go
llamaSprite.SetFrameTime(0.1)  // 10 FPS (faster)
llamaSprite.SetFrameTime(1.0)  // 1 FPS (slower)
```

### Add more sprites:
```go
sprite2 := sprite.NewSpriteSheet(...)
sprite3 := sprite.NewSpriteSheet(...)

e.gameStateSprites[types.SPRITE] = []types.Sprite{
    llamaSprite,
    sprite2,
    sprite3,
}
```

### Control frames manually:
```go
sprite.SetCurrentFrame(0)  // Jump to first frame
sprite.SetCurrentFrame(3)  // Jump to fourth frame
```

## Architecture

```
Engine
├── gameStatePipelines[SPRITE] = [TexturedPipeline]
├── gameStateSprites[SPRITE] = [sprite1, sprite2, ...]
│
├── Update(deltaTime)
│   ├── For each sprite in current state
│   │   └── sprite.Update(deltaTime)  // Advances animation
│   └── loadSpriteTextures()
│
└── Render()
    ├── For each sprite in current state
    │   ├── renderData = sprite.GetSpriteRenderData()
    │   └── DrawTexturedRect(renderData.UV)  // UV changes per frame
    └── canvasManager.Render()
```

## Thread Safety

All sprite access is protected by `stateLock` mutex:
- State changes lock before modifying
- Update/Render lock briefly to get sprite array
- No data races when switching states

