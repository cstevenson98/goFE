# Simple Renderer Interface Plan

## Overview
A simple renderer interface that extends the CanvasManager to handle 2D sprite and background rendering. This will be the foundation for loading and displaying sprites from sprite sheets and background images.

## Core Interface

### Renderer Interface
```go
type Renderer interface {
    // Asset Loading
    LoadSpriteSheet(path string, frameWidth, frameHeight int) (SpriteSheet, error)
    LoadBackground(path string) (Background, error)
    
    // Rendering Methods
    DrawSprite(sprite Sprite, position Vector2) error
    DrawSpriteFrame(spriteSheet SpriteSheet, frameIndex int, position Vector2) error
    DrawBackground(background Background, position Vector2) error
    
    // Batch Rendering (for performance)
    BeginBatch() error
    EndBatch() error
    
    // Canvas Management
    GetCanvasManager() CanvasManager
}
```

## Supporting Types

### Vector2
```go
type Vector2 struct {
    X float64
    Y float64
}
```

### Sprite
```go
type Sprite struct {
    Texture   Texture
    UV        UVRect
    Width     int
    Height    int
}
```

### SpriteSheet
```go
type SpriteSheet struct {
    Texture     Texture
    FrameWidth  int
    FrameHeight int
    FrameCount  int
    Frames      []UVRect
}
```

### Background
```go
type Background struct {
    Texture Texture
    Width   int
    Height  int
}
```

### UVRect
```go
type UVRect struct {
    U float64  // Left (0.0 to 1.0)
    V float64  // Top (0.0 to 1.0)
    W float64  // Width (0.0 to 1.0)
    H float64  // Height (0.0 to 1.0)
}
```

### Texture
```go
type Texture interface {
    Width() int
    Height() int
    ID() string
}
```

## Implementation Strategy

### Phase 1: Basic Types
1. Create Vector2, UVRect, and basic types
2. Define Texture interface
3. Create error types for renderer

### Phase 2: Asset Loading
1. Implement PNG texture loading
2. Create SpriteSheet from loaded texture
3. Implement Background loading
4. Add error handling for failed loads

### Phase 3: Basic Rendering
1. Implement DrawSprite method
2. Implement DrawSpriteFrame method
3. Implement DrawBackground method
4. Test with simple sprites

### Phase 4: Batch Rendering
1. Implement BeginBatch/EndBatch
2. Optimize rendering performance
3. Add sprite sorting

## File Structure
```
renderer/
├── renderer.go        # Renderer interface and implementation
├── types.go          # Vector2, UVRect, etc.
├── texture.go        # Texture loading and management
├── sprite.go         # Sprite and SpriteSheet handling
├── background.go     # Background handling
└── batch.go          # Batch rendering
```

## Example Usage
```go
// Create renderer with canvas manager
renderer := NewRenderer(canvasManager)

// Load a sprite sheet
spriteSheet, err := renderer.LoadSpriteSheet("sprites/player.png", 32, 32)
if err != nil {
    log.Fatal(err)
}

// Load a background
background, err := renderer.LoadBackground("backgrounds/level1.png")
if err != nil {
    log.Fatal(err)
}

// Render background
err = renderer.DrawBackground(background, Vector2{X: 0, Y: 0})
if err != nil {
    log.Fatal(err)
}

// Render sprite frame
err = renderer.DrawSpriteFrame(spriteSheet, 0, Vector2{X: 100, Y: 100})
if err != nil {
    log.Fatal(err)
}
```

## Key Features
- Simple interface for 2D rendering
- Sprite sheet support with frame indexing
- Background image rendering
- Basic batch rendering for performance
- Clean separation from CanvasManager
- Easy to test and mock

## Next Steps
1. Implement basic types
2. Create texture loading system
3. Implement sprite rendering
4. Test with real sprite sheets from the web
5. Add batch rendering optimization
