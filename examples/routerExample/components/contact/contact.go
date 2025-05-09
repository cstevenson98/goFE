//go:generate go run github.com/valyala/quicktemplate/qtc

package contact

import (
	"github.com/cstevenson98/goFE/pkg/goFE"
	"github.com/google/uuid"
	"syscall/js"
)

// Props defines the contact component props
type Props struct{}

// submissionState only tracks if the form has been submitted
type submissionState struct {
	submitted bool
}

// Contact represents the contact page component
type Contact struct {
	id        uuid.UUID
	formID    uuid.UUID
	nameID    uuid.UUID
	emailID   uuid.UUID
	messageID uuid.UUID
	submitID  uuid.UUID
	
	// Form data stored as regular variables
	name      string
	email     string
	message   string
	
	// Only the submission status is tracked in state
	state     *goFE.State[submissionState]
	setState  func(*submissionState)
}

// NewContact creates a new contact component
func NewContact(_ Props) *Contact {
	println("Contact: Creating new contact component")
	c := &Contact{
		id:        uuid.New(),
		formID:    uuid.New(),
		nameID:    uuid.New(),
		emailID:   uuid.New(),
		messageID: uuid.New(),
		submitID:  uuid.New(),
		name:      "",
		email:     "",
		message:   "",
	}
	c.state, c.setState = goFE.NewState[submissionState](c, &submissionState{
		submitted: false,
	})
	println("Contact: Component created with ID:", c.id.String())
	return c
}

// GetID returns the component ID
func (c *Contact) GetID() uuid.UUID {
	return c.id
}

// Render renders the contact component
func (c *Contact) Render() string {
	println("Contact: Rendering contact form, submitted:", c.state.Value.submitted)
	return ContactTemplate(
		c.id.String(),
		c.formID.String(),
		c.nameID.String(),
		c.emailID.String(),
		c.messageID.String(),
		c.submitID.String(),
		c.name,
		c.email,
		c.message,
		c.state.Value.submitted,
	)
}

// GetChildren returns child components
func (c *Contact) GetChildren() []goFE.Component {
	return nil
}

// elementExists checks if a DOM element with the given ID exists
func elementExists(id string) bool {
	element := js.Global().Get("document").Call("getElementById", id)
	return !element.IsNull() && !element.IsUndefined()
}

// InitEventListeners sets up event listeners
func (c *Contact) InitEventListeners() {
	println("Contact: Initializing event listeners")
	doc := goFE.GetDocument()
	
	// We need to wait a moment for the DOM to update before adding event listeners
	// Use setTimeout to delay event listener attachment
	js.Global().Get("setTimeout").Invoke(js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		println("Contact: Setting up delayed event listeners")
		
		// Form submission handler
		if elementExists(c.formID.String()) {
			println("Contact: Adding submit listener to form:", c.formID.String())
			doc.AddEventListener(c.formID, "submit", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				println("Contact: Form submitted")
				// Prevent default form submission
				args[0].Call("preventDefault")
				
				println("Contact: Form data - Name:", c.name, "Email:", c.email, "Message length:", len(c.message))
				
				// Only update state when the form is submitted
				c.setState(&submissionState{
					submitted: true,
				})
				
				println("Contact: Form marked as submitted")
				return nil
			}))
		} else {
			println("Contact: Form element not found:", c.formID.String())
		}
		
		// Input change handlers - update instance variables directly instead of state
		if elementExists(c.nameID.String()) {
			println("Contact: Adding input listener to name field:", c.nameID.String())
			doc.AddEventListener(c.nameID, "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				// Get the value from the input element (this is the element that triggered the event)
				c.name = this.Get("value").String()
				println("Contact: Name input changed:", c.name)
				return nil
			}))
		} else {
			println("Contact: Name input element not found:", c.nameID.String())
		}
		
		if elementExists(c.emailID.String()) {
			println("Contact: Adding input listener to email field:", c.emailID.String())
			doc.AddEventListener(c.emailID, "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				// Get the value from the input element (this is the element that triggered the event)
				c.email = this.Get("value").String()
				println("Contact: Email input changed:", c.email)
				return nil
			}))
		} else {
			println("Contact: Email input element not found:", c.emailID.String())
		}
		
		if elementExists(c.messageID.String()) {
			println("Contact: Adding input listener to message field:", c.messageID.String())
			doc.AddEventListener(c.messageID, "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				// Get the value from the textarea element (this is the element that triggered the event)
				c.message = this.Get("value").String()
				println("Contact: Message input changed, length:", len(c.message))
				return nil
			}))
		} else {
			println("Contact: Message textarea element not found:", c.messageID.String())
		}
		
		// Reset button handler (only visible after submission)
		if elementExists(c.submitID.String()) {
			println("Contact: Adding click listener to reset button:", c.submitID.String())
			doc.AddEventListener(c.submitID, "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				println("Contact: Reset button clicked")
				if c.state.Value.submitted {
					println("Contact: Clearing form")
					// Reset the form data
					c.name = ""
					c.email = ""
					c.message = ""
					// Update state to hide the thank you message
					c.setState(&submissionState{
						submitted: false,
					})
				}
				return nil
			}))
		} else {
			println("Contact: Reset button element not found:", c.submitID.String())
		}
		
		println("Contact: All event listeners initialized")
		return nil
	}), 10) // 10ms delay to ensure DOM is updated
	
	println("Contact: Scheduled event listener setup")
} 