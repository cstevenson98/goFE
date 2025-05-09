//go:generate go run github.com/valyala/quicktemplate/qtc

package home

import (
	"github.com/cstevenson98/goFE/pkg/goFE"
	"github.com/google/uuid"
)

// Props defines the home component props
type Props struct{}

// Home represents the home page component
type Home struct {
	id uuid.UUID
}

// NewHome creates a new home component
func NewHome(_ Props) *Home {
	println("Home: Creating new home component")
	h := &Home{
		id: uuid.New(),
	}
	println("Home: Component created with ID:", h.id.String())
	return h
}

// GetID returns the component ID
func (h *Home) GetID() uuid.UUID {
	return h.id
}

// Render renders the home component
func (h *Home) Render() string {
	println("Home: Rendering home component")
	return HomeTemplate(h.id.String())
}

// GetChildren returns child components
func (h *Home) GetChildren() []goFE.Component {
	return nil
}

// InitEventListeners sets up event listeners
func (h *Home) InitEventListeners() {
	println("Home: Initializing event listeners (none needed)")
	// No event listeners needed for this component
} 