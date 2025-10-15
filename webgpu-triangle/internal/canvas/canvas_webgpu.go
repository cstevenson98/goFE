//go:build js

package canvas

import (
	"fmt"
	"syscall/js"
	"unsafe"

	"github.com/cogentcore/webgpu/wgpu"
	"github.com/conor/webgpu-triangle/internal/types"
)

// WebGPUCanvasManager implements CanvasManager using cogentcore/webgpu library
type WebGPUCanvasManager struct {
	// Canvas element
	canvas js.Value

	// WebGPU resources using cogentcore library
	instance *wgpu.Instance
	adapter  *wgpu.Adapter
	surface  *wgpu.Surface
	device   *wgpu.Device
	queue    *wgpu.Queue
	config   *wgpu.SurfaceConfiguration

	// Pipelines - available pipelines
	trianglePipeline *wgpu.RenderPipeline
	spritePipeline   *wgpu.RenderPipeline
	texturedPipeline *wgpu.RenderPipeline

	// Active pipelines - pipelines to execute in order
	activePipelines []types.PipelineType

	// Buffers
	vertexBuffer  *wgpu.Buffer
	uniformBuffer *wgpu.Buffer

	// Texture resources
	sampler          *wgpu.Sampler
	bindGroupLayout  *wgpu.BindGroupLayout
	loadedTextures   map[string]*wgpu.Texture
	currentTexture   *wgpu.Texture
	textureBindGroup *wgpu.BindGroup

	// Batch rendering
	batchMode         bool
	stagedVertices    []float32
	stagedVertexCount int

	// Status
	initialized bool
	error       string
}

// NewWebGPUCanvasManager creates a new WebGPU canvas manager
func NewWebGPUCanvasManager() *WebGPUCanvasManager {
	return &WebGPUCanvasManager{
		loadedTextures: make(map[string]*wgpu.Texture),
	}
}

// Initialize sets up the WebGPU canvas
func (w *WebGPUCanvasManager) Initialize(canvasID string) error {
	println("DEBUG: WebGPUCanvasManager.Initialize called for canvas:", canvasID)

	// Get the canvas element
	canvas := js.Global().Get("document").Call("getElementById", canvasID)
	if canvas.IsUndefined() || canvas.IsNull() {
		err := "Canvas element not found"
		println("ERROR:", err)
		return &CanvasError{Message: err}
	}

	w.canvas = canvas
	println("DEBUG: Canvas element found")

	// Get canvas dimensions (set by JavaScript)
	width := uint32(canvas.Get("width").Int())
	height := uint32(canvas.Get("height").Int())

	// Fallback to window size if canvas dimensions are 0
	if width == 0 || height == 0 {
		width = uint32(js.Global().Get("innerWidth").Int())
		height = uint32(js.Global().Get("innerHeight").Int())
		canvas.Set("width", width)
		canvas.Set("height", height)
	}

	println("DEBUG: Canvas size:", width, "x", height)

	// Create WebGPU instance
	w.instance = wgpu.CreateInstance(nil)
	if w.instance == nil {
		err := "WebGPU not supported"
		println("ERROR:", err)
		return &CanvasError{Message: err}
	}
	println("DEBUG: WebGPU instance created")

	// Create surface from canvas
	w.surface = w.instance.CreateSurface(&wgpu.SurfaceDescriptor{
		Canvas: canvas,
		Label:  "Main Canvas Surface",
	})
	println("DEBUG: Surface created from canvas")

	// Request adapter
	println("DEBUG: Requesting adapter...")
	adapter, err := w.instance.RequestAdapter(&wgpu.RequestAdapterOptions{
		CompatibleSurface: w.surface,
	})
	if err != nil {
		errMsg := fmt.Sprintf("Failed to request adapter: %v", err)
		println("ERROR:", errMsg)
		return &CanvasError{Message: errMsg}
	}
	w.adapter = adapter
	println("DEBUG: Adapter obtained")

	// Request device
	println("DEBUG: Requesting device...")
	device, err := adapter.RequestDevice(nil)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to request device: %v", err)
		println("ERROR:", errMsg)
		return &CanvasError{Message: errMsg}
	}
	w.device = device
	w.queue = device.GetQueue()
	println("DEBUG: Device and queue obtained")

	// Configure surface with actual canvas dimensions
	caps := w.surface.GetCapabilities(w.adapter)
	w.config = &wgpu.SurfaceConfiguration{
		Usage:       wgpu.TextureUsageRenderAttachment,
		Format:      caps.Formats[0],
		Width:       width,
		Height:      height,
		PresentMode: wgpu.PresentModeFifo,
		AlphaMode:   caps.AlphaModes[0],
	}
	w.surface.Configure(w.adapter, w.device, w.config)
	println("DEBUG: Surface configured")

	// Create pipelines
	if err := w.createTrianglePipeline(); err != nil {
		errMsg := fmt.Sprintf("Failed to create triangle pipeline: %v", err)
		println("ERROR:", errMsg)
		return &CanvasError{Message: errMsg}
	}

	if err := w.createSpritePipeline(); err != nil {
		errMsg := fmt.Sprintf("Failed to create sprite pipeline: %v", err)
		println("ERROR:", errMsg)
		return &CanvasError{Message: errMsg}
	}

	if err := w.createTexturedPipeline(); err != nil {
		errMsg := fmt.Sprintf("Failed to create textured pipeline: %v", err)
		println("ERROR:", errMsg)
		return &CanvasError{Message: errMsg}
	}

	// Create sampler
	if err := w.createSampler(); err != nil {
		errMsg := fmt.Sprintf("Failed to create sampler: %v", err)
		println("ERROR:", errMsg)
		return &CanvasError{Message: errMsg}
	}

	// Create vertex buffer
	if err := w.createSpriteVertexBuffer(); err != nil {
		errMsg := fmt.Sprintf("Failed to create vertex buffer: %v", err)
		println("ERROR:", errMsg)
		return &CanvasError{Message: errMsg}
	}

	println("DEBUG: WebGPU setup complete")

	w.initialized = true
	return nil
}

// SetPipelines sets the active pipelines to be executed in order
// This method handles graceful switching between pipeline configurations
func (w *WebGPUCanvasManager) SetPipelines(pipelines []types.PipelineType) error {
	if !w.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	// Validate that all requested pipelines are available
	for _, pipelineType := range pipelines {
		switch pipelineType {
		case types.TrianglePipeline:
			if w.trianglePipeline == nil {
				return &CanvasError{Message: "Triangle pipeline not available"}
			}
		case types.SpritePipeline:
			if w.spritePipeline == nil {
				return &CanvasError{Message: "Sprite pipeline not available"}
			}
		case types.TexturedPipeline:
			if w.texturedPipeline == nil {
				return &CanvasError{Message: "Textured pipeline not available"}
			}
		default:
			return &CanvasError{Message: fmt.Sprintf("Unknown pipeline type: %v", pipelineType)}
		}
	}

	// Clear any staged vertices from previous pipeline configuration
	w.stagedVertexCount = 0
	w.stagedVertices = nil

	// Set the new active pipelines
	w.activePipelines = make([]types.PipelineType, len(pipelines))
	copy(w.activePipelines, pipelines)

	println("DEBUG: Active pipelines set:", pipelines)
	return nil
}

// createTrianglePipeline creates the basic triangle rendering pipeline
func (w *WebGPUCanvasManager) createTrianglePipeline() error {
	shaderCode := `
@vertex
fn vs_main(@builtin(vertex_index) vertexIndex: u32) -> @builtin(position) vec4f {
	var pos = array<vec2f, 3>(
		vec2f( 0.0,  0.5),
		vec2f(-0.5, -0.5),
		vec2f( 0.5, -0.5)
	);
	return vec4f(pos[vertexIndex], 0.0, 1.0);
}

@fragment
fn fs_main() -> @location(0) vec4f {
	return vec4f(1.0, 0.0, 0.0, 1.0);
}
`

	shaderModule, err := w.device.CreateShaderModule(&wgpu.ShaderModuleDescriptor{
		Label:          "Triangle Shader",
		WGSLDescriptor: &wgpu.ShaderModuleWGSLDescriptor{Code: shaderCode},
	})
	if err != nil {
		return err
	}
	defer shaderModule.Release()

	pipeline, err := w.device.CreateRenderPipeline(&wgpu.RenderPipelineDescriptor{
		Label: "Triangle Pipeline",
		Vertex: wgpu.VertexState{
			Module:     shaderModule,
			EntryPoint: "vs_main",
		},
		Primitive: wgpu.PrimitiveState{
			Topology:  wgpu.PrimitiveTopologyTriangleList,
			FrontFace: wgpu.FrontFaceCCW,
			CullMode:  wgpu.CullModeNone,
		},
		Multisample: wgpu.MultisampleState{
			Count:                  1,
			Mask:                   0xFFFFFFFF,
			AlphaToCoverageEnabled: false,
		},
		Fragment: &wgpu.FragmentState{
			Module:     shaderModule,
			EntryPoint: "fs_main",
			Targets: []wgpu.ColorTargetState{
				{
					Format:    w.config.Format,
					Blend:     &wgpu.BlendStateReplace,
					WriteMask: wgpu.ColorWriteMaskAll,
				},
			},
		},
	})
	if err != nil {
		return err
	}

	w.trianglePipeline = pipeline
	println("DEBUG: Triangle pipeline created")
	return nil
}

// createSpritePipeline creates the colored sprite rendering pipeline
func (w *WebGPUCanvasManager) createSpritePipeline() error {
	shaderCode := `
struct VertexOutput {
	@builtin(position) position: vec4f,
	@location(0) color: vec4f,
}

@vertex
fn vs_main(
	@location(0) position: vec2f,
	@location(1) color: vec4f
) -> VertexOutput {
	var output: VertexOutput;
	output.position = vec4f(position, 0.0, 1.0);
	output.color = color;
	return output;
}

@fragment
fn fs_main(@location(0) color: vec4f) -> @location(0) vec4f {
	return color;
}
`

	shaderModule, err := w.device.CreateShaderModule(&wgpu.ShaderModuleDescriptor{
		Label:          "Sprite Shader",
		WGSLDescriptor: &wgpu.ShaderModuleWGSLDescriptor{Code: shaderCode},
	})
	if err != nil {
		return err
	}
	defer shaderModule.Release()

	pipeline, err := w.device.CreateRenderPipeline(&wgpu.RenderPipelineDescriptor{
		Label: "Sprite Pipeline",
		Vertex: wgpu.VertexState{
			Module:     shaderModule,
			EntryPoint: "vs_main",
			Buffers: []wgpu.VertexBufferLayout{
				{
					ArrayStride: 24, // 6 floats * 4 bytes
					Attributes: []wgpu.VertexAttribute{
						{
							ShaderLocation: 0,
							Offset:         0,
							Format:         wgpu.VertexFormatFloat32x2,
						},
						{
							ShaderLocation: 1,
							Offset:         8,
							Format:         wgpu.VertexFormatFloat32x4,
						},
					},
				},
			},
		},
		Primitive: wgpu.PrimitiveState{
			Topology:  wgpu.PrimitiveTopologyTriangleList,
			FrontFace: wgpu.FrontFaceCCW,
			CullMode:  wgpu.CullModeNone,
		},
		Multisample: wgpu.MultisampleState{
			Count:                  1,
			Mask:                   0xFFFFFFFF,
			AlphaToCoverageEnabled: false,
		},
		Fragment: &wgpu.FragmentState{
			Module:     shaderModule,
			EntryPoint: "fs_main",
			Targets: []wgpu.ColorTargetState{
				{
					Format: w.config.Format,
					Blend: &wgpu.BlendState{
						Color: wgpu.BlendComponent{
							SrcFactor: wgpu.BlendFactorSrcAlpha,
							DstFactor: wgpu.BlendFactorOneMinusSrcAlpha,
							Operation: wgpu.BlendOperationAdd,
						},
						Alpha: wgpu.BlendComponent{
							SrcFactor: wgpu.BlendFactorOne,
							DstFactor: wgpu.BlendFactorOneMinusSrcAlpha,
							Operation: wgpu.BlendOperationAdd,
						},
					},
					WriteMask: wgpu.ColorWriteMaskAll,
				},
			},
		},
	})
	if err != nil {
		return err
	}

	w.spritePipeline = pipeline
	println("DEBUG: Sprite pipeline created")
	return nil
}

// createTexturedPipeline creates the textured sprite rendering pipeline
func (w *WebGPUCanvasManager) createTexturedPipeline() error {
	shaderCode := `
struct VertexOutput {
	@builtin(position) position: vec4f,
	@location(0) uv: vec2f,
}

@vertex
fn vs_main(
	@location(0) position: vec2f,
	@location(1) uv: vec2f
) -> VertexOutput {
	var output: VertexOutput;
	output.position = vec4f(position, 0.0, 1.0);
	output.uv = uv;
	return output;
}

@group(0) @binding(0) var textureSampler: sampler;
@group(0) @binding(1) var textureData: texture_2d<f32>;

@fragment
fn fs_main(@location(0) uv: vec2f) -> @location(0) vec4f {
	return textureSample(textureData, textureSampler, uv);
}
`

	shaderModule, err := w.device.CreateShaderModule(&wgpu.ShaderModuleDescriptor{
		Label:          "Textured Sprite Shader",
		WGSLDescriptor: &wgpu.ShaderModuleWGSLDescriptor{Code: shaderCode},
	})
	if err != nil {
		return err
	}
	defer shaderModule.Release()

	// Create bind group layout
	bindGroupLayout, err := w.device.CreateBindGroupLayout(&wgpu.BindGroupLayoutDescriptor{
		Label: "Texture Bind Group Layout",
		Entries: []wgpu.BindGroupLayoutEntry{
			{
				Binding:    0,
				Visibility: wgpu.ShaderStageFragment,
				Sampler: wgpu.SamplerBindingLayout{
					Type: wgpu.SamplerBindingTypeFiltering,
				},
			},
			{
				Binding:    1,
				Visibility: wgpu.ShaderStageFragment,
				Texture: wgpu.TextureBindingLayout{
					SampleType:    wgpu.TextureSampleTypeFloat,
					ViewDimension: wgpu.TextureViewDimension2D,
				},
			},
		},
	})
	if err != nil {
		return err
	}
	w.bindGroupLayout = bindGroupLayout

	// Create pipeline layout
	pipelineLayout, err := w.device.CreatePipelineLayout(&wgpu.PipelineLayoutDescriptor{
		Label:            "Textured Pipeline Layout",
		BindGroupLayouts: []*wgpu.BindGroupLayout{bindGroupLayout},
	})
	if err != nil {
		return err
	}
	defer pipelineLayout.Release()

	pipeline, err := w.device.CreateRenderPipeline(&wgpu.RenderPipelineDescriptor{
		Label:  "Textured Pipeline",
		Layout: pipelineLayout,
		Vertex: wgpu.VertexState{
			Module:     shaderModule,
			EntryPoint: "vs_main",
			Buffers: []wgpu.VertexBufferLayout{
				{
					ArrayStride: 16, // 4 floats * 4 bytes
					Attributes: []wgpu.VertexAttribute{
						{
							ShaderLocation: 0,
							Offset:         0,
							Format:         wgpu.VertexFormatFloat32x2,
						},
						{
							ShaderLocation: 1,
							Offset:         8,
							Format:         wgpu.VertexFormatFloat32x2,
						},
					},
				},
			},
		},
		Primitive: wgpu.PrimitiveState{
			Topology:  wgpu.PrimitiveTopologyTriangleList,
			FrontFace: wgpu.FrontFaceCCW,
			CullMode:  wgpu.CullModeNone,
		},
		Multisample: wgpu.MultisampleState{
			Count:                  1,
			Mask:                   0xFFFFFFFF,
			AlphaToCoverageEnabled: false,
		},
		Fragment: &wgpu.FragmentState{
			Module:     shaderModule,
			EntryPoint: "fs_main",
			Targets: []wgpu.ColorTargetState{
				{
					Format: w.config.Format,
					Blend: &wgpu.BlendState{
						Color: wgpu.BlendComponent{
							SrcFactor: wgpu.BlendFactorSrcAlpha,
							DstFactor: wgpu.BlendFactorOneMinusSrcAlpha,
							Operation: wgpu.BlendOperationAdd,
						},
						Alpha: wgpu.BlendComponent{
							SrcFactor: wgpu.BlendFactorOne,
							DstFactor: wgpu.BlendFactorOneMinusSrcAlpha,
							Operation: wgpu.BlendOperationAdd,
						},
					},
					WriteMask: wgpu.ColorWriteMaskAll,
				},
			},
		},
	})
	if err != nil {
		return err
	}

	w.texturedPipeline = pipeline
	println("DEBUG: Textured pipeline created")
	return nil
}

// createSampler creates a texture sampler
func (w *WebGPUCanvasManager) createSampler() error {
	sampler, err := w.device.CreateSampler(&wgpu.SamplerDescriptor{
		Label:         "Texture Sampler",
		AddressModeU:  wgpu.AddressModeClampToEdge,
		AddressModeV:  wgpu.AddressModeClampToEdge,
		AddressModeW:  wgpu.AddressModeClampToEdge,
		MagFilter:     wgpu.FilterModeLinear,
		MinFilter:     wgpu.FilterModeLinear,
		MipmapFilter:  wgpu.MipmapFilterModeLinear,
		LodMinClamp:   0,
		LodMaxClamp:   32,
		MaxAnisotropy: 1,
	})
	if err != nil {
		return err
	}

	w.sampler = sampler
	println("DEBUG: Sampler created")
	return nil
}

// createSpriteVertexBuffer creates a dynamic vertex buffer for sprite rendering
func (w *WebGPUCanvasManager) createSpriteVertexBuffer() error {
	bufferSize := uint64(1024 * 24) // 1024 vertices * 24 bytes per vertex

	vertexBuffer, err := w.device.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "Sprite Vertex Buffer",
		Size:  bufferSize,
		Usage: wgpu.BufferUsageVertex | wgpu.BufferUsageCopyDst,
	})
	if err != nil {
		return err
	}

	w.vertexBuffer = vertexBuffer
	println("DEBUG: Vertex buffer created, size:", bufferSize)
	return nil
}

// Render draws the current frame
func (w *WebGPUCanvasManager) Render() error {
	if !w.initialized {
		return nil
	}

	return w.renderFrame()
}

// renderFrame performs the actual rendering
func (w *WebGPUCanvasManager) renderFrame() error {
	// Get current texture
	nextTexture, err := w.surface.GetCurrentTexture()
	if err != nil {
		return err
	}

	view, err := nextTexture.CreateView(nil)
	if err != nil {
		return err
	}
	defer view.Release()

	// Create command encoder
	encoder, err := w.device.CreateCommandEncoder(&wgpu.CommandEncoderDescriptor{
		Label: "Command Encoder",
	})
	if err != nil {
		return err
	}
	defer encoder.Release()

	// Begin render pass
	renderPass := encoder.BeginRenderPass(&wgpu.RenderPassDescriptor{
		Label: "Main Render Pass",
		ColorAttachments: []wgpu.RenderPassColorAttachment{
			{
				View:       view,
				LoadOp:     wgpu.LoadOpClear,
				StoreOp:    wgpu.StoreOpStore,
				ClearValue: wgpu.Color{R: 0.0, G: 0.0, B: 0.0, A: 1.0},
			},
		},
	})

	// Execute active pipelines in order
	for _, pipelineType := range w.activePipelines {
		w.executePipeline(renderPass, pipelineType)
	}

	renderPass.End()
	renderPass.Release()

	// Submit command buffer
	cmdBuffer, err := encoder.Finish(nil)
	if err != nil {
		return err
	}
	defer cmdBuffer.Release()

	w.queue.Submit(cmdBuffer)
	w.surface.Present()

	// Clear staged vertices after rendering
	w.stagedVertexCount = 0

	return nil
}

// canvasToNDC converts canvas coordinates to Normalized Device Coordinates
func (w *WebGPUCanvasManager) canvasToNDC(x, y float64) (float32, float32) {
	width := w.canvas.Get("width").Float()
	height := w.canvas.Get("height").Float()

	ndcX := float32((x/width)*2.0 - 1.0)
	ndcY := float32(1.0 - (y/height)*2.0)

	return ndcX, ndcY
}

// executePipeline executes a specific pipeline type during rendering
func (w *WebGPUCanvasManager) executePipeline(renderPass *wgpu.RenderPassEncoder, pipelineType types.PipelineType) {
	switch pipelineType {
	case types.TrianglePipeline:
		if w.trianglePipeline != nil {
			renderPass.SetPipeline(w.trianglePipeline)
			renderPass.Draw(3, 1, 0, 0)
		}
	case types.SpritePipeline:
		if w.spritePipeline != nil && w.stagedVertexCount > 0 {
			renderPass.SetPipeline(w.spritePipeline)
			renderPass.SetVertexBuffer(0, w.vertexBuffer, 0, wgpu.WholeSize)
			renderPass.Draw(uint32(w.stagedVertexCount), 1, 0, 0)
		}
	case types.TexturedPipeline:
		if w.texturedPipeline != nil && w.stagedVertexCount > 0 && w.currentTexture != nil && w.textureBindGroup != nil {
			renderPass.SetPipeline(w.texturedPipeline)
			renderPass.SetBindGroup(0, w.textureBindGroup, nil)
			renderPass.SetVertexBuffer(0, w.vertexBuffer, 0, wgpu.WholeSize)
			renderPass.Draw(uint32(w.stagedVertexCount), 1, 0, 0)
		}
	}
}

// DrawColoredRect draws a colored rectangle
func (w *WebGPUCanvasManager) DrawColoredRect(position types.Vector2, size types.Vector2, color [4]float32) error {
	if !w.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	vertices := w.generateQuadVertices(position, size, color)

	if w.batchMode {
		w.stagedVertices = append(w.stagedVertices, vertices...)
		println("DEBUG: Batched rectangle at", position.X, position.Y)
	} else {
		// Immediate mode - upload and stage
		w.queue.WriteBuffer(w.vertexBuffer, 0, float32SliceToBytes(vertices))
		w.stagedVertexCount = len(vertices) / 6 // 6 floats per vertex
		println("DEBUG: Immediate mode - Staged", w.stagedVertexCount, "vertices")
	}

	return nil
}

// generateQuadVertices generates vertices for a colored rectangle
func (w *WebGPUCanvasManager) generateQuadVertices(pos types.Vector2, size types.Vector2, color [4]float32) []float32 {
	x0 := pos.X
	y0 := pos.Y
	x1 := pos.X + size.X
	y1 := pos.Y + size.Y

	ndcX0, ndcY0 := w.canvasToNDC(x0, y0)
	ndcX1, ndcY1 := w.canvasToNDC(x1, y1)

	return []float32{
		// Triangle 1
		ndcX0, ndcY0, color[0], color[1], color[2], color[3],
		ndcX1, ndcY0, color[0], color[1], color[2], color[3],
		ndcX0, ndcY1, color[0], color[1], color[2], color[3],
		// Triangle 2
		ndcX1, ndcY0, color[0], color[1], color[2], color[3],
		ndcX1, ndcY1, color[0], color[1], color[2], color[3],
		ndcX0, ndcY1, color[0], color[1], color[2], color[3],
	}
}

// LoadTexture loads a PNG texture from a URL
func (w *WebGPUCanvasManager) LoadTexture(path string) error {
	if !w.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	if _, exists := w.loadedTextures[path]; exists {
		println("DEBUG: Texture already loaded:", path)
		return nil
	}

	println("DEBUG: Loading texture from:", path)

	image := js.Global().Get("Image").New()
	imageLoaded := make(chan bool)

	image.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		println("DEBUG: Image loaded successfully:", path)
		gpuTexture := w.uploadTextureToGPU(image)
		w.loadedTextures[path] = gpuTexture
		imageLoaded <- true
		return nil
	}))

	image.Set("onerror", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		println("ERROR: Failed to load image:", path)
		imageLoaded <- false
		return nil
	}))

	image.Set("src", path)

	go func() {
		<-imageLoaded
	}()

	return nil
}

// uploadTextureToGPU uploads an image to GPU memory
func (w *WebGPUCanvasManager) uploadTextureToGPU(image js.Value) *wgpu.Texture {
	width := uint32(image.Get("width").Int())
	height := uint32(image.Get("height").Int())

	println("DEBUG: Uploading texture to GPU - Size:", width, "x", height)

	texture, err := w.device.CreateTexture(&wgpu.TextureDescriptor{
		Label: "Loaded Texture",
		Size: wgpu.Extent3D{
			Width:              width,
			Height:             height,
			DepthOrArrayLayers: 1,
		},
		MipLevelCount: 1,
		SampleCount:   1,
		Dimension:     wgpu.TextureDimension2D,
		Format:        wgpu.TextureFormatRGBA8Unorm,
		Usage:         wgpu.TextureUsageTextureBinding | wgpu.TextureUsageCopyDst | wgpu.TextureUsageRenderAttachment,
	})
	if err != nil {
		println("ERROR: Failed to create texture:", err)
		return nil
	}

	// Create a canvas to get pixel data
	tempCanvas := js.Global().Get("document").Call("createElement", "canvas")
	tempCanvas.Set("width", width)
	tempCanvas.Set("height", height)
	ctx := tempCanvas.Call("getContext", "2d")
	ctx.Call("drawImage", image, 0, 0)

	imageData := ctx.Call("getImageData", 0, 0, width, height)
	jsData := imageData.Get("data")

	// Convert JS Uint8ClampedArray to Go bytes
	dataLen := jsData.Get("length").Int()
	data := make([]byte, dataLen)
	js.CopyBytesToGo(data, jsData)

	// Write texture data
	w.queue.WriteTexture(
		&wgpu.ImageCopyTexture{
			Texture: texture,
		},
		data,
		&wgpu.TextureDataLayout{
			BytesPerRow:  width * 4,
			RowsPerImage: height,
		},
		&wgpu.Extent3D{
			Width:              width,
			Height:             height,
			DepthOrArrayLayers: 1,
		},
	)

	println("DEBUG: Texture uploaded to GPU successfully")
	return texture
}

// DrawTexturedRect draws a textured rectangle
func (w *WebGPUCanvasManager) DrawTexturedRect(texturePath string, position types.Vector2, size types.Vector2, uv types.UVRect) error {
	if !w.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	gpuTexture, exists := w.loadedTextures[texturePath]
	if !exists {
		println("DEBUG: Texture not loaded:", texturePath)
		return &CanvasError{Message: "Texture not loaded: " + texturePath}
	}

	vertices := w.generateTexturedQuadVertices(position, size, uv)

	// Create bind group for this texture
	w.textureBindGroup = w.createTextureBindGroup(gpuTexture)
	w.currentTexture = gpuTexture

	if w.batchMode {
		// Batch mode - accumulate vertices
		w.stagedVertices = append(w.stagedVertices, vertices...)
		println("DEBUG: Batched textured rect at", position.X, position.Y)
	} else {
		// Immediate mode - upload and stage
		w.queue.WriteBuffer(w.vertexBuffer, 0, float32SliceToBytes(vertices))
		w.stagedVertexCount = len(vertices) / 4 // 4 floats per vertex
		println("DEBUG: Immediate mode - Drew textured rect -", w.stagedVertexCount, "vertices")
	}

	return nil
}

// generateTexturedQuadVertices generates vertices for a textured rectangle
func (w *WebGPUCanvasManager) generateTexturedQuadVertices(pos types.Vector2, size types.Vector2, uv types.UVRect) []float32 {
	x0 := pos.X
	y0 := pos.Y
	x1 := pos.X + size.X
	y1 := pos.Y + size.Y

	ndcX0, ndcY0 := w.canvasToNDC(x0, y0)
	ndcX1, ndcY1 := w.canvasToNDC(x1, y1)

	u0 := float32(uv.U)
	v0 := float32(uv.V)
	u1 := float32(uv.U + uv.W)
	v1 := float32(uv.V + uv.H)

	return []float32{
		// Triangle 1
		ndcX0, ndcY0, u0, v0,
		ndcX1, ndcY0, u1, v0,
		ndcX0, ndcY1, u0, v1,
		// Triangle 2
		ndcX1, ndcY0, u1, v0,
		ndcX1, ndcY1, u1, v1,
		ndcX0, ndcY1, u0, v1,
	}
}

// createTextureBindGroup creates a bind group for a specific texture
func (w *WebGPUCanvasManager) createTextureBindGroup(texture *wgpu.Texture) *wgpu.BindGroup {
	textureView, err := texture.CreateView(nil)
	if err != nil {
		println("ERROR: Failed to create texture view:", err)
		return nil
	}

	bindGroup, err := w.device.CreateBindGroup(&wgpu.BindGroupDescriptor{
		Label:  "Texture Bind Group",
		Layout: w.bindGroupLayout,
		Entries: []wgpu.BindGroupEntry{
			{
				Binding: 0,
				Sampler: w.sampler,
			},
			{
				Binding:     1,
				TextureView: textureView,
			},
		},
	})
	if err != nil {
		println("ERROR: Failed to create bind group:", err)
		return nil
	}

	return bindGroup
}

// BeginBatch starts batch rendering mode
func (w *WebGPUCanvasManager) BeginBatch() error {
	if !w.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	w.batchMode = true
	w.stagedVertices = make([]float32, 0)
	w.stagedVertexCount = 0

	println("DEBUG: BeginBatch - Batch mode enabled")
	return nil
}

// EndBatch ends batch rendering mode and uploads all batched vertices
func (w *WebGPUCanvasManager) EndBatch() error {
	if !w.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	if !w.batchMode {
		println("DEBUG: EndBatch called but not in batch mode")
		return nil
	}

	err := w.FlushBatch()
	if err != nil {
		return err
	}

	w.batchMode = false
	println("DEBUG: EndBatch - Batch mode disabled,", w.stagedVertexCount, "vertices uploaded")

	return nil
}

// FlushBatch uploads accumulated vertices to GPU
func (w *WebGPUCanvasManager) FlushBatch() error {
	if !w.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}

	if len(w.stagedVertices) == 0 {
		println("DEBUG: FlushBatch - No vertices to flush")
		w.stagedVertexCount = 0
		return nil
	}

	w.queue.WriteBuffer(w.vertexBuffer, 0, float32SliceToBytes(w.stagedVertices))

	// Calculate vertex count based on vertex format
	// Textured vertices: 4 floats per vertex (position.xy + uv.xy)
	// Colored vertices: 6 floats per vertex (position.xy + color.rgba)
	// For now, we'll assume textured format if we have a currentTexture
	if w.currentTexture != nil {
		w.stagedVertexCount = len(w.stagedVertices) / 4
	} else {
		w.stagedVertexCount = len(w.stagedVertices) / 6
	}

	println("DEBUG: FlushBatch - Uploaded", len(w.stagedVertices), "floats (", w.stagedVertexCount, "vertices )")

	return nil
}

// Cleanup releases resources
func (w *WebGPUCanvasManager) Cleanup() error {
	if w.trianglePipeline != nil {
		w.trianglePipeline.Release()
	}
	if w.spritePipeline != nil {
		w.spritePipeline.Release()
	}
	if w.texturedPipeline != nil {
		w.texturedPipeline.Release()
	}
	if w.vertexBuffer != nil {
		w.vertexBuffer.Release()
	}
	if w.sampler != nil {
		w.sampler.Release()
	}
	if w.bindGroupLayout != nil {
		w.bindGroupLayout.Release()
	}
	if w.queue != nil {
		w.queue.Release()
	}
	if w.device != nil {
		w.device.Release()
	}
	if w.surface != nil {
		w.surface.Release()
	}
	if w.instance != nil {
		w.instance.Release()
	}

	w.SetStatus(false, "Cleaned up")
	return nil
}

// GetStatus returns the current status
func (w *WebGPUCanvasManager) GetStatus() (bool, string) {
	return w.initialized, w.error
}

// SetStatus updates the status (simplified - no UI elements)
func (w *WebGPUCanvasManager) SetStatus(initialized bool, message string) {
	w.initialized = initialized
	w.error = message
	println("STATUS:", message)
}

// Stub implementations for interface compliance
func (w *WebGPUCanvasManager) DrawTexture(texture types.Texture, position types.Vector2, size types.Vector2, uv types.UVRect) error {
	color := [4]float32{0.0, 0.5, 1.0, 1.0}
	return w.DrawColoredRect(position, size, color)
}

func (w *WebGPUCanvasManager) DrawTextureRotated(texture types.Texture, position types.Vector2, size types.Vector2, uv types.UVRect, rotation float64) error {
	println("DEBUG: DrawTextureRotated STUB")
	return nil
}

func (w *WebGPUCanvasManager) DrawTextureScaled(texture types.Texture, position types.Vector2, size types.Vector2, uv types.UVRect, scale types.Vector2) error {
	println("DEBUG: DrawTextureScaled STUB")
	return nil
}

func (w *WebGPUCanvasManager) GetSpritePipeline() types.Pipeline {
	return &types.WebGPUPipeline{Valid: w.spritePipeline != nil}
}

func (w *WebGPUCanvasManager) GetBackgroundPipeline() types.Pipeline {
	return &types.WebGPUPipeline{Valid: false}
}

func (w *WebGPUCanvasManager) ClearCanvas() error {
	if !w.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}
	return nil
}

// Helper functions
func float32SliceToBytes(f []float32) []byte {
	bytes := make([]byte, len(f)*4)
	for i, v := range f {
		// Convert float32 to uint32 bits
		bits := *(*uint32)(unsafe.Pointer(&v))
		bytes[i*4] = byte(bits)
		bytes[i*4+1] = byte(bits >> 8)
		bytes[i*4+2] = byte(bits >> 16)
		bytes[i*4+3] = byte(bits >> 24)
	}
	return bytes
}
