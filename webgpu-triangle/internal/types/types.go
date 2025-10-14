package types

import "syscall/js"

// Vector2 represents a 2D vector
type Vector2 struct {
	X float64
	Y float64
}

// UVRect represents UV coordinates for texture sampling
type UVRect struct {
	U float64 // Left (0.0 to 1.0)
	V float64 // Top (0.0 to 1.0)
	W float64 // Width (0.0 to 1.0)
	H float64 // Height (0.0 to 1.0)
}

// Texture represents a loaded texture
type Texture interface {
	GetWidth() int
	GetHeight() int
	GetID() string
	GetWebGPUTexture() js.Value
}

// WebGPUTexture implements the Texture interface
type WebGPUTexture struct {
	Width   int
	Height  int
	ID      string
	texture js.Value
}

func (t *WebGPUTexture) GetWidth() int              { return t.Width }
func (t *WebGPUTexture) GetHeight() int             { return t.Height }
func (t *WebGPUTexture) GetID() string              { return t.ID }
func (t *WebGPUTexture) GetWebGPUTexture() js.Value { return t.texture }

// NewWebGPUTexture creates a new WebGPUTexture with the given parameters
func NewWebGPUTexture(width, height int, id string, texture js.Value) *WebGPUTexture {
	return &WebGPUTexture{
		Width:   width,
		Height:  height,
		ID:      id,
		texture: texture,
	}
}

// Pipeline represents a WebGPU render pipeline
type Pipeline interface {
	GetWebGPUPipeline() js.Value
	IsValid() bool
}

// WebGPUPipeline implements the Pipeline interface
type WebGPUPipeline struct {
	pipeline js.Value
	Valid    bool
}

func (p *WebGPUPipeline) GetWebGPUPipeline() js.Value { return p.pipeline }
func (p *WebGPUPipeline) IsValid() bool               { return p.Valid }

// SpriteVertex represents a vertex for sprite rendering
type SpriteVertex struct {
	Position Vector2 // Screen position
	UV       Vector2 // Texture coordinates
}

// SpriteUniforms represents uniform data for sprite rendering
type SpriteUniforms struct {
	Transform [16]float64 // 4x4 matrix as array
	Color     [4]float64  // RGBA color
}

// DemoState represents the current demo state
type DemoState struct {
	CurrentDemo int
	DemoTime    float64
	TotalTime   float64
}

// Demo represents a single demo
type Demo struct {
	Name        string
	Description string
	Duration    float64
	Setup       func() error
	Update      func(deltaTime float64) error
	Render      func() error
	Cleanup     func() error
}
