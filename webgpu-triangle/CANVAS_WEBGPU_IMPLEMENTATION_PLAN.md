# Canvas WebGPU Implementation Plan
**Project**: WebGPU Go WASM Sprite Rendering  
**Goal**: Build out canvas rendering logic starting with basic rectangles, then progressing to sprite display

---

## Current State Analysis

### ✅ What's Already Implemented
1. **Working WebGPU Triangle Rendering**
   - Basic WebGPU initialization in `canvas.go`
   - Simple triangle shader and pipeline
   - Render loop using `requestAnimationFrame`
   - WebGL fallback for unsupported browsers

2. **Type System**
   - `Vector2` - 2D position/size representation
   - `UVRect` - Texture coordinate rectangles
   - `Texture` interface with `WebGPUTexture` implementation
   - `Pipeline` interface with `WebGPUPipeline` implementation
   - `SpriteVertex` - Vertex structure for sprite rendering

3. **Canvas Interface**
   - `CanvasManager` interface defined
   - Stub methods for `DrawTexture`, `DrawTextureRotated`, `DrawTextureScaled`
   - Stub methods for batch rendering (`BeginBatch`, `EndBatch`, `FlushBatch`)
   - Stub methods for pipeline access (`GetSpritePipeline`, `GetBackgroundPipeline`)

4. **Sprite System (Not Yet Integrated)**
   - `SpriteRenderer` for managing sprites
   - Sprite sheet loading using JavaScript Image API
   - Animation system with frame tracking
   - UV coordinate calculation for sprite frames

### ❌ What's Missing
1. **WebGPU Sprite Pipeline** - No separate pipeline for textured rendering
2. **Sprite Shaders** - No WGSL shaders for texture sampling
3. **Vertex Buffer Management** - No dynamic vertex buffer for sprites
4. **Texture Upload to GPU** - No mechanism to upload textures to WebGPU
5. **DrawTexture Implementation** - Currently just a stub with println
6. **Batch Rendering** - No vertex batching implementation

---

## Implementation Phases

### Phase 1: Basic Colored Rectangle Rendering
**Goal**: Render a solid-colored rectangle using WebGPU (no textures yet)

#### Tasks:
1. **Create Sprite Vertex Shader (Colored)**
   ```wgsl
   struct VertexOutput {
       @builtin(position) position: vec4f,
       @location(0) color: vec4f,
   }
   
   @vertex
   fn vs_main(
       @location(0) position: vec2f,
       @location(1) color: vec4f
   ) -> VertexOutput {
       var output: VertexOutput;
       output.position = vec4f(position, 0.0, 1.0);
       output.color = color;
       return output;
   }
   ```

2. **Create Sprite Fragment Shader (Colored)**
   ```wgsl
   @fragment
   fn fs_main(@location(0) color: vec4f) -> @location(0) vec4f {
       return color;
   }
   ```

3. **Create Sprite Pipeline in `setupWebGPUTriangle`**
   - Add sprite pipeline creation after triangle pipeline
   - Configure vertex buffer layout (position + color)
   - Enable alpha blending for transparency

4. **Implement Dynamic Vertex Buffer**
   - Create a vertex buffer that can be updated each frame
   - Structure: `[x, y, r, g, b, a]` per vertex
   - 6 vertices per rectangle (2 triangles)

5. **Implement `DrawTexture` for Colored Rectangles**
   - Convert position/size to normalized device coordinates (-1 to 1)
   - Generate 6 vertices for a quad (2 triangles)
   - Upload vertices to GPU
   - Draw using sprite pipeline

6. **Update Main Loop**
   - Integrate sprite rendering into existing render loop
   - Keep triangle rendering separate

#### Success Criteria:
- ✅ Blue rectangle renders at position (100, 100) with size 64x64
- ✅ No crashes or WebGPU errors
- ✅ Console shows "Drawing rectangle" debug messages
- ✅ Original triangle still renders

#### Files to Modify:
- `internal/canvas/canvas.go` - Implement DrawTexture, create sprite pipeline
- `cmd/webgpu-triangle/main.go` - Test with colored rectangle

---

### Phase 2: Multiple Colored Rectangles
**Goal**: Render multiple rectangles with different colors and positions

#### Tasks:
1. **Implement Vertex Batching**
   - Create a slice to accumulate vertices
   - Add `BeginBatch()` - clear vertex slice
   - Update `DrawTexture()` - append vertices to slice
   - Add `FlushBatch()` - upload all vertices and draw
   - Add `EndBatch()` - call FlushBatch

2. **Update Render Loop**
   - Call BeginBatch at start of frame
   - Draw multiple rectangles
   - Call EndBatch at end of frame

3. **Test with Multiple Rectangles**
   - Draw 3-5 rectangles at different positions
   - Use different colors for each

#### Success Criteria:
- ✅ Multiple colored rectangles render simultaneously
- ✅ Each at correct position
- ✅ Batch rendering is more efficient than individual draws
- ✅ No flickering or visual artifacts

#### Files to Modify:
- `internal/canvas/canvas.go` - Implement batching
- `cmd/webgpu-triangle/main.go` - Test with multiple rectangles

---

### Phase 3: Texture Sampling Support
**Goal**: Load and display actual PNG textures on rectangles

#### Tasks:
1. **Create Textured Sprite Shader (Vertex)**
   ```wgsl
   struct VertexOutput {
       @builtin(position) position: vec4f,
       @location(0) uv: vec2f,
   }
   
   @vertex
   fn vs_main(
       @location(0) position: vec2f,
       @location(1) uv: vec2f
   ) -> VertexOutput {
       var output: VertexOutput;
       output.position = vec4f(position, 0.0, 1.0);
       output.uv = uv;
       return output;
   }
   ```

2. **Create Textured Sprite Shader (Fragment)**
   ```wgsl
   @group(0) @binding(0) var textureSampler: sampler;
   @group(0) @binding(1) var textureData: texture_2d<f32>;
   
   @fragment
   fn fs_main(@location(0) uv: vec2f) -> @location(0) vec4f {
       return textureSample(textureData, textureSampler, uv);
   }
   ```

3. **Implement Texture Upload to GPU**
   - Create method `uploadTextureToGPU(imageData js.Value) js.Value`
   - Use `device.createTexture()` with image dimensions
   - Use `queue.writeTexture()` to upload pixel data
   - Create texture view for binding

4. **Implement Sampler Creation**
   - Create a sampler with linear filtering
   - Configure wrap mode (clamp/repeat)

5. **Create Bind Group for Texture**
   - Create bind group layout (sampler + texture)
   - Create bind group with actual texture and sampler
   - Update pipeline layout

6. **Update Vertex Buffer Layout**
   - Change from `[x, y, r, g, b, a]` to `[x, y, u, v]`
   - Position + UV coordinates

7. **Update `DrawTexture` Implementation**
   - Check if texture has valid GPU texture
   - If not, upload it
   - Generate UV coordinates from UVRect parameter
   - Create/update bind group with texture
   - Draw with texture pipeline

#### Success Criteria:
- ✅ PNG image loads and displays on rectangle
- ✅ Texture filtering works (no pixelation with linear filter)
- ✅ UV coordinates correctly map texture to quad
- ✅ Multiple textures can be used

#### Files to Modify:
- `internal/canvas/canvas.go` - Add texture upload, update shaders
- `cmd/webgpu-triangle/main.go` - Test with simple PNG

---

### Phase 4: Sprite Sheet Support
**Goal**: Display specific frames from a sprite sheet using UV coordinates

#### Tasks:
1. **Integrate with Sprite System**
   - Connect `SpriteRenderer` to `CanvasManager`
   - Use sprite sheet's frame UV calculations
   - Test with llama.png or similar sprite sheet

2. **Implement UV Coordinate Mapping**
   - Ensure UVRect correctly maps to sprite frames
   - Support partial texture rendering
   - Handle sprite sheet atlases

3. **Test with Sprite Sheet**
   - Load a sprite sheet (e.g., 4x4 grid)
   - Display individual frames
   - Verify UV calculations are correct

#### Success Criteria:
- ✅ Individual sprite frames display correctly
- ✅ UV coordinates properly isolate each frame
- ✅ Multiple sprites from same sheet render correctly

#### Files to Modify:
- `internal/sprite/sprite.go` - Integrate with canvas
- `cmd/webgpu-triangle/main.go` - Test with sprite sheet

---

### Phase 5: Sprite Animation
**Goal**: Animate sprites by cycling through frames

#### Tasks:
1. **Update Render Loop**
   - Call `sprite.Update(deltaTime)` to advance frames
   - Call `sprite.Render()` to draw current frame

2. **Test Animation**
   - Create animated sprite
   - Start animation
   - Verify smooth frame transitions

#### Success Criteria:
- ✅ Sprites animate through frames
- ✅ Animation speed is controllable
- ✅ Smooth transitions (no stuttering)

#### Files to Modify:
- `cmd/webgpu-triangle/main.go` - Add animation test

---

### Phase 6: Advanced Features & Optimization
**Goal**: Add rotation, scaling, and optimize batch rendering

#### Tasks:
1. **Implement `DrawTextureRotated`**
   - Add rotation matrix to vertex shader
   - Rotate around sprite center
   - Update vertex positions

2. **Implement `DrawTextureScaled`**
   - Add scale factor to vertex calculations
   - Scale around sprite center

3. **Optimize Batch Rendering**
   - Implement texture atlasing
   - Minimize pipeline switches
   - Use instanced rendering if beneficial

4. **Performance Testing**
   - Test with 100+ sprites
   - Measure FPS
   - Profile and optimize bottlenecks

#### Success Criteria:
- ✅ Rotation works correctly
- ✅ Scaling works correctly
- ✅ Can render 100+ sprites at 60 FPS
- ✅ No memory leaks

---

## Technical Implementation Details

### Vertex Buffer Structure Evolution

**Phase 1 (Colored Rectangles):**
```go
type ColoredVertex struct {
    Position [2]float32  // x, y
    Color    [4]float32  // r, g, b, a
}
```

**Phase 3+ (Textured Sprites):**
```go
type TexturedVertex struct {
    Position [2]float32  // x, y
    UV       [2]float32  // u, v
}
```

### Coordinate System Conversion

WebGPU uses Normalized Device Coordinates (NDC):
- X: -1 (left) to +1 (right)
- Y: -1 (top) to +1 (bottom)

Canvas coordinates:
- X: 0 (left) to width (right)
- Y: 0 (top) to height (bottom)

Conversion formula:
```go
ndcX = (canvasX / canvasWidth) * 2.0 - 1.0
ndcY = (canvasY / canvasHeight) * 2.0 - 1.0
```

### Pipeline Configuration

**Sprite Pipeline:**
```go
{
    vertex: {
        module: spriteShaderModule,
        entryPoint: "vs_main",
        buffers: [{
            arrayStride: 16, // 4 floats * 4 bytes
            attributes: [
                { shaderLocation: 0, offset: 0, format: "float32x2" },  // position
                { shaderLocation: 1, offset: 8, format: "float32x2" },  // uv
            ]
        }]
    },
    fragment: {
        module: spriteShaderModule,
        entryPoint: "fs_main",
        targets: [{
            format: canvasFormat,
            blend: {
                color: { srcFactor: "src-alpha", dstFactor: "one-minus-src-alpha", operation: "add" },
                alpha: { srcFactor: "one", dstFactor: "one-minus-src-alpha", operation: "add" }
            }
        }]
    },
    primitive: {
        topology: "triangle-list"
    }
}
```

### Render Pass Integration

Current render loop draws triangle. Need to:
1. Keep triangle rendering as-is (for now)
2. Add sprite rendering pass after triangle
3. Eventually merge into single pass with multiple draw calls

**Render Order:**
1. Clear screen (black)
2. Draw triangle (existing)
3. Draw sprites (new)
4. Present

---

## Code Structure Organization

### New Methods in `canvas.go`:

```go
// Shader management
func (w *WebGPUCanvasManager) createSpriteShaders() error
func (w *WebGPUCanvasManager) createSpritePipeline() error

// Texture management
func (w *WebGPUCanvasManager) uploadTextureToGPU(texture types.Texture) (js.Value, error)
func (w *WebGPUCanvasManager) createBindGroupForTexture(gpuTexture js.Value) (js.Value, error)

// Vertex buffer management
func (w *WebGPUCanvasManager) createVertexBuffer(size int) error
func (w *WebGPUCanvasManager) updateVertexBuffer(vertices []float32) error

// Coordinate conversion
func (w *WebGPUCanvasManager) canvasToNDC(x, y float64) (float32, float32)
func (w *WebGPUCanvasManager) generateQuadVertices(pos types.Vector2, size types.Vector2, uv types.UVRect) []float32

// Rendering
func (w *WebGPUCanvasManager) renderSprites() error
```

---

## Testing Strategy

### Visual Tests (Each Phase)
1. **Phase 1**: See a blue rectangle at (100, 100)
2. **Phase 2**: See 3-5 colored rectangles at different positions
3. **Phase 3**: See a textured rectangle with PNG image
4. **Phase 4**: See individual sprite frames from sprite sheet
5. **Phase 5**: See animated sprite cycling through frames
6. **Phase 6**: See rotating and scaling sprites

### Debug Output
- Log all WebGPU operations
- Print vertex data before upload
- Log texture upload status
- Print pipeline creation success/failure

### Error Handling
- Check for WebGPU errors after each operation
- Graceful degradation if texture upload fails
- Clear error messages in console

---

## Development Workflow

### Build & Test Commands
```bash
# Build and run
make dev

# Clean build
make clean build serve

# Run tests
make test
```

### Browser Console Monitoring
- Watch for DEBUG: messages
- Check for WebGPU warnings
- Monitor FPS using browser DevTools

---

## Performance Targets

- **Phase 1-2**: 60 FPS with 10 rectangles
- **Phase 3**: 60 FPS with 10 textured sprites
- **Phase 4-5**: 60 FPS with 20 animated sprites
- **Phase 6**: 60 FPS with 100+ sprites (with batching)

---

## Risk Areas & Mitigation

### Risk 1: Coordinate System Confusion
- **Mitigation**: Create helper functions early, test with known positions

### Risk 2: Texture Upload Complexity
- **Mitigation**: Start with small test images, add debug logging

### Risk 3: Pipeline State Management
- **Mitigation**: Keep triangle and sprite pipelines separate initially

### Risk 4: Memory Leaks with js.Value
- **Mitigation**: Properly release JavaScript objects, monitor memory

---

## Next Immediate Steps

1. **Start Phase 1 Implementation**
   - Create sprite shaders (colored, no texture)
   - Add sprite pipeline creation to `setupWebGPUTriangle`
   - Implement coordinate conversion helpers
   - Implement basic `DrawTexture` with colored rectangles

2. **Test Early and Often**
   - Test after each major change
   - Verify with console logs
   - Check browser DevTools for WebGPU errors

3. **Document as You Go**
   - Add inline comments explaining WebGPU calls
   - Update this plan as implementation reveals new details

---

## Success Definition

**Phase 1 Complete**: Blue rectangle renders using WebGPU sprite pipeline  
**Phase 3 Complete**: PNG texture displays on rectangle  
**Phase 5 Complete**: Animated sprite from sprite sheet  
**Final Goal**: Full-featured sprite rendering system with batching and performance optimization

---

**Status**: Ready to begin Phase 1 implementation

