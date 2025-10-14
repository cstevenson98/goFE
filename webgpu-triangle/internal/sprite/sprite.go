package sprite

import (
	"syscall/js"
	"time"

	"github.com/conor/webgpu-triangle/internal/canvas"
	"github.com/conor/webgpu-triangle/internal/types"
)

// SpriteSheet represents a spritesheet with animation frames
type SpriteSheet struct {
	texture     js.Value
	width       int
	height      int
	frameWidth  int
	frameHeight int
	cols        int
	rows        int
	totalFrames int
}

// Sprite represents an animated sprite
type Sprite struct {
	spritesheet    *SpriteSheet
	position       types.Vector2
	size           types.Vector2
	currentFrame   int
	animationSpeed float64 // frames per second
	lastFrameTime  time.Time
	isAnimating    bool
	loop           bool
}

// SpriteRenderer handles sprite rendering and animation
type SpriteRenderer struct {
	canvasManager canvas.CanvasManager
	sprites       []*Sprite
	spritesheets  map[string]*SpriteSheet
}

// NewSpriteRenderer creates a new sprite renderer
func NewSpriteRenderer(canvasManager canvas.CanvasManager) *SpriteRenderer {
	return &SpriteRenderer{
		canvasManager: canvasManager,
		sprites:       make([]*Sprite, 0),
		spritesheets:  make(map[string]*SpriteSheet),
	}
}

// LoadSpriteSheet loads a spritesheet from an image
func (sr *SpriteRenderer) LoadSpriteSheet(id, imagePath string, frameWidth, frameHeight int) (*SpriteSheet, error) {
	println("DEBUG: Loading spritesheet", id, "from", imagePath)

	// Create a new spritesheet
	spritesheet := &SpriteSheet{
		width:       0,
		height:      0,
		frameWidth:  frameWidth,
		frameHeight: frameHeight,
		cols:        0,
		rows:        0,
		totalFrames: 0,
	}

	// Load the image using JavaScript
	image := js.Global().Get("Image").New()

	// Create the onload handler and store it to prevent garbage collection
	onloadHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		println("DEBUG: Image loaded successfully")

		// Get the actual image dimensions
		spritesheet.width = image.Get("width").Int()
		spritesheet.height = image.Get("height").Int()
		spritesheet.cols = spritesheet.width / frameWidth
		spritesheet.rows = spritesheet.height / frameHeight
		spritesheet.totalFrames = spritesheet.cols * spritesheet.rows
		spritesheet.texture = image

		println("DEBUG: Loaded spritesheet", id, "- Size:", spritesheet.width, "x", spritesheet.height, "Frames:", spritesheet.totalFrames)

		// Update the existing spritesheet in the map
		if existing, exists := sr.spritesheets[id]; exists {
			existing.width = spritesheet.width
			existing.height = spritesheet.height
			existing.cols = spritesheet.cols
			existing.rows = spritesheet.rows
			existing.totalFrames = spritesheet.totalFrames
			existing.texture = spritesheet.texture
		}
		return nil
	})

	// Set up the onload handler
	image.Set("onload", onloadHandler)

	// Set up the onerror handler
	image.Set("onerror", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		println("DEBUG: Failed to load image:", imagePath)
		return nil
	}))

	// Set the image source
	println("DEBUG: Setting image source to:", imagePath)
	image.Set("src", imagePath)
	println("DEBUG: Image source set, waiting for load...")

	// For now, assume the image will load and return the spritesheet
	// The actual loading will happen asynchronously
	sr.spritesheets[id] = spritesheet
	return spritesheet, nil
}

// CreateSprite creates a new sprite from a spritesheet
func (sr *SpriteRenderer) CreateSprite(spritesheetID string, position types.Vector2, size types.Vector2) (*Sprite, error) {
	spritesheet, exists := sr.spritesheets[spritesheetID]
	if !exists {
		return nil, &canvas.CanvasError{Message: "Spritesheet not found: " + spritesheetID}
	}

	// Check if the spritesheet texture is loaded
	if spritesheet.texture.IsUndefined() || spritesheet.texture.IsNull() {
		println("DEBUG: Spritesheet texture not loaded yet for", spritesheetID)
		// Still create the sprite, it will be rendered once the texture loads
	}

	sprite := &Sprite{
		spritesheet:    spritesheet,
		position:       position,
		size:           size,
		currentFrame:   0,
		animationSpeed: 8.0, // 8 frames per second
		lastFrameTime:  time.Now(),
		isAnimating:    false,
		loop:           true,
	}

	sr.sprites = append(sr.sprites, sprite)
	println("DEBUG: Created sprite at position", position.X, position.Y)

	return sprite, nil
}

// StartAnimation starts sprite animation
func (s *Sprite) StartAnimation() {
	s.isAnimating = true
	s.lastFrameTime = time.Now()
	println("DEBUG: Started animation for sprite")
}

// StopAnimation stops sprite animation
func (s *Sprite) StopAnimation() {
	s.isAnimating = false
	println("DEBUG: Stopped animation for sprite")
}

// Update updates sprite animation
func (s *Sprite) Update(deltaTime float64) {
	if !s.isAnimating {
		return
	}

	now := time.Now()
	timeSinceLastFrame := now.Sub(s.lastFrameTime).Seconds()

	if timeSinceLastFrame >= 1.0/s.animationSpeed {
		s.currentFrame++
		if s.currentFrame >= s.spritesheet.totalFrames {
			if s.loop {
				s.currentFrame = 0
			} else {
				s.currentFrame = s.spritesheet.totalFrames - 1
				s.isAnimating = false
			}
		}
		s.lastFrameTime = now
	}
}

// GetCurrentFrameUV returns the UV coordinates for the current frame
func (s *Sprite) GetCurrentFrameUV() types.UVRect {
	frameCol := s.currentFrame % s.spritesheet.cols
	frameRow := s.currentFrame / s.spritesheet.cols

	u := float64(frameCol) / float64(s.spritesheet.cols)
	v := float64(frameRow) / float64(s.spritesheet.rows)
	w := 1.0 / float64(s.spritesheet.cols)
	h := 1.0 / float64(s.spritesheet.rows)

	return types.UVRect{U: u, V: v, W: w, H: h}
}

// Render renders all sprites
func (sr *SpriteRenderer) Render() error {
	for _, sprite := range sr.sprites {
		// Check if spritesheet texture is valid
		if sprite.spritesheet.texture.IsUndefined() || sprite.spritesheet.texture.IsNull() {
			println("DEBUG: Spritesheet texture not loaded yet, skipping sprite")
			continue
		}

		// Create a WebGPUTexture from the spritesheet
		texture := types.NewWebGPUTexture(
			sprite.spritesheet.width,
			sprite.spritesheet.height,
			"spritesheet-"+sprite.spritesheet.texture.String(),
			sprite.spritesheet.texture,
		)

		// Get UV coordinates for current frame
		uv := sprite.GetCurrentFrameUV()

		// Use the canvas manager to draw the texture
		err := sr.canvasManager.DrawTexture(texture, sprite.position, sprite.size, uv)
		if err != nil {
			println("DEBUG: Failed to draw sprite:", err.Error())
			// Don't return error, just continue with other sprites
			continue
		}

		println("DEBUG: Rendered sprite frame", sprite.currentFrame, "at position", sprite.position.X, sprite.position.Y)
	}

	return nil
}

// Update updates all sprites
func (sr *SpriteRenderer) Update(deltaTime float64) {
	for _, sprite := range sr.sprites {
		sprite.Update(deltaTime)
	}
}

// SetSpritePosition sets a sprite's position
func (s *Sprite) SetPosition(position types.Vector2) {
	s.position = position
}

// GetPosition returns a sprite's position
func (s *Sprite) GetPosition() types.Vector2 {
	return s.position
}

// SetAnimationSpeed sets the animation speed
func (s *Sprite) SetAnimationSpeed(speed float64) {
	s.animationSpeed = speed
}

// SetLoop sets whether the animation loops
func (s *Sprite) SetLoop(loop bool) {
	s.loop = loop
}
