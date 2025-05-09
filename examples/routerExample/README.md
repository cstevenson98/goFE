# GoFE Router Example

This example demonstrates how to create a simple client-side router using GoFE. The router handles navigation between different views without full page reloads, manages browser history, and supports the back/forward buttons.

## Features

- **Client-side routing**: Navigate between pages without reloading the application
- **Browser history integration**: Updates the URL and maintains browser history
- **Dynamic component mounting**: Loads and unloads components based on the current route
- **Active route highlighting**: Visually indicates the current route in the navigation
- **Form handling with state**: Demonstrates form state management in the Contact page

## Application Structure

```
routerExample/
├── main.go                           # Main application entry point
├── components/
│   ├── router/                       # Router component
│   │   ├── router.go                 # Router implementation
│   │   └── router.qtpl               # Router template
│   ├── home/                         # Home page component
│   │   ├── home.go                   # Home implementation
│   │   └── home.qtpl                 # Home template
│   ├── about/                        # About page component
│   │   ├── about.go                  # About implementation
│   │   └── about.qtpl                # About template
│   └── contact/                      # Contact page component
│       ├── contact.go                # Contact implementation with form
│       └── contact.qtpl              # Contact template
```

## How It Works

### 1. Router Component

The router component is the core of this example. It:

- Maintains a map of routes to components
- Listens for navigation events (link clicks)
- Updates the browser URL using the History API
- Renders the appropriate component based on the current path
- Handles back/forward button navigation via the popstate event

```go
// Route definition
routes: map[string]ViewCreator{
    "/":        func() goFE.Component { return home.NewHome(home.Props{}) },
    "/about":   func() goFE.Component { return about.NewAbout(about.Props{}) },
    "/contact": func() goFE.Component { return contact.NewContact(contact.Props{}) },
}
```

### 2. Page Components

Each page is a standard GoFE component that implements the `Component` interface. 

The Contact page demonstrates form handling with state management:
- Captures form input in state
- Handles form submission
- Displays a success message
- Provides a way to reset the form

### 3. Navigation

The router captures click events on navigation links and:
1. Prevents the default browser navigation
2. Updates the URL using `history.pushState`
3. Updates the router's state with the new path
4. Renders the appropriate component

### 4. Browser History

The router listens for the browser's `popstate` event (triggered when the user clicks back/forward buttons) and updates the current view accordingly.

## Running the Example

Build the application:

```bash
GOOS=js GOARCH=wasm go build -o main.wasm
```

Copy the WebAssembly support files and set up a simple HTML page:

```bash
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
```

Create an `index.html` that loads the WebAssembly module:

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>GoFE Router Example</title>
    <script src="wasm_exec.js"></script>
    <link href="https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css" rel="stylesheet">
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

Serve the files:

```bash
python -m http.server
```

Navigate to `http://localhost:8000` to see the example. 