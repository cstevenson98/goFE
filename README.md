# goFE

goFE is a lightweight WebAssembly-based frontend framework for Go, designed to create interactive web applications using pure Go code. It features a React-like component model with state management, event handling, and a virtual DOM-like rendering approach.

## Features

- **Component-Based Architecture**: Build UIs with reusable and composable components
- **State Management**: Type-safe state with generics and automatic re-rendering
- **Event Handling**: Simple DOM event binding through WebAssembly
- **Dynamic Component Arrays**: Efficiently manage lists of components
- **Templating**: HTML generation with QuickTemplate
- **Minimal Dependencies**: Small WASM binary size for fast loading

## Installation

```bash
go get github.com/cstevenson98/goFE
```

## Getting Started

### 1. Create a Basic Component

```go
//go:generate go run github.com/valyala/quicktemplate/qtc

package counter

import (
	"github.com/cstevenson98/goFE/pkg/goFE"
	"github.com/google/uuid"
	"syscall/js"
)

// Component state
type counterState struct {
	count int
}

// Component props
type Props struct{}

// Counter component
type Counter struct {
	id      uuid.UUID
	lowerID uuid.UUID
	raiseID uuid.UUID
	state   *goFE.State[counterState]
	setState func(*counterState)
}

// Constructor
func NewCounter(props *Props) *Counter {
	counter := &Counter{
		id:      uuid.New(),
		lowerID: uuid.New(),
		raiseID: uuid.New(),
	}
	counter.state, counter.setState = goFE.NewState[counterState](counter, &counterState{count: 0})
	return counter
}

// Required Component interface methods
func (c *Counter) GetID() uuid.UUID {
	return c.id
}

func (c *Counter) GetChildren() []goFE.Component {
	return nil
}

func (c *Counter) InitEventListeners() {
	goFE.GetDocument().AddEventListener(c.lowerID, "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		c.setState(&counterState{count: c.state.Value.count - 1})
		return nil
	}))
	goFE.GetDocument().AddEventListener(c.raiseID, "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		c.setState(&counterState{count: c.state.Value.count + 1})
		return nil
	}))
}

func (c *Counter) Render() string {
	return CounterTemplate(c.id.String(), c.state.Value.count, c.lowerID.String(), c.raiseID.String())
}
```

### 2. Create a Template

Create a QuickTemplate file (e.g., `counterTmpl.qtpl`):

```qtpl
{% func CounterTemplate(id string, count int, lowerButtonID, raiseButtonID string) %}
  <div id="{%s id %}" class="flex justify-between items-center bg-gray-100">
    <button id="{%s lowerButtonID %}">-</button>
    <span>{%d count %}</span>
    <button id="{%s raiseButtonID %}">+</button>
  </div>
{% endfunc %}
```

Generate the template code:

```bash
go generate ./...
```

### 3. Create Main Application

```go
package main

import (
	"github.com/your-username/your-project/components/counter"
	"github.com/cstevenson98/goFE/pkg/goFE"
)

func main() {
	// Initialize the framework
	goFE.Init(&goFE.Logger{
		Level: goFE.DEBUG,
	})
	
	// Create document with root component
	goFE.SetDocument(goFE.NewDocument([]goFE.Component{
		counter.NewCounter(counter.Props{}),
	}))
	
	// Initialize the document
	goFE.GetDocument().Init()
	
	// Keep the program running
	<-make(chan bool)
}
```

### 4. Build for WebAssembly

```bash
GOOS=js GOARCH=wasm go build -o main.wasm
```

### 5. Create HTML File

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>GoFE App</title>
    <script src="wasm_exec.js"></script>
    <script>
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject)
            .then((result) => { go.run(result.instance); });
    </script>
</head>
<body>
    <div id="root"></div>
</body>
</html>
```

Copy `wasm_exec.js` from your Go installation:

```bash
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
```

## Key Concepts

### Components

Components are the building blocks of your UI. Each component must implement the `Component` interface:

```go
type Component interface {
	Render() string
	GetID() uuid.UUID
	GetChildren() []Component
	InitEventListeners()
}
```

### State Management

goFE uses a generic state system similar to React's state:

```go
// Create a new state
state, setState := goFE.NewState[MyStateType](component, &initialState)

// Update state
setState(&newState)

// Add an effect (similar to React's useEffect)
state.AddEffect(func(value *MyStateType) {
    // React to state changes
})
```

### Dynamic Component Arrays

Manage lists of components efficiently:

```go
// Update an array of components based on new data
goFE.UpdateComponentArray[*MyComponent, MyProps](
    &componentArray,     // Existing components
    newLength,           // New number of components
    MyComponent.New,     // Constructor function
    newPropsArray        // Optional new props for each component
)
```

### Event Handling

Add event listeners to DOM elements:

```go
goFE.GetDocument().AddEventListener(elementID, "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
    // Event handler
    return nil
}))
```

## Advanced Example: Data Fetching

This example demonstrates fetching data from an API using WebAssembly:

```go
import (
    "context"
    "encoding/json"
    "github.com/cstevenson98/goFE/pkg/goFE"
    fetch "marwan.io/wasm-fetch"
    "time"
)

type apiData struct {
    // Your data structure
}

type dataState struct {
    loading bool
    data    *apiData
    error   string
}

// In your component
data, setData := goFE.NewState[dataState](component, &dataState{loading: true})

// Fetch data
go func() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    res, err := fetch.Fetch("https://api.example.com/data", &fetch.Opts{
        Method: fetch.MethodGet,
        Signal: ctx,
    })
    
    if err != nil {
        setData(&dataState{loading: false, error: err.Error()})
        return
    }
    
    var result apiData
    if err := json.Unmarshal(res.Body, &result); err != nil {
        setData(&dataState{loading: false, error: err.Error()})
        return
    }
    
    setData(&dataState{loading: false, data: &result})
}()
```

## Tips and Best Practices

1. **Component Organization**: Keep components in separate packages with their templates
2. **State Design**: Keep state minimal and focused on what the component needs
3. **Event Cleanup**: Be aware that event listeners need proper cleanup in complex applications
4. **Performance**: Use `UpdateComponentArray` for efficient list rendering
5. **Error Handling**: Always handle errors in async operations

## License

[License information goes here]

## Contributing

[Contribution guidelines go here]
