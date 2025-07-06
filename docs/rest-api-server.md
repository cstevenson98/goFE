# REST API Server Interface for Music Composer Assistant

## Overview
The REST API server provides a clean interface to the core business logic through a collection of handlers. Each handler encapsulates specific functionality and provides HTTP endpoints for client interaction.

## Architecture

### Core Server Structure
```go
type Server struct {
    router        *mux.Router
    handlers      map[string]Handler
    middleware    []Middleware
    config        *ServerConfig
}

type ServerConfig struct {
    Port          string
    Host          string
    ReadTimeout   time.Duration
    WriteTimeout  time.Duration
    MaxHeaderSize int
    CORSEnabled   bool
    Logging       bool
}

type Handler interface {
    RegisterRoutes(router *mux.Router)
    GetName() string
}
```

### Handler Interface
```go
type Handler interface {
    RegisterRoutes(router *mux.Router)
    GetName() string
    GetVersion() string
    HealthCheck() error
}

type BaseHandler struct {
    name    string
    version string
    logger  *log.Logger
}

func (bh *BaseHandler) GetName() string {
    return bh.name
}

func (bh *BaseHandler) GetVersion() string {
    return bh.version
}

func (bh *BaseHandler) HealthCheck() error {
    return nil
}
```

## Handler Implementations

### 1. Prompt Handler
```go
type PromptHandler struct {
    BaseHandler
    promptEngine *PromptEngine
}

func NewPromptHandler(promptEngine *PromptEngine) *PromptHandler {
    return &PromptHandler{
        BaseHandler: BaseHandler{
            name:    "prompt",
            version: "1.0.0",
            logger:  log.New(os.Stdout, "[PROMPT] ", log.LstdFlags),
        },
        promptEngine: promptEngine,
    }
}

func (ph *PromptHandler) RegisterRoutes(router *mux.Router) {
    promptRouter := router.PathPrefix("/api/prompt").Subrouter()
    
    promptRouter.HandleFunc("/templates", ph.ListTemplates).Methods("GET")
    promptRouter.HandleFunc("/templates/{name}", ph.GetTemplate).Methods("GET")
}


```

### 2. Diff Handler
```go
type DiffHandler struct {
    BaseHandler
    diffGenerator *DiffGenerator
}

func NewDiffHandler(diffGenerator *DiffGenerator) *DiffHandler {
    return &DiffHandler{
        BaseHandler: BaseHandler{
            name:    "diff",
            version: "1.0.0",
            logger:  log.New(os.Stdout, "[DIFF] ", log.LstdFlags),
        },
        diffGenerator: diffGenerator,
    }
}

func (dh *DiffHandler) RegisterRoutes(router *mux.Router) {
    diffRouter := router.PathPrefix("/api/diff").Subrouter()
    
    diffRouter.HandleFunc("/get", dh.GetDiff).Methods("GET")
    diffRouter.HandleFunc("/render", dh.RenderDiff).Methods("POST")
    diffRouter.HandleFunc("/cache/clear", dh.ClearCache).Methods("POST")
}

// GET /api/diff/get
type GetDiffResponse struct {
    Success bool       `json:"success"`
    Data    *DiffResult `json:"data,omitempty"`
    Error   string     `json:"error,omitempty"`
}

func (dh *DiffHandler) GetDiff(w http.ResponseWriter, r *http.Request) {
    // Get diff from system state (e.g., from context manager)
    diffID := r.URL.Query().Get("id")
    if diffID == "" {
        dh.respondWithError(w, http.StatusBadRequest, "Missing diff ID")
        return
    }
    
    result, err := dh.diffGenerator.GetDiffFromState(diffID)
    if err != nil {
        dh.respondWithError(w, http.StatusNotFound, fmt.Sprintf("Diff not found: %v", err))
        return
    }
    
    dh.respondWithJSON(w, http.StatusOK, GetDiffResponse{
        Success: true,
        Data:    result,
    })
}
```

### 3. LilyPond Handler
```go
type LilyPondHandler struct {
    BaseHandler
    lilypondProcessor *LilyPondProcessor
}

func NewLilyPondHandler(lilypondProcessor *LilyPondProcessor) *LilyPondHandler {
    return &LilyPondHandler{
        BaseHandler: BaseHandler{
            name:    "lilypond",
            version: "1.0.0",
            logger:  log.New(os.Stdout, "[LILYPOND] ", log.LstdFlags),
        },
        lilypondProcessor: lilypondProcessor,
    }
}

func (lph *LilyPondHandler) RegisterRoutes(router *mux.Router) {
    lilypondRouter := router.PathPrefix("/api/lilypond").Subrouter()
    
    lilypondRouter.HandleFunc("/compile", lph.CompileDocument).Methods("POST")
    lilypondRouter.HandleFunc("/validate", lph.ValidateSyntax).Methods("POST")
    lilypondRouter.HandleFunc("/pdf/{id}", lph.GetPDF).Methods("GET")
    lilypondRouter.HandleFunc("/health", lph.HealthCheck).Methods("GET")
}

// POST /api/lilypond/compile
type CompileDocumentRequest struct {
    LilyPond string         `json:"lilypond"`
    Options  CompileOptions `json:"options,omitempty"`
}

type CompileDocumentResponse struct {
    Success bool          `json:"success"`
    Data    *CompileResult `json:"data,omitempty"`
    Error   string        `json:"error,omitempty"`
}

func (lph *LilyPondHandler) CompileDocument(w http.ResponseWriter, r *http.Request) {
    var req CompileDocumentRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        lph.respondWithError(w, http.StatusBadRequest, "Invalid request body")
        return
    }
    
    result, err := lph.lilypondProcessor.CompileToPDF(req.LilyPond)
    if err != nil {
        lph.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Compilation failed: %v", err))
        return
    }
    
    lph.respondWithJSON(w, http.StatusOK, CompileDocumentResponse{
        Success: result.Success,
        Data:    result,
    })
}

// POST /api/lilypond/validate
type ValidateSyntaxRequest struct {
    LilyPond string `json:"lilypond"`
}

type ValidateSyntaxResponse struct {
    Success bool              `json:"success"`
    Data    *ValidationResult `json:"data,omitempty"`
    Error   string            `json:"error,omitempty"`
}

func (lph *LilyPondHandler) ValidateSyntax(w http.ResponseWriter, r *http.Request) {
    var req ValidateSyntaxRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        lph.respondWithError(w, http.StatusBadRequest, "Invalid request body")
        return
    }
    
    result := lph.lilypondProcessor.ValidateSyntax(req.LilyPond)
    
    lph.respondWithJSON(w, http.StatusOK, ValidateSyntaxResponse{
        Success: result.IsValid,
        Data:    result,
    })
}

// GET /api/lilypond/pdf/{id}
func (lph *LilyPondHandler) GetPDF(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    
    // Get PDF from cache/storage
    pdfData, err := lph.lilypondProcessor.GetPDF(id)
    if err != nil {
        lph.respondWithError(w, http.StatusNotFound, "PDF not found")
        return
    }
    
    w.Header().Set("Content-Type", "application/pdf")
    w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%s.pdf", id))
    w.Write(pdfData)
}
```

### 4. Context Handler
```go
type ContextHandler struct {
    BaseHandler
    contextManager *ContextManager
}

func NewContextHandler(contextManager *ContextManager) *ContextHandler {
    return &ContextHandler{
        BaseHandler: BaseHandler{
            name:    "context",
            version: "1.0.0",
            logger:  log.New(os.Stdout, "[CONTEXT] ", log.LstdFlags),
        },
        contextManager: contextManager,
    }
}

func (ch *ContextHandler) RegisterRoutes(router *mux.Router) {
    contextRouter := router.PathPrefix("/api/context").Subrouter()
    
    contextRouter.HandleFunc("/document", ch.GetDocument).Methods("GET")
    contextRouter.HandleFunc("/save", ch.SaveCurrent).Methods("POST")
}

// GET /api/context/document
type GetDocumentResponse struct {
    Success bool        `json:"success"`
    Data    *Document   `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

func (ch *ContextHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
    // Get the current document from the context manager
    document, err := ch.contextManager.GetCurrentDocument()
    if err != nil {
        ch.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get document: %v", err))
        return
    }
    
    ch.respondWithJSON(w, http.StatusOK, GetDocumentResponse{
        Success: true,
        Data:    document,
    })
}

// POST /api/context/save
type SaveCurrentRequest struct {
    Document *Document `json:"document"`
}

type SaveCurrentResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message,omitempty"`
    Error   string `json:"error,omitempty"`
}

func (ch *ContextHandler) SaveCurrent(w http.ResponseWriter, r *http.Request) {
    var req SaveCurrentRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        ch.respondWithError(w, http.StatusBadRequest, "Invalid request body")
        return
    }
    
    if req.Document == nil {
        ch.respondWithError(w, http.StatusBadRequest, "Document is required")
        return
    }
    
    // Save the document to the context manager
    err := ch.contextManager.SaveCurrentDocument(req.Document)
    if err != nil {
        ch.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to save document: %v", err))
        return
    }
    
    ch.respondWithJSON(w, http.StatusOK, SaveCurrentResponse{
        Success: true,
        Message: "Document saved successfully",
    })
}
```

## Server Implementation

### Main Server Setup
```go
type Server struct {
    router        *mux.Router
    handlers      map[string]Handler
    middleware    []Middleware
    config        *ServerConfig
    logger        *log.Logger
}

func NewServer(config *ServerConfig) *Server {
    return &Server{
        router:     mux.NewRouter(),
        handlers:   make(map[string]Handler),
        middleware: []Middleware{},
        config:     config,
        logger:     log.New(os.Stdout, "[SERVER] ", log.LstdFlags),
    }
}

func (s *Server) RegisterHandler(handler Handler) {
    s.handlers[handler.GetName()] = handler
    handler.RegisterRoutes(s.router)
    s.logger.Printf("Registered handler: %s v%s", handler.GetName(), handler.GetVersion())
}

func (s *Server) AddMiddleware(middleware Middleware) {
    s.middleware = append(s.middleware, middleware)
}

func (s *Server) Start() error {
    // Apply middleware
    for _, m := range s.middleware {
        s.router.Use(m.Handle)
    }
    
    // Add global routes
    s.addGlobalRoutes()
    
    // Create HTTP server
    server := &http.Server{
        Addr:           s.config.Host + ":" + s.config.Port,
        Handler:        s.router,
        ReadTimeout:    s.config.ReadTimeout,
        WriteTimeout:   s.config.WriteTimeout,
        MaxHeaderBytes: s.config.MaxHeaderSize,
    }
    
    s.logger.Printf("Starting server on %s:%s", s.config.Host, s.config.Port)
    return server.ListenAndServe()
}
```

### Global Routes
```go
func (s *Server) addGlobalRoutes() {
    // Health check
    s.router.HandleFunc("/health", s.healthCheck).Methods("GET")
    
    // API info
    s.router.HandleFunc("/api/info", s.apiInfo).Methods("GET")
    
    // Handler status
    s.router.HandleFunc("/api/handlers", s.listHandlers).Methods("GET")
    s.router.HandleFunc("/api/handlers/{name}/health", s.handlerHealth).Methods("GET")
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
    status := map[string]interface{}{
        "status":    "healthy",
        "timestamp": time.Now().UTC(),
        "version":   "1.0.0",
    }
    
    // Check all handlers
    handlerStatus := make(map[string]string)
    for name, handler := range s.handlers {
        if err := handler.HealthCheck(); err != nil {
            handlerStatus[name] = "unhealthy"
            status["status"] = "degraded"
        } else {
            handlerStatus[name] = "healthy"
        }
    }
    status["handlers"] = handlerStatus
    
    s.respondWithJSON(w, http.StatusOK, status)
}

func (s *Server) apiInfo(w http.ResponseWriter, r *http.Request) {
    info := map[string]interface{}{
        "name":        "Music Composer Assistant API",
        "version":     "1.0.0",
        "description": "REST API for music composition with LilyPond",
        "handlers":    s.getHandlerInfo(),
        "endpoints":   s.getEndpointInfo(),
    }
    
    s.respondWithJSON(w, http.StatusOK, info)
}

func (s *Server) getHandlerInfo() map[string]interface{} {
    info := make(map[string]interface{})
    for name, handler := range s.handlers {
        info[name] = map[string]string{
            "version": handler.GetVersion(),
            "status":  "active",
        }
    }
    return info
}
```

## Middleware

### Common Middleware
```go
type Middleware interface {
    Handle(next http.Handler) http.Handler
}

// Logging middleware
type LoggingMiddleware struct {
    logger *log.Logger
}

func (lm *LoggingMiddleware) Handle(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Create response writer wrapper to capture status
        wrapped := &responseWriter{ResponseWriter: w}
        
        next.ServeHTTP(wrapped, r)
        
        duration := time.Since(start)
        lm.logger.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapped.status, duration)
    })
}

// CORS middleware
type CORSMiddleware struct {
    allowedOrigins []string
}

func (cm *CORSMiddleware) Handle(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        origin := r.Header.Get("Origin")
        if cm.isAllowedOrigin(origin) {
            w.Header().Set("Access-Control-Allow-Origin", origin)
        }
        
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

// Rate limiting middleware
type RateLimitMiddleware struct {
    limiter *rate.Limiter
}

func (rlm *RateLimitMiddleware) Handle(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !rlm.limiter.Allow() {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

## Error Handling

### Standardized Error Responses
```go
type APIError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

func (h *BaseHandler) respondWithError(w http.ResponseWriter, statusCode int, message string) {
    error := APIError{
        Code:    statusCode,
        Message: message,
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(error)
}

func (h *BaseHandler) respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(data)
}
```

## Usage Example

### Server Setup
```go
func main() {
    // Create server config
    config := &ServerConfig{
        Port:          "8080",
        Host:          "localhost",
        ReadTimeout:   30 * time.Second,
        WriteTimeout:  30 * time.Second,
        MaxHeaderSize: 1 << 20,
        CORSEnabled:   true,
        Logging:       true,
    }
    
    // Create server
    server := NewServer(config)
    
    // Initialize handlers
    promptEngine := NewPromptEngine(4000, 8000)
    diffGenerator := NewDiffGenerator(3, 1000, "/tmp")
    lilypondProcessor := NewLilyPondProcessor("/tmp", "lilypond", 30*time.Second, 3)
    contextManager := NewContextManager()
    
    // Register handlers
    server.RegisterHandler(NewPromptHandler(promptEngine))
    server.RegisterHandler(NewDiffHandler(diffGenerator))
    server.RegisterHandler(NewLilyPondHandler(lilypondProcessor))
    server.RegisterHandler(NewContextHandler(contextManager))
    
    // Add middleware
    server.AddMiddleware(&LoggingMiddleware{logger: log.New(os.Stdout, "[ACCESS] ", log.LstdFlags)})
    server.AddMiddleware(&CORSMiddleware{allowedOrigins: []string{"*"}})
    server.AddMiddleware(&RateLimitMiddleware{limiter: rate.NewLimiter(rate.Every(time.Second), 100)})
    
    // Start server
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}
```

## Testing

### Handler Tests
```go
func TestDiffHandler_GetDiff(t *testing.T) {
    diffGenerator := NewDiffGenerator(3, 1000, "/tmp")
    handler := NewDiffHandler(diffGenerator)
    
    request := httptest.NewRequest("GET", "/api/diff/get?id=test-diff-123", nil)
    response := httptest.NewRecorder()
    
    handler.GetDiff(response, request)
    
    assert.Equal(t, http.StatusOK, response.Code)
    
    var resp GetDiffResponse
    json.Unmarshal(response.Body.Bytes(), &resp)
    assert.True(t, resp.Success)
    assert.NotNil(t, resp.Data)
}
```

## API Documentation

### Endpoint Summary
```
GET    /health                    - Server health check
GET    /api/info                  - API information
GET    /api/handlers              - List all handlers
GET    /api/handlers/{name}/health - Handler health check

GET    /api/prompt/templates      - List templates
GET    /api/prompt/templates/{name} - Get template

GET    /api/diff/get              - Get diff from system state
POST   /api/diff/render           - Render diff
POST   /api/diff/cache/clear      - Clear cache

POST   /api/lilypond/compile      - Compile LilyPond
POST   /api/lilypond/validate     - Validate syntax
GET    /api/lilypond/pdf/{id}     - Get PDF
GET    /api/lilypond/health       - LilyPond health check

GET    /api/context/document      - Get current document
POST   /api/context/save          - Save current document
``` 