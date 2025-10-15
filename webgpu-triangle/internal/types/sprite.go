package types

// SpriteRenderData contains all data needed to render a sprite
type SpriteRenderData struct {
	TexturePath string
	Position    Vector2
	Size        Vector2
	UV          UVRect
	Visible     bool
}

// Sprite is the interface that all sprite types must implement
type Sprite interface {
	// GetSpriteRenderData returns the data needed to render this sprite
	GetSpriteRenderData() SpriteRenderData

	// Update updates the sprite's state (animation, etc.)
	Update(deltaTime float64)

	// SetPosition sets the sprite's position
	SetPosition(pos Vector2)

	// GetPosition returns the sprite's current position
	GetPosition() Vector2

	// SetVisible sets whether the sprite should be rendered
	SetVisible(visible bool)

	// IsVisible returns whether the sprite is visible
	IsVisible() bool
}
