//go:generate go run github.com/valyala/quicktemplate/qtc

package components

import (
	"fmt"
	"syscall/js"

	"github.com/cstevenson98/goFE/pkg/goFE"
	"github.com/cstevenson98/goFE/pkg/goFE/utils"
	"github.com/cstevenson98/goFE/pkg/shared"
	"github.com/google/uuid"
)

// AnthropicRequest represents a request to the Anthropic agent
type AnthropicRequest struct {
	Message string `json:"message"`
}

// AnthropicResponse represents a response from the Anthropic agent
type AnthropicResponse struct {
	Response string `json:"response"`
}

// StreamSetupResponse represents the response when setting up a stream
type StreamSetupResponse struct {
	SessionId string `json:"sessionId"`
}

// LilyPondDocument represents a LilyPond document from the API
type LilyPondDocument struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type AnthropicAgentExample struct {
	id                     uuid.UUID
	promptInputID          uuid.UUID
	sendButtonID           uuid.UUID
	streamButtonID         uuid.UUID
	formID                 uuid.UUID
	editAreaID             uuid.UUID
	pdfViewerID            uuid.UUID
	documentSelectID       uuid.UUID
	createDocumentButtonID uuid.UUID
	saveDocumentButtonID   uuid.UUID
	compileButtonID        uuid.UUID
	deleteButtonID         uuid.UUID
	sourceButtonID         uuid.UUID
	filePathButtonID       uuid.UUID

	// State
	promptInput       string
	assistantResponse string
	editContent       string
	currentDocumentID string
	lilypondDocuments []LilyPondDocument
	loading           bool
	streaming         bool
	compiling         bool
	error             string
	eventSource       *utils.EventSource
	tokenCount        int
	documentInfo      string
	pdfUrl            string
}

func NewAnthropicAgentExample() *AnthropicAgentExample {
	component := &AnthropicAgentExample{
		id:                     uuid.New(),
		promptInputID:          uuid.New(),
		sendButtonID:           uuid.New(),
		streamButtonID:         uuid.New(),
		formID:                 uuid.New(),
		editAreaID:             uuid.New(),
		pdfViewerID:            uuid.New(),
		documentSelectID:       uuid.New(),
		createDocumentButtonID: uuid.New(),
		saveDocumentButtonID:   uuid.New(),
		compileButtonID:        uuid.New(),
		deleteButtonID:         uuid.New(),
		sourceButtonID:         uuid.New(),
		filePathButtonID:       uuid.New(),
		editContent:            "",
	}

	// Load documents immediately
	component.loadDocuments()

	return component
}

func (a *AnthropicAgentExample) GetID() uuid.UUID {
	return a.id
}

func (a *AnthropicAgentExample) GetChildren() []goFE.Component {
	return nil
}

func (a *AnthropicAgentExample) InitEventListeners() {
	// Add form submission handler
	goFE.GetDocument().AddEventListener(a.formID, "submit", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Prevent default form submission
		args[0].Call("preventDefault")
		a.sendMessage()
		return nil
	}))

	// Add event listener for textarea input
	goFE.GetDocument().AddEventListener(a.promptInputID, "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		a.promptInput = this.Get("value").String()
		return nil
	}))

	// Add event listener for Enter key in textarea
	goFE.GetDocument().AddEventListener(a.promptInputID, "keydown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		if event.Get("key").String() == "Enter" && !event.Get("shiftKey").Bool() {
			event.Call("preventDefault")
			a.sendMessage()
		}
		return nil
	}))

	// Add event listener for stream button
	goFE.GetDocument().AddEventListener(a.streamButtonID, "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		a.sendStreamMessage()
		return nil
	}))

	// Add event listener for edit area
	goFE.GetDocument().AddEventListener(a.editAreaID, "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		a.editContent = this.Get("value").String()
		return nil
	}))

	// Add event listener for document select
	goFE.GetDocument().AddEventListener(a.documentSelectID, "change", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		a.currentDocumentID = this.Get("value").String()
		if a.currentDocumentID != "" {
			a.loadDocument()
		}
		return nil
	}))

	// Add event listener for create document button
	goFE.GetDocument().AddEventListener(a.createDocumentButtonID, "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		a.createDocument()
		return nil
	}))

	// Add event listener for save document button
	goFE.GetDocument().AddEventListener(a.saveDocumentButtonID, "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		a.saveDocument()
		return nil
	}))

	// Add event listener for compile button
	goFE.GetDocument().AddEventListener(a.compileButtonID, "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		a.compileDocument()
		return nil
	}))

	// Add event listener for delete button
	goFE.GetDocument().AddEventListener(a.deleteButtonID, "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		a.deleteDocument()
		return nil
	}))

	// Add event listener for source button
	goFE.GetDocument().AddEventListener(a.sourceButtonID, "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		a.getDocumentSource()
		return nil
	}))

	// Add event listener for file path button
	goFE.GetDocument().AddEventListener(a.filePathButtonID, "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		a.getDocumentFilePath()
		return nil
	}))
}

func (a *AnthropicAgentExample) sendMessage() {
	if a.promptInput == "" {
		a.error = "Please enter a prompt"
		a.updateUI()
		return
	}

	// Set loading state
	a.loading = true
	a.error = ""
	a.assistantResponse = ""
	a.tokenCount = 0
	a.updateUI()

	// Prepare request
	request := AnthropicRequest{
		Message: a.promptInput,
	}

	// Send message to the Anthropic agent API
	response, err := utils.PostJSON[shared.APIResponse[AnthropicResponse]]("http://localhost:8081/api/chat", request)

	// Reset loading state
	a.loading = false

	if err != nil {
		a.error = fmt.Sprintf("Error sending message: %v", err)
		a.assistantResponse = ""
	} else {
		a.assistantResponse = response.Data.Data.Response
		a.error = ""
		// Clear the input after successful send
		a.promptInput = ""
	}

	a.updateUI()
}

func (a *AnthropicAgentExample) sendStreamMessage() {
	if a.promptInput == "" {
		a.error = "Please enter a prompt"
		a.updateUI()
		return
	}

	// Set streaming state
	a.streaming = true
	a.error = ""
	a.assistantResponse = ""
	a.tokenCount = 0
	a.updateUI()

	// Prepare request
	request := AnthropicRequest{
		Message: a.promptInput,
	}

	// Set up the stream session
	response, err := utils.PostJSON[shared.APIResponse[StreamSetupResponse]]("http://localhost:8081/api/chat/stream", request)

	if err != nil {
		a.streaming = false
		a.error = fmt.Sprintf("Error setting up stream: %v", err)
		a.updateUI()
		return
	}

	sessionId := response.Data.Data.SessionId
	if sessionId == "" {
		a.streaming = false
		a.error = "Failed to get session ID for streaming"
		a.updateUI()
		return
	}

	// Create EventSource for streaming
	streamURL := fmt.Sprintf("http://localhost:8081/api/chat/stream/%s", sessionId)
	a.eventSource = utils.CreateEventSource(streamURL)

	// Handle message events (streaming chunks)
	a.eventSource.AddEventListener("message", func(event utils.EventSourceEvent) {
		a.assistantResponse += event.Data
		// Count tokens (rough approximation: split by whitespace)
		tokens := len(event.Data)
		a.tokenCount += tokens
		a.updateUI()
	})

	// Handle complete event
	a.eventSource.AddEventListener("complete", func(event utils.EventSourceEvent) {
		a.streaming = false
		a.promptInput = "" // Clear input after successful stream
		if a.eventSource != nil {
			a.eventSource.Close()
			a.eventSource = nil
		}
		a.updateUI()
	})

	// Handle error events
	a.eventSource.AddEventListener("error", func(event utils.EventSourceEvent) {
		a.streaming = false
		a.error = fmt.Sprintf("Stream error: %s", event.Data)
		if a.eventSource != nil {
			a.eventSource.Close()
			a.eventSource = nil
		}
		a.updateUI()
	})
}

func (a *AnthropicAgentExample) loadDocuments() {
	// Load LilyPond documents from the API
	lilypondResponse, err := utils.GetJSON[shared.APIResponse[map[string]interface{}]]("http://localhost:8081/api/lilypond")

	if err != nil {
		println("Error loading LilyPond documents:", err.Error())
	} else if lilypondResponse.Data.Success {
		if documentsData, ok := lilypondResponse.Data.Data["documents"]; ok {
			if documentsArray, ok := documentsData.([]interface{}); ok {
				a.lilypondDocuments = make([]LilyPondDocument, 0, len(documentsArray))
				for _, doc := range documentsArray {
					if docMap, ok := doc.(map[string]interface{}); ok {
						document := LilyPondDocument{
							ID:        getString(docMap, "id"),
							Title:     getString(docMap, "title"),
							Content:   getString(docMap, "content"),
							Status:    getString(docMap, "status"),
							CreatedAt: getString(docMap, "created_at"),
							UpdatedAt: getString(docMap, "updated_at"),
						}
						a.lilypondDocuments = append(a.lilypondDocuments, document)
					}
				}
				println("Loaded", len(a.lilypondDocuments), "LilyPond documents")
			}
		}
	} else {
		println("Failed to load LilyPond documents:", lilypondResponse.Data.Error)
	}
}

// Helper function to safely extract string values from interface{} maps
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func (a *AnthropicAgentExample) loadDocument() {
	if a.currentDocumentID == "" {
		return
	}

	// Clear PDF URL when loading a new document
	a.pdfUrl = ""

	// Try to load as LilyPond document
	lilypondResponse, err := utils.GetJSON[shared.APIResponse[map[string]interface{}]]("http://localhost:8081/api/lilypond/" + a.currentDocumentID)

	if err == nil && lilypondResponse.Data.Success {
		if content, ok := lilypondResponse.Data.Data["content"]; ok {
			if contentStr, ok := content.(string); ok {
				a.editContent = contentStr
				a.updateUI()
				return
			}
		}
	}

	a.error = "Failed to load document"
	a.updateUI()
}

func (a *AnthropicAgentExample) createDocument() {
	// Initialize editor with basic LilyPond template for new documents
	if a.editContent == "" {
		a.editContent = `\version "2.22.1"

\header {
  title = "New Music Score"
  composer = "Music Assistant"
}

\score {
  \new Staff {
    \clef treble
    \time 4/4
    \key c \major
    
    c'4 d'4 e'4 f'4 |
    g'4 a'4 b'4 c''4 |
  }
}`
	}

	// Create a new LilyPond document
	request := struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}{
		Title:   "New Music Score",
		Content: a.editContent,
	}

	response, err := utils.PostJSON[shared.APIResponse[map[string]interface{}]]("http://localhost:8081/api/lilypond", request)

	if err != nil {
		a.error = fmt.Sprintf("Error creating document: %v", err)
	} else {
		if id, ok := response.Data.Data["id"]; ok {
			if idStr, ok := id.(string); ok {
				a.currentDocumentID = idStr
				a.refreshDocuments() // Refresh document list
				a.error = ""
			}
		}
	}

	a.updateUI()
}

func (a *AnthropicAgentExample) saveDocument() {
	if a.currentDocumentID == "" {
		a.error = "No document selected to save"
		a.updateUI()
		return
	}

	// Update the document
	request := struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}{
		Title:   "Updated Music Score",
		Content: a.editContent,
	}

	_, err := utils.PutJSON[shared.APIResponse[map[string]interface{}]]("http://localhost:8081/api/lilypond/"+a.currentDocumentID, request)

	if err != nil {
		a.error = fmt.Sprintf("Error saving document: %v", err)
	} else {
		a.error = ""
	}

	a.updateUI()
}

func (a *AnthropicAgentExample) compileDocument() {
	if a.currentDocumentID == "" {
		a.error = "No document selected to compile"
		a.updateUI()
		return
	}

	// Set compiling state
	a.compiling = true
	a.error = ""
	a.documentInfo = ""
	a.pdfUrl = ""
	a.updateUI()

	// Compile the LilyPond document using the new function that handles error bodies
	url := "http://localhost:8081/api/lilypond/" + a.currentDocumentID + "/compile"
	response, errorResponse, err := utils.PostJSONWithErrorBody[shared.APIResponse[map[string]interface{}]](url, struct{}{})

	// Reset compiling state
	a.compiling = false

	if err != nil {
		// If we have an error response with body, use that for detailed error info
		if errorResponse != nil {
			errorMsg := ""

			// Add stderr information if available (this contains the actual LilyPond error)
			if stderr, ok := errorResponse.Data.Data["stderr"].(string); ok && stderr != "" {
				errorMsg = stderr
			} else {
				// Fallback to the generic error message
				errorMsg = errorResponse.Data.Error
			}

			// Add stdout information if available
			if stdout, ok := errorResponse.Data.Data["stdout"].(string); ok && stdout != "" {
				errorMsg += fmt.Sprintf("\n\nSTDOUT:\n%s", stdout)
			}

			a.error = errorMsg
		} else {
			a.error = fmt.Sprintf("Error compiling document: %v", err)
		}
		a.pdfUrl = "" // Clear PDF URL to show error in PDF container
	} else if !response.Data.Success {
		// Handle API error response - include stdout/stderr if available
		errorMsg := ""

		// Add stderr information if available (this contains the actual LilyPond error)
		if stderr, ok := response.Data.Data["stderr"].(string); ok && stderr != "" {
			errorMsg = stderr
		} else {
			// Fallback to the generic error message
			errorMsg = response.Data.Error
		}

		// Add stdout information if available
		if stdout, ok := response.Data.Data["stdout"].(string); ok && stdout != "" {
			errorMsg += fmt.Sprintf("\n\nSTDOUT:\n%s", stdout)
		}

		a.error = errorMsg
		a.pdfUrl = "" // Clear PDF URL to show error in PDF container
	} else {
		a.documentInfo = "Document compiled successfully!"
		a.error = ""

		// Set the PDF URL for display
		a.pdfUrl = fmt.Sprintf("http://localhost:8081/api/lilypond/%s/pdf", a.currentDocumentID)
	}

	a.updateUI()
}

func (a *AnthropicAgentExample) deleteDocument() {
	if a.currentDocumentID == "" {
		a.error = "No document selected to delete"
		a.updateUI()
		return
	}

	// Delete the LilyPond document
	_, err := utils.DeleteJSON[shared.APIResponse[map[string]interface{}]]("http://localhost:8081/api/lilypond/" + a.currentDocumentID)

	if err != nil {
		a.error = fmt.Sprintf("Error deleting document: %v", err)
	} else {
		// Clear current document state
		a.currentDocumentID = ""
		a.editContent = ""
		a.documentInfo = ""
		a.pdfUrl = ""
		a.error = ""

		// Refresh document list to update dropdown
		a.refreshDocuments()
		return // Don't call updateUI() here since refreshDocuments() calls it
	}

	a.updateUI()
}

func (a *AnthropicAgentExample) getDocumentSource() {
	if a.currentDocumentID == "" {
		a.error = "No document selected"
		a.updateUI()
		return
	}

	// Get the LilyPond source code
	response, err := utils.GetJSON[shared.APIResponse[map[string]interface{}]]("http://localhost:8081/api/lilypond/" + a.currentDocumentID + "/source")

	if err != nil {
		a.error = fmt.Sprintf("Error getting document source: %v", err)
	} else {
		if source, ok := response.Data.Data["source"]; ok {
			if sourceStr, ok := source.(string); ok {
				a.documentInfo = fmt.Sprintf("Document Source:\n%s", sourceStr)
				a.error = ""
			}
		}
	}

	a.updateUI()
}

func (a *AnthropicAgentExample) getDocumentFilePath() {
	if a.currentDocumentID == "" {
		a.error = "No document selected"
		a.updateUI()
		return
	}

	// Get the LilyPond file path
	response, err := utils.GetJSON[shared.APIResponse[map[string]interface{}]]("http://localhost:8081/api/lilypond/" + a.currentDocumentID + "/file-path")

	if err != nil {
		a.error = fmt.Sprintf("Error getting document file path: %v", err)
	} else {
		if filePath, ok := response.Data.Data["file_path"]; ok {
			if filePathStr, ok := filePath.(string); ok {
				a.documentInfo = fmt.Sprintf("Document File Path:\n%s", filePathStr)
				a.error = ""
			}
		}
	}

	a.updateUI()
}

func (a *AnthropicAgentExample) updateUI() {
	// Re-render the component
	html := a.Render()
	js.Global().Get("document").Get("body").Set("innerHTML", html)

	// Re-attach event listeners
	a.InitEventListeners()
}

func (a *AnthropicAgentExample) Render() string {
	return AnthropicAgentExampleTemplate(
		a.id.String(),
		a.promptInputID.String(),
		a.sendButtonID.String(),
		a.streamButtonID.String(),
		a.formID.String(),
		a.editAreaID.String(),
		a.pdfViewerID.String(),
		a.documentSelectID.String(),
		a.createDocumentButtonID.String(),
		a.saveDocumentButtonID.String(),
		a.compileButtonID.String(),
		a.deleteButtonID.String(),
		a.sourceButtonID.String(),
		a.filePathButtonID.String(),
		a.promptInput,
		a.assistantResponse,
		a.editContent,
		a.currentDocumentID,
		a.lilypondDocuments,
		a.loading,
		a.streaming,
		a.compiling,
		a.error,
		a.tokenCount,
		a.documentInfo,
		a.pdfUrl,
	)
}

func (a *AnthropicAgentExample) refreshDocuments() {
	a.loadDocuments()
	a.updateUI()
}
