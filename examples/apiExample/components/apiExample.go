//go:generate go run github.com/valyala/quicktemplate/qtc	

package components

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/cstevenson98/goFE/pkg/goFE"
	"github.com/cstevenson98/goFE/pkg/goFE/utils"
	"github.com/cstevenson98/goFE/pkg/shared"
	"github.com/google/uuid"
)

type APIExample struct {
	id       uuid.UUID
	fetchID  uuid.UUID
	response string
	loading  bool
	error    string
}

func NewAPIExample() *APIExample {
	return &APIExample{
		id:      uuid.New(),
		fetchID: uuid.New(),
	}
}

func (a *APIExample) GetID() uuid.UUID {
	return a.id
}

func (a *APIExample) GetChildren() []goFE.Component {
	return nil
}

func (a *APIExample) InitEventListeners() {
	goFE.GetDocument().AddEventListener(a.fetchID, "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		a.fetchEndpoints()
		return nil
	}))
}

func (a *APIExample) fetchEndpoints() {
	// Set loading state
	a.loading = true
	a.error = ""
	a.response = ""
	a.updateUI()

	// Fetch from the API
	response, err := utils.GetJSON[shared.APIResponse[map[string]interface{}]]("http://localhost:8080/endpoints")

	// Reset loading state
	a.loading = false

	if err != nil {
		a.error = fmt.Sprintf("Error fetching endpoints: %v", err)
		a.response = ""
	} else {
		// Format the JSON response
		jsonBytes, _ := json.MarshalIndent(response.Data, "", "  ")
		a.response = string(jsonBytes)
		a.error = ""
	}

	a.updateUI()
}

func (a *APIExample) updateUI() {
	// Re-render the component
	html := a.Render()
	js.Global().Get("document").Get("body").Set("innerHTML", html)

	// Re-attach event listeners
	a.InitEventListeners()
}

func (a *APIExample) Render() string {
	return APIExampleTemplate(a.id.String(), a.fetchID.String(), a.response, a.loading, a.error)
}
