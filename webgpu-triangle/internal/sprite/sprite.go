package sprite

import (
	"github.com/conor/webgpu-triangle/internal/types"
)

// SpriteSheet represents an animated sprite sheet with n x m frames
type SpriteSheet struct {
	// Texture information
	texturePath string

	// Position and size
	position Vector2
	size     Vector2

	// Movement
	velocity     Vector2 // Movement velocity (pixels per second)
	screenWidth  float64 // Screen width for wrapping
	screenHeight float64 // Screen height for wrapping

	// Sprite sheet grid
	columns int // n columns
	rows    int // m rows

	// Animation state
	currentFrame int
	totalFrames  int
	frameTime    float64 // Time per frame in seconds
	elapsed      float64 // Elapsed time for current frame

	// Rendering state
	visible bool
}

// Vector2 is a 2D vector (position, size, etc.)
type Vector2 struct {
	X float64
	Y float64
}

// NewSpriteSheet creates a new sprite sheet
// texturePath: path to the sprite sheet texture
// position: world position of the sprite
// size: display size of the sprite
// columns: number of columns (n) in the sprite sheet
// rows: number of rows (m) in the sprite sheet
func NewSpriteSheet(texturePath string, position Vector2, size Vector2, columns, rows int) *SpriteSheet {
	return &SpriteSheet{
		texturePath:  texturePath,
		position:     position,
		size:         size,
		velocity:     Vector2{X: 0, Y: 0}, // No movement by default
		screenWidth:  800,                 // Default screen width
		screenHeight: 600,                 // Default screen height
		columns:      columns,
		rows:         rows,
		currentFrame: 0,
		totalFrames:  columns * rows,
		frameTime:    0.1, // Default 10 FPS
		elapsed:      0.0,
		visible:      true,
	}
}

// GetSpriteRenderData returns the data needed to render this sprite
func (s *SpriteSheet) GetSpriteRenderData() types.SpriteRenderData {
	// Calculate UV coordinates for the current frame
	frameWidth := 1.0 / float64(s.columns)
	frameHeight := 1.0 / float64(s.rows)

	frameX := s.currentFrame % s.columns
	frameY := s.currentFrame / s.columns

	uv := types.UVRect{
		U: float64(frameX) * frameWidth,
		V: float64(frameY) * frameHeight,
		W: frameWidth,
		H: frameHeight,
	}

	return types.SpriteRenderData{
		TexturePath: s.texturePath,
		Position:    types.Vector2{X: s.position.X, Y: s.position.Y},
		Size:        types.Vector2{X: s.size.X, Y: s.size.Y},
		UV:          uv,
		Visible:     s.visible,
	}
}

// Update updates the sprite's animation state and position
func (s *SpriteSheet) Update(deltaTime float64) {
	// Update animation
	if s.totalFrames > 1 {
		s.elapsed += deltaTime

		// Advance to next frame if enough time has passed
		if s.elapsed >= s.frameTime {
			s.elapsed -= s.frameTime
			s.currentFrame = (s.currentFrame + 1) % s.totalFrames
		}
	}

	// Update position based on velocity
	s.position.X += s.velocity.X * deltaTime
	s.position.Y += s.velocity.Y * deltaTime

	// Screen wrapping - loop back when going off screen
	if s.velocity.X > 0 { // Moving right
		if s.position.X > s.screenWidth {
			s.position.X = -s.size.X // Wrap to left side (just off screen)
		}
	} else if s.velocity.X < 0 { // Moving left
		if s.position.X+s.size.X < 0 {
			s.position.X = s.screenWidth // Wrap to right side
		}
	}

	if s.velocity.Y > 0 { // Moving down
		if s.position.Y > s.screenHeight {
			s.position.Y = -s.size.Y // Wrap to top
		}
	} else if s.velocity.Y < 0 { // Moving up
		if s.position.Y+s.size.Y < 0 {
			s.position.Y = s.screenHeight // Wrap to bottom
		}
	}
}

// SetPosition sets the sprite's position
func (s *SpriteSheet) SetPosition(pos types.Vector2) {
	s.position = Vector2{X: pos.X, Y: pos.Y}
}

// GetPosition returns the sprite's current position
func (s *SpriteSheet) GetPosition() types.Vector2 {
	return types.Vector2{X: s.position.X, Y: s.position.Y}
}

// SetVisible sets whether the sprite should be rendered
func (s *SpriteSheet) SetVisible(visible bool) {
	s.visible = visible
}

// IsVisible returns whether the sprite is visible
func (s *SpriteSheet) IsVisible() bool {
	return s.visible
}

// SetFrameTime sets the time per frame for animation (in seconds)
func (s *SpriteSheet) SetFrameTime(frameTime float64) {
	s.frameTime = frameTime
}

// SetCurrentFrame sets the current frame index (0-based)
func (s *SpriteSheet) SetCurrentFrame(frame int) {
	if frame >= 0 && frame < s.totalFrames {
		s.currentFrame = frame
	}
}

// GetCurrentFrame returns the current frame index
func (s *SpriteSheet) GetCurrentFrame() int {
	return s.currentFrame
}

// GetTotalFrames returns the total number of frames
func (s *SpriteSheet) GetTotalFrames() int {
	return s.totalFrames
}

// SetVelocity sets the sprite's velocity (pixels per second)
func (s *SpriteSheet) SetVelocity(velocity Vector2) {
	s.velocity = velocity
}

// GetVelocity returns the sprite's current velocity
func (s *SpriteSheet) GetVelocity() Vector2 {
	return s.velocity
}

// SetScreenBounds sets the screen boundaries for wrapping
func (s *SpriteSheet) SetScreenBounds(width, height float64) {
	s.screenWidth = width
	s.screenHeight = height
}
