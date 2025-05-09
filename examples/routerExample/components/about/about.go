//go:generate go run github.com/valyala/quicktemplate/qtc

package about

import (
	"github.com/cstevenson98/goFE/pkg/goFE"
	"github.com/google/uuid"
)

// Props defines the about component props
type Props struct{}

// About represents the about page component
type About struct {
	id uuid.UUID
}

// NewAbout creates a new about component
func NewAbout(_ Props) *About {
	println("About: Creating new about component")
	a := &About{
		id: uuid.New(),
	}
	println("About: Component created with ID:", a.id.String())
	return a
}

// GetID returns the component ID
func (a *About) GetID() uuid.UUID {
	return a.id
}

// Render renders the about component
func (a *About) Render() string {
	println("About: Rendering about component")
	return AboutTemplate(a.id.String())
}

// GetChildren returns child components
func (a *About) GetChildren() []goFE.Component {
	return nil
}

// InitEventListeners sets up event listeners
func (a *About) InitEventListeners() {
	println("About: Initializing event listeners (none needed)")
	// No event listeners needed for this component
} 