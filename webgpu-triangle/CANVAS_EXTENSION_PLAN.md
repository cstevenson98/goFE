# Canvas Manager Extension Plan

## Overview
Minimal extensions to the CanvasManager interface to support 2D sprite rendering. This focuses on adding the necessary WebGPU shaders, pipelines, and rendering capabilities for texture-based rendering.

## Current CanvasManager Interface
```go
type CanvasManager interface {
    Initialize(canvasID string) error
    Render() error
    Cleanup() error
    GetStatus() (bool, string)
    SetStatus(initialized bool, message string)
}
```

## Extended CanvasManager Interface
```go
type CanvasManager interface {
    // Existing methods
    Initialize(canvasID string) error
    Render() error
    Cleanup() error
    GetStatus() (bool, string)
    SetStatus(initialized bool, message string)
    
    // New rendering methods
    DrawTexture(texture Texture, position Vector2, size Vector2, uv UVRect) error
    DrawTextureRotated(texture Texture, position Vector2, size Vector2, uv UVRect, rotation float64) error
    DrawTextureScaled(texture Texture, position Vector2, size Vector2, uv UVRect, scale Vector2) error
    
    // Batch rendering
    BeginBatch() error
    EndBatch() error
    FlushBatch() error
    
    // Pipeline management
    GetSpritePipeline() Pipeline
    GetBackgroundPipeline() Pipeline
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
    GetWebGPUTexture() js.Value  // Internal WebGPU texture
}
```

### Pipeline
```go
type Pipeline interface {
    GetWebGPUPipeline() js.Value  // Internal WebGPU pipeline
    IsValid() bool
}
```

## WebGPU Shader Extensions

### Sprite Vertex Shader
```wgsl
@vertex
fn vs_main(@builtin(vertex_index) vertexIndex: u32, 
           @location(0) position: vec2f,
           @location(1) uv: vec2f) -> @builtin(position) vec4f {
    return vec4f(position, 0.0, 1.0);
}
```

### Sprite Fragment Shader
```wgsl
@fragment
fn fs_main(@location(0) uv: vec2f, 
           @binding(0) texture: texture_2d<f32>,
           @binding(1) sampler: sampler) -> @location(0) vec4f {
    return textureSample(texture, sampler, uv);
}
```

### Background Vertex Shader
```wgsl
@vertex
fn vs_main(@builtin(vertex_index) vertexIndex: u32, 
           @location(0) position: vec2f,
           @location(1) uv: vec2f) -> @builtin(position) vec4f {
    return vec4f(position, 0.0, 1.0);
}
```

### Background Fragment Shader
```wgsl
@fragment
fn fs_main(@location(0) uv: vec2f, 
           @binding(0) texture: texture_2d<f32>,
           @binding(1) sampler: sampler) -> @location(0) vec4f {
    return textureSample(texture, sampler, uv);
}
```

## Implementation Strategy

### Phase 1: Basic Types
1. Add Vector2, UVRect types
2. Create Texture interface
3. Create Pipeline interface
4. Add error types

### Phase 2: Shader Management
1. Create sprite vertex shader
2. Create sprite fragment shader
3. Create background shaders
4. Add shader compilation

### Phase 3: Pipeline Creation
1. Create sprite render pipeline
2. Create background render pipeline
3. Add pipeline caching
4. Add pipeline validation

### Phase 4: Rendering Methods
1. Implement DrawTexture
2. Implement DrawTextureRotated
3. Implement DrawTextureScaled
4. Add vertex buffer management

### Phase 5: Batch Rendering
1. Implement BeginBatch/EndBatch
2. Add vertex batching
3. Add texture batching
4. Optimize rendering performance

## WebGPU Implementation Details

### Vertex Buffer Structure
```go
type SpriteVertex struct {
    Position Vector2  // Screen position
    UV       Vector2  // Texture coordinates
}
```

### Uniform Buffer Structure
```go
type SpriteUniforms struct {
    Transform Matrix4x4  // Model-view-projection matrix
    Color     Vector4    // Tint color
}
```

### Pipeline Configuration
```go
type PipelineConfig struct {
    VertexShader   string
    FragmentShader string
    VertexLayout   VertexLayout
    BindGroupLayout BindGroupLayout
    Primitive      PrimitiveState
    DepthStencil   DepthStencilState
    Multisample    MultisampleState
}
```

## File Structure
```
canvas/
├── canvas.go           # Extended CanvasManager interface
├── types.go           # Vector2, UVRect, Texture, Pipeline
├── shaders.go         # WebGPU shader definitions
├── pipelines.go       # Pipeline creation and management
├── rendering.go       # Texture rendering methods
├── batching.go        # Batch rendering implementation
└── textures.go        # Texture loading and management
```

## Key Features
- **Texture Rendering** - Draw textures with position, size, UV coordinates
- **Rotation Support** - Rotate textures around center
- **Scaling Support** - Scale textures independently
- **Batch Rendering** - Efficient rendering of multiple sprites
- **Pipeline Management** - Separate pipelines for sprites and backgrounds
- **Shader System** - Customizable vertex and fragment shaders

## WebGPU Pipeline Setup
1. **Create Shader Modules** - Compile vertex and fragment shaders
2. **Create Bind Group Layout** - Define texture and sampler bindings
3. **Create Pipeline Layout** - Define pipeline layout
4. **Create Render Pipeline** - Combine shaders and layout
5. **Create Bind Groups** - Bind textures and samplers

## Performance Considerations
- **Vertex Batching** - Batch multiple sprites into single draw call
- **Texture Atlasing** - Combine multiple textures into single atlas
- **Pipeline Caching** - Cache compiled pipelines
- **Uniform Buffer Management** - Efficient uniform buffer updates

## Error Handling
- **Shader Compilation Errors** - Handle shader compilation failures
- **Pipeline Creation Errors** - Handle pipeline creation failures
- **Texture Loading Errors** - Handle texture loading failures
- **Rendering Errors** - Handle rendering failures gracefully

## Testing Strategy
- **Unit Tests** - Test individual methods
- **Integration Tests** - Test pipeline creation and rendering
- **Performance Tests** - Test batch rendering performance
- **Visual Tests** - Test actual rendering output

## Next Steps
1. Implement basic types and interfaces
2. Create WebGPU shaders for sprite rendering
3. Implement pipeline creation and management
4. Add texture rendering methods
5. Implement batch rendering for performance
6. Test with real sprite textures
