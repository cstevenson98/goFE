# Sprite Rendering Implementation Plan

## Overview
This document outlines an iterative approach to implementing sprite rendering functionality in our WebGPU-based Go/WASM game engine. We'll build from basic geometric shapes to full sprite sheet rendering, testing each step visually.

## Current State
- ✅ WebGPU triangle rendering working
- ✅ CanvasManager interface with stubs for sprite methods
- ✅ Basic types (Vector2, UVRect, Texture, Pipeline) defined
- ✅ Demo sandbox framework in place

## Implementation Phases

### Phase 1: Basic Rectangle Rendering
**Goal**: Render a simple blue rectangle on a black background

**Steps**:
1. Create a basic vertex buffer for a rectangle (4 vertices)
2. Create a simple shader that renders solid colors
3. Implement `DrawTexture` to draw a blue rectangle instead of using 2D context
4. Test: Should see a blue rectangle on black background

**Files to modify**:
- `canvas.go`: Implement basic rectangle rendering in `DrawTexture`
- Add vertex buffer management
- Add basic shader compilation

**Success criteria**:
- Blue rectangle visible on screen
- No crashes or errors
- Console shows "DEBUG: DrawTexture called" with position/size

### Phase 2: Multiple Rectangles
**Goal**: Render multiple rectangles at different positions

**Steps**:
1. Extend vertex buffer to handle multiple rectangles
2. Add position parameter to rectangle rendering
3. Test with 3-5 rectangles at different positions
4. Each rectangle should be a different color

**Success criteria**:
- Multiple colored rectangles visible
- Each rectangle at correct position
- Smooth rendering without flicker

### Phase 3: Animated Rectangles
**Goal**: Make rectangles move and animate

**Steps**:
1. Add time-based animation to rectangle positions
2. Implement smooth movement (sine wave, linear, etc.)
3. Add color animation (cycling through colors)
4. Test with moving, color-changing rectangles

**Success criteria**:
- Rectangles move smoothly
- Colors change over time
- Animation is smooth (60fps)

### Phase 4: Texture Loading
**Goal**: Load and display actual PNG textures

**Steps**:
1. Implement texture loading from PNG files
2. Create WebGPU texture from loaded image
3. Modify shader to sample from texture
4. Test with a simple PNG image

**Success criteria**:
- PNG image loads successfully
- Image displays correctly on rectangle
- No texture loading errors

### Phase 5: Sprite Sheet Support
**Goal**: Display specific regions of a sprite sheet

**Steps**:
1. Implement UV coordinate mapping
2. Modify shader to handle texture coordinates
3. Add sprite sheet parsing (frame dimensions)
4. Test with llama.png sprite sheet

**Success criteria**:
- Specific frames from sprite sheet display
- UV coordinates work correctly
- Multiple sprites from same sheet

### Phase 6: Animation System
**Goal**: Animate through sprite sheet frames

**Steps**:
1. Add frame timing system
2. Implement frame cycling
3. Add animation speed control
4. Test with animated llama sprites

**Success criteria**:
- Sprites animate through frames
- Animation speed is controllable
- Smooth frame transitions

### Phase 7: Batch Rendering
**Goal**: Render multiple sprites efficiently

**Steps**:
1. Implement vertex batching
2. Add instanced rendering
3. Optimize for many sprites
4. Test with 100+ animated sprites

**Success criteria**:
- Many sprites render smoothly
- Good performance (60fps)
- Memory usage is reasonable

## Technical Implementation Details

### Vertex Buffer Structure
```go
type SpriteVertex struct {
    Position [2]float32  // x, y
    TexCoord [2]float32  // u, v
    Color    [4]float32  // r, g, b, a
}
```

### Shader Requirements
- Vertex shader: Transform positions, pass through texture coordinates
- Fragment shader: Sample texture, apply color tinting
- Uniforms: Projection matrix, texture sampler

### WebGPU Pipeline
- Input: Vertex buffer with position, texcoord, color
- Output: Rendered sprites with texture sampling
- State: Blend mode for transparency

## Testing Strategy

### Visual Tests
- Each phase must show visible results on screen
- Use different colors/positions to verify correctness
- Test edge cases (zero size, off-screen, etc.)

### Performance Tests
- Monitor frame rate during development
- Test with increasing numbers of sprites
- Profile memory usage

### Error Handling
- Graceful fallbacks for missing textures
- Clear error messages for debugging
- Robust handling of invalid parameters

## File Organization

### Core Files
- `canvas.go`: Main rendering implementation
- `types.go`: Data structures and interfaces
- `main.go`: Demo and testing code

### Future Files
- `shaders.go`: Shader compilation and management
- `texture.go`: Texture loading and management
- `sprite.go`: Sprite animation logic

## Success Metrics

### Phase 1-2: Basic Rendering
- ✅ Blue rectangle visible
- ✅ Multiple rectangles render
- ✅ No crashes or errors

### Phase 3: Animation
- ✅ Smooth movement (60fps)
- ✅ Color changes over time
- ✅ Responsive to user input

### Phase 4-5: Textures
- ✅ PNG images load correctly
- ✅ Sprite sheets work
- ✅ UV coordinates accurate

### Phase 6-7: Advanced Features
- ✅ Sprite animation smooth
- ✅ Batch rendering efficient
- ✅ Performance acceptable

## Next Steps

1. **Start with Phase 1**: Implement basic rectangle rendering
2. **Test thoroughly**: Each phase must work before moving to next
3. **Iterate quickly**: Small, testable changes
4. **Document issues**: Keep track of problems and solutions

## Notes

- Keep existing WebGPU triangle working throughout
- Use demo sandbox to test each phase
- Maintain clean, readable code
- Add comprehensive error handling
- Test on different browsers/devices
