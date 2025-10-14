# WebGPU Triangle - Go WASM Game Engine

A WebGPU-based game engine written in Go that compiles to WebAssembly (WASM) for browser-based games and graphics applications.

## Project Structure

This project follows Go's idiomatic project layout:

```
webgpu-triangle/
├── cmd/                          # Main applications
│   └── webgpu-triangle/          # Main application entry point
│       └── main.go              # Application entry point
├── internal/                     # Private application code
│   ├── canvas/                   # Canvas and rendering management
│   │   ├── canvas.go            # WebGPU/WebGL canvas implementation
│   │   ├── canvas_test.go       # Canvas unit tests
│   │   ├── canvas_unit_test.go  # Additional canvas tests
│   │   ├── canvas_integration_test.go # Integration tests
│   │   └── mock_canvas.go       # Mock canvas for testing
│   ├── sprite/                   # Sprite rendering and animation
│   │   └── sprite.go            # Sprite system implementation
│   └── types/                    # Shared type definitions
│       └── types.go             # Core data structures
├── pkg/                          # Public library code (if any)
├── assets/                       # Static assets
│   ├── index.html               # Main HTML file
│   ├── llama.png                # Sprite sheet image
│   ├── test-webgl.html          # WebGL test page
│   └── wasm_exec_tinygo.js      # TinyGo WASM runtime
├── go.mod                        # Go module definition
├── go.sum                        # Go module checksums
├── main.wasm                     # Compiled WebAssembly binary
├── SPRITE_IMPLEMENTATION_PLAN.md # Development roadmap
├── RENDERER_PLAN.md              # Renderer architecture plan
├── CANVAS_EXTENSION_PLAN.md     # Canvas extension plan
└── README.md                     # This file
```

## Architecture

### Core Components

1. **Canvas Management** (`internal/canvas/`)
   - WebGPU and WebGL rendering backends
   - Canvas initialization and management
   - Render pipeline setup
   - Fallback mechanisms for browser compatibility

2. **Sprite System** (`internal/sprite/`)
   - Sprite sheet loading and management
   - Animation frame handling
   - Sprite rendering and positioning
   - Animation timing and control

3. **Type System** (`internal/types/`)
   - Core data structures (Vector2, UVRect, etc.)
   - Texture and pipeline interfaces
   - WebGPU-specific types

### Key Features

- **WebGPU Support**: Modern GPU-accelerated rendering
- **WebGL Fallback**: Automatic fallback for older browsers
- **Sprite Animation**: Frame-based sprite animation system
- **Cross-Platform**: Runs in any modern web browser
- **Go Performance**: Compiled to efficient WebAssembly

## Development Status

This project is currently in active development. See the implementation plans for detailed roadmap:

- [Sprite Implementation Plan](SPRITE_IMPLEMENTATION_PLAN.md)
- [Renderer Plan](RENDERER_PLAN.md)
- [Canvas Extension Plan](CANVAS_EXTENSION_PLAN.md)

## Building and Running

### Prerequisites

- Go 1.24.3 or later
- TinyGo (for WebAssembly compilation)
- Modern web browser with WebGPU support

### Build Commands

```bash
# Build WebAssembly binary
tinygo build -o main.wasm -target wasm cmd/webgpu-triangle/main.go

# Run development server (Python example)
python -m http.server 8000
```

### Development

```bash
# Run tests
go test ./...

# Run specific package tests
go test ./internal/canvas/
go test ./internal/sprite/
```

## Usage

1. Open `assets/index.html` in a web browser
2. The application will automatically initialize WebGPU or fall back to WebGL
3. View the animated triangle and sprite rendering demos

## Browser Compatibility

- **WebGPU**: Chrome 113+, Edge 113+
- **WebGL Fallback**: All modern browsers
- **WebAssembly**: All modern browsers

## Contributing

1. Follow Go's idiomatic project structure
2. Add tests for new functionality
3. Update documentation for API changes
4. Ensure cross-browser compatibility

## License

This project is open source. See individual files for specific licensing information.
