package utils

import (
	"encoding/json"
	"fmt"
	"syscall/js"
)

// HTTPMethod represents HTTP methods
type HTTPMethod string

const (
	GET    HTTPMethod = "GET"
	POST   HTTPMethod = "POST"
	PUT    HTTPMethod = "PUT"
	DELETE HTTPMethod = "DELETE"
	PATCH  HTTPMethod = "PATCH"
)

// FetchOptions represents options for fetch requests
type FetchOptions struct {
	Method  HTTPMethod        `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body"`
	Timeout int               `json:"timeout"` // in milliseconds
	Mode    string            `json:"mode"`    // cors, no-cors, same-origin
	Cache   string            `json:"cache"`   // default, no-cache, reload, force-cache
}

// FetchResponse represents a fetch response
type FetchResponse[T any] struct {
	Data    T      `json:"data"`
	Status  int    `json:"status"`
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
}

// FetchError represents a fetch error
type FetchError struct {
	Message string `json:"message"`
	Status  int    `json:"status,omitempty"`
}

func (e FetchError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.Status, e.Message)
}

// FetchJSON performs a type-safe JSON fetch request
func FetchJSON[T any](url string, options *FetchOptions) (*FetchResponse[T], error) {
	// Default options
	if options == nil {
		options = &FetchOptions{
			Method:  GET,
			Headers: make(map[string]string),
		}
	}

	// Set default headers if not provided
	if options.Headers == nil {
		options.Headers = make(map[string]string)
	}
	if _, exists := options.Headers["Content-Type"]; !exists && options.Body != nil {
		options.Headers["Content-Type"] = "application/json"
	}

	// Prepare fetch options for JavaScript
	jsOptions := js.Global().Get("Object").New()
	jsOptions.Set("method", string(options.Method))

	// Set headers
	if len(options.Headers) > 0 {
		jsHeaders := js.Global().Get("Object").New()
		for key, value := range options.Headers {
			jsHeaders.Set(key, value)
		}
		jsOptions.Set("headers", jsHeaders)
	}

	// Set body if provided
	if options.Body != nil {
		bodyJSON, err := json.Marshal(options.Body)
		if err != nil {
			return nil, &FetchError{Message: fmt.Sprintf("Failed to marshal body: %v", err)}
		}
		jsOptions.Set("body", string(bodyJSON))
	}

	// Set additional options
	if options.Timeout > 0 {
		jsOptions.Set("timeout", options.Timeout)
	}
	if options.Mode != "" {
		jsOptions.Set("mode", options.Mode)
	}
	if options.Cache != "" {
		jsOptions.Set("cache", options.Cache)
	}

	// Perform the fetch
	fetchPromise := js.Global().Call("fetch", url, jsOptions)

	// Wait for the promise to resolve
	done := make(chan js.Value, 1)
	errorChan := make(chan error, 1)

	fetchPromise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		response := args[0]
		done <- response
		return nil
	})).Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		errorChan <- &FetchError{Message: args[0].Get("message").String()}
		return nil
	}))

	// Wait for response or error
	select {
	case response := <-done:
		// Check if response is ok
		if !response.Get("ok").Bool() {
			status := response.Get("status").Int()
			return nil, &FetchError{
				Message: fmt.Sprintf("HTTP %d: %s", status, response.Get("statusText").String()),
				Status:  status,
			}
		}

		// Parse JSON response
		jsonPromise := response.Call("json")
		jsonDone := make(chan js.Value, 1)
		jsonErrorChan := make(chan error, 1)

		jsonPromise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			jsonData := args[0]
			jsonDone <- jsonData
			return nil
		})).Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			jsonErrorChan <- &FetchError{Message: "Failed to parse JSON response"}
			return nil
		}))

		select {
		case jsonData := <-jsonDone:
			// Convert JavaScript object to Go struct
			var result T
			jsonStr := js.Global().Get("JSON").Call("stringify", jsonData).String()
			if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
				return nil, &FetchError{Message: fmt.Sprintf("Failed to unmarshal response: %v", err)}
			}

			return &FetchResponse[T]{
				Data:   result,
				Status: response.Get("status").Int(),
				OK:     true,
			}, nil

		case err := <-jsonErrorChan:
			return nil, err
		}

	case err := <-errorChan:
		return nil, err
	}
}

// FetchJSONWithResponse performs a type-safe JSON fetch and returns the full response
func FetchJSONWithResponse[T any](url string, options *FetchOptions) (*FetchResponse[T], error) {
	return FetchJSON[T](url, options)
}

// FetchText performs a text fetch request
func FetchText(url string, options *FetchOptions) (string, error) {
	if options == nil {
		options = &FetchOptions{Method: GET}
	}

	jsOptions := js.Global().Get("Object").New()
	jsOptions.Set("method", string(options.Method))

	// Set headers
	if len(options.Headers) > 0 {
		jsHeaders := js.Global().Get("Object").New()
		for key, value := range options.Headers {
			jsHeaders.Set(key, value)
		}
		jsOptions.Set("headers", jsHeaders)
	}

	fetchPromise := js.Global().Call("fetch", url, jsOptions)

	done := make(chan js.Value, 1)
	errorChan := make(chan error, 1)

	fetchPromise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		response := args[0]
		done <- response
		return nil
	})).Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		errorChan <- &FetchError{Message: args[0].Get("message").String()}
		return nil
	}))

	select {
	case response := <-done:
		if !response.Get("ok").Bool() {
			status := response.Get("status").Int()
			return "", &FetchError{
				Message: fmt.Sprintf("HTTP %d: %s", status, response.Get("statusText").String()),
				Status:  status,
			}
		}

		textPromise := response.Call("text")
		textDone := make(chan string, 1)
		textErrorChan := make(chan error, 1)

		textPromise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			textData := args[0].String()
			textDone <- textData
			return nil
		})).Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			textErrorChan <- &FetchError{Message: "Failed to read text response"}
			return nil
		}))

		select {
		case text := <-textDone:
			return text, nil
		case err := <-textErrorChan:
			return "", err
		}

	case err := <-errorChan:
		return "", err
	}
}

// FetchBlob performs a blob fetch request
func FetchBlob(url string, options *FetchOptions) (js.Value, error) {
	if options == nil {
		options = &FetchOptions{Method: GET}
	}

	jsOptions := js.Global().Get("Object").New()
	jsOptions.Set("method", string(options.Method))

	// Set headers
	if len(options.Headers) > 0 {
		jsHeaders := js.Global().Get("Object").New()
		for key, value := range options.Headers {
			jsHeaders.Set(key, value)
		}
		jsOptions.Set("headers", jsHeaders)
	}

	fetchPromise := js.Global().Call("fetch", url, jsOptions)

	done := make(chan js.Value, 1)
	errorChan := make(chan error, 1)

	fetchPromise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		response := args[0]
		done <- response
		return nil
	})).Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		errorChan <- &FetchError{Message: args[0].Get("message").String()}
		return nil
	}))

	select {
	case response := <-done:
		if !response.Get("ok").Bool() {
			status := response.Get("status").Int()
			return js.Undefined(), &FetchError{
				Message: fmt.Sprintf("HTTP %d: %s", status, response.Get("statusText").String()),
				Status:  status,
			}
		}

		blobPromise := response.Call("blob")
		blobDone := make(chan js.Value, 1)
		blobErrorChan := make(chan error, 1)

		blobPromise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			blobData := args[0]
			blobDone <- blobData
			return nil
		})).Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			blobErrorChan <- &FetchError{Message: "Failed to read blob response"}
			return nil
		}))

		select {
		case blob := <-blobDone:
			return blob, nil
		case err := <-blobErrorChan:
			return js.Undefined(), err
		}

	case err := <-errorChan:
		return js.Undefined(), err
	}
}

// CreateWebSocket creates a WebSocket connection
func CreateWebSocket(url string) js.Value {
	return js.Global().Get("WebSocket").New(url)
}

// Helper functions for common HTTP operations

// GetJSON performs a GET request and returns JSON data
func GetJSON[T any](url string) (*FetchResponse[T], error) {
	return FetchJSON[T](url, &FetchOptions{Method: GET})
}

// PostJSON performs a POST request with JSON data
func PostJSON[T any, U any](url string, data U) (*FetchResponse[T], error) {
	return FetchJSON[T](url, &FetchOptions{
		Method: POST,
		Body:   data,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	})
}

// PutJSON performs a PUT request with JSON data
func PutJSON[T any, U any](url string, data U) (*FetchResponse[T], error) {
	return FetchJSON[T](url, &FetchOptions{
		Method: PUT,
		Body:   data,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	})
}

// DeleteJSON performs a DELETE request
func DeleteJSON[T any](url string) (*FetchResponse[T], error) {
	return FetchJSON[T](url, &FetchOptions{Method: DELETE})
}
