package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cstevenson98/goFE/backends/anthropic-agent/internal/core/lilypond"
	"github.com/cstevenson98/goFE/backends/anthropic-agent/internal/types"
	"github.com/gorilla/mux"
)

// LilyPondHandler handles LilyPond operations
type LilyPondHandler struct {
	lilypondProcessor *lilypond.LilyPondProcessor
	documents         map[string]*LilyPondDocument
	mutex             sync.RWMutex
}

// LilyPondDocument represents a LilyPond document
type LilyPondDocument struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Status    string    `json:"status"` // "draft", "compiled", "error"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewLilyPondHandler creates a new LilyPond handler
func NewLilyPondHandler(lp *lilypond.LilyPondProcessor) *LilyPondHandler {
	handler := &LilyPondHandler{
		lilypondProcessor: lp,
		documents:         make(map[string]*LilyPondDocument),
	}

	// Load existing documents from disk on startup
	handler.LoadExistingDocuments()

	return handler
}

// CreateLilyPondDocumentRequest represents a request to create a LilyPond document
type CreateLilyPondDocumentRequest struct {
	Title   string `json:"title" validate:"required"`
	Content string `json:"content" validate:"required"`
}

// UpdateLilyPondDocumentRequest represents a request to update a LilyPond document
type UpdateLilyPondDocumentRequest struct {
	Title   string `json:"title" validate:"required"`
	Content string `json:"content" validate:"required"`
}

// CompileLilyPondDocumentRequest represents a request to compile a LilyPond document
type CompileLilyPondDocumentRequest struct {
	Options *lilypond.CompileOptions `json:"options,omitempty"`
}

// CreateLilyPondDocument handles LilyPond document creation
func (lh *LilyPondHandler) CreateLilyPondDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var request CreateLilyPondDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		lh.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if request.Title == "" {
		lh.respondWithError(w, http.StatusBadRequest, "Title is required")
		return
	}

	if request.Content == "" {
		lh.respondWithError(w, http.StatusBadRequest, "Content is required")
		return
	}

	// Generate document ID
	id := fmt.Sprintf("lilypond_%d", time.Now().UnixNano())

	// Create document
	doc := &LilyPondDocument{
		ID:        id,
		Title:     request.Title,
		Content:   request.Content,
		Status:    "draft",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store document
	lh.mutex.Lock()
	lh.documents[id] = doc
	lh.mutex.Unlock()

	// Save LilyPond source file to output directory
	if err := lh.lilypondProcessor.SaveLilyPondSource(id, doc.Content); err != nil {
		log.Printf("Warning: Failed to save LilyPond source file: %v", err)
	}

	// Create response
	response := types.APIResponse[LilyPondDocument]{
		Data:    *doc,
		Success: true,
		Message: "LilyPond document created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListLilyPondDocuments handles listing all LilyPond documents
func (lh *LilyPondHandler) ListLilyPondDocuments(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	// Get documents
	lh.mutex.RLock()
	documents := make([]*LilyPondDocument, 0, len(lh.documents))
	for _, doc := range lh.documents {
		documents = append(documents, doc)
	}
	lh.mutex.RUnlock()

	// Calculate pagination
	total := len(documents)
	totalPages := (total + limit - 1) / limit
	start := (page - 1) * limit
	end := start + limit
	if end > total {
		end = total
	}

	// Get page of documents
	var pageDocuments []*LilyPondDocument
	if start < total {
		pageDocuments = documents[start:end]
	}

	// Create response
	response := types.APIResponse[map[string]interface{}]{
		Data: map[string]interface{}{
			"documents": pageDocuments,
			"pagination": map[string]interface{}{
				"page":       page,
				"limit":      limit,
				"total":      total,
				"totalPages": totalPages,
			},
		},
		Success: true,
		Message: "LilyPond documents retrieved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetLilyPondDocument handles retrieving a specific LilyPond document
func (lh *LilyPondHandler) GetLilyPondDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	lh.mutex.RLock()
	doc, exists := lh.documents[id]
	lh.mutex.RUnlock()

	if !exists {
		lh.respondWithError(w, http.StatusNotFound, "LilyPond document not found")
		return
	}

	response := types.APIResponse[LilyPondDocument]{
		Data:    *doc,
		Success: true,
		Message: "LilyPond document retrieved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateLilyPondDocument handles updating a LilyPond document
func (lh *LilyPondHandler) UpdateLilyPondDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	var request UpdateLilyPondDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		lh.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if request.Title == "" {
		lh.respondWithError(w, http.StatusBadRequest, "Title is required")
		return
	}

	if request.Content == "" {
		lh.respondWithError(w, http.StatusBadRequest, "Content is required")
		return
	}

	lh.mutex.Lock()
	doc, exists := lh.documents[id]
	if !exists {
		lh.mutex.Unlock()
		lh.respondWithError(w, http.StatusNotFound, "LilyPond document not found")
		return
	}

	// Update document
	doc.Title = request.Title
	doc.Content = request.Content
	doc.Status = "draft" // Reset status when content changes
	doc.UpdatedAt = time.Now()
	lh.mutex.Unlock()

	// Save updated LilyPond source file to output directory
	if err := lh.lilypondProcessor.SaveLilyPondSource(id, doc.Content); err != nil {
		log.Printf("Warning: Failed to save updated LilyPond source file: %v", err)
	}

	response := types.APIResponse[LilyPondDocument]{
		Data:    *doc,
		Success: true,
		Message: "LilyPond document updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteLilyPondDocument handles deleting a LilyPond document
func (lh *LilyPondHandler) DeleteLilyPondDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	lh.mutex.Lock()
	doc, exists := lh.documents[id]
	if !exists {
		lh.mutex.Unlock()
		lh.respondWithError(w, http.StatusNotFound, "LilyPond document not found")
		return
	}

	// Store document info before deletion for cleanup
	documentID := doc.ID
	delete(lh.documents, id)
	lh.mutex.Unlock()

	// Clean up associated files for this document
	log.Printf("Cleaning up files for document: %s", documentID)
	lh.cleanupDocumentFiles(documentID)

	// Also clean up any orphaned files that might exist
	log.Printf("Cleaning up orphaned files...")
	lh.CleanupOrphanedFiles()

	response := types.APIResponse[map[string]interface{}]{
		Data: map[string]interface{}{
			"id": documentID,
		},
		Success: true,
		Message: "LilyPond document deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetLilyPondSource handles retrieving the LilyPond source code
func (lh *LilyPondHandler) GetLilyPondSource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	lh.mutex.RLock()
	doc, exists := lh.documents[id]
	lh.mutex.RUnlock()

	if !exists {
		lh.respondWithError(w, http.StatusNotFound, "LilyPond document not found")
		return
	}

	response := types.APIResponse[map[string]interface{}]{
		Data: map[string]interface{}{
			"id":     id,
			"source": doc.Content,
		},
		Success: true,
		Message: "LilyPond source code retrieved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetLilyPondFilePath handles retrieving the file path
func (lh *LilyPondHandler) GetLilyPondFilePath(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	lh.mutex.RLock()
	_, exists := lh.documents[id]
	lh.mutex.RUnlock()

	if !exists {
		lh.respondWithError(w, http.StatusNotFound, "LilyPond document not found")
		return
	}

	filePath := filepath.Join(lh.lilypondProcessor.GetOutputDir(), id+".ly")

	response := types.APIResponse[map[string]interface{}]{
		Data: map[string]interface{}{
			"id":        id,
			"file_path": filePath,
		},
		Success: true,
		Message: "LilyPond file path retrieved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetLilyPondPDFPath handles retrieving the PDF file path
func (lh *LilyPondHandler) GetLilyPondPDFPath(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	lh.mutex.RLock()
	doc, exists := lh.documents[id]
	lh.mutex.RUnlock()

	if !exists {
		lh.respondWithError(w, http.StatusNotFound, "LilyPond document not found")
		return
	}

	if doc.Status != "compiled" {
		lh.respondWithError(w, http.StatusBadRequest, "Document has not been compiled")
		return
	}

	pdfPath := filepath.Join(lh.lilypondProcessor.GetOutputDir(), id+".pdf")

	response := types.APIResponse[map[string]interface{}]{
		Data: map[string]interface{}{
			"id":       id,
			"pdf_path": pdfPath,
		},
		Success: true,
		Message: "LilyPond PDF path retrieved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetLilyPondTempDir handles retrieving the temporary directory
func (lh *LilyPondHandler) GetLilyPondTempDir(w http.ResponseWriter, r *http.Request) {
	response := types.APIResponse[map[string]interface{}]{
		Data: map[string]interface{}{
			"output_dir": lh.lilypondProcessor.GetOutputDir(),
		},
		Success: true,
		Message: "LilyPond output directory retrieved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CompileLilyPondDocument handles compiling a LilyPond document
func (lh *LilyPondHandler) CompileLilyPondDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	lh.mutex.RLock()
	doc, exists := lh.documents[id]
	lh.mutex.RUnlock()

	if !exists {
		lh.respondWithError(w, http.StatusNotFound, "LilyPond document not found")
		return
	}

	// Compile the document with the document ID
	result, err := lh.lilypondProcessor.CompileToPDFWithID(doc.Content, id)

	// Update document status
	lh.mutex.Lock()
	if err != nil || !result.Success {
		doc.Status = "error"
	} else {
		doc.Status = "compiled"
	}
	doc.UpdatedAt = time.Now()
	lh.mutex.Unlock()

	if err != nil {
		// Return error response with stdout/stderr information
		errorData := map[string]interface{}{
			"id":     id,
			"status": doc.Status,
		}

		// Always include stdout/stderr since result is guaranteed to be non-nil now
		errorData["stdout"] = result.Stdout
		errorData["stderr"] = result.Stderr

		response := types.APIResponse[map[string]interface{}]{
			Data:    errorData,
			Success: false,
			Error:   fmt.Sprintf("LilyPond compilation failed: %v", err),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Determine if the compilation was actually successful
	success := result.Success && doc.Status == "compiled"

	response := types.APIResponse[map[string]interface{}]{
		Data: map[string]interface{}{
			"id":       id,
			"status":   doc.Status,
			"pdf_path": result.OutputPath,
			"stdout":   result.Stdout,
			"stderr":   result.Stderr,
		},
		Success: success,
		Message: func() string {
			if success {
				return "LilyPond document compiled successfully"
			}
			return "LilyPond compilation completed with errors"
		}(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetLilyPondPDF handles downloading the compiled PDF
func (lh *LilyPondHandler) GetLilyPondPDF(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	lh.mutex.RLock()
	doc, exists := lh.documents[id]
	lh.mutex.RUnlock()

	if !exists {
		lh.respondWithError(w, http.StatusNotFound, "LilyPond document not found")
		return
	}

	if doc.Status != "compiled" {
		lh.respondWithError(w, http.StatusBadRequest, "Document has not been compiled")
		return
	}

	// Get the PDF file path
	pdfPath := filepath.Join(lh.lilypondProcessor.GetOutputDir(), id+".pdf")

	// Check if the PDF file exists
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		lh.respondWithError(w, http.StatusNotFound, "PDF file not found")
		return
	}

	// Read the PDF file
	pdfData, err := os.ReadFile(pdfPath)
	if err != nil {
		lh.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to read PDF file: %v", err))
		return
	}

	// Serve the PDF
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%s.pdf", id))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfData)))
	w.Write(pdfData)
}

// ValidateLilyPondSyntax handles syntax validation
func (lh *LilyPondHandler) ValidateLilyPondSyntax(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var request struct {
		Content string `json:"content" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		lh.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if request.Content == "" {
		lh.respondWithError(w, http.StatusBadRequest, "Content is required")
		return
	}

	// Validate syntax
	result := lh.lilypondProcessor.ValidateSyntax(request.Content)

	response := types.APIResponse[lilypond.ValidationResult]{
		Data:    *result,
		Success: result.IsValid,
		Message: "Syntax validation completed",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// cleanupDocumentFiles removes the associated .ly and .pdf files for a document
func (lh *LilyPondHandler) cleanupDocumentFiles(documentID string) {
	outputDir := lh.lilypondProcessor.GetOutputDir()

	// Remove .ly file
	lyFile := filepath.Join(outputDir, documentID+".ly")
	if err := os.Remove(lyFile); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Warning: Failed to remove LilyPond file %s: %v", lyFile, err)
		} else {
			log.Printf("Info: LilyPond file %s does not exist (already removed)", lyFile)
		}
	} else {
		log.Printf("Successfully removed LilyPond file: %s", lyFile)
	}

	// Remove .pdf file
	pdfFile := filepath.Join(outputDir, documentID+".pdf")
	if err := os.Remove(pdfFile); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Warning: Failed to remove PDF file %s: %v", pdfFile, err)
		} else {
			log.Printf("Info: PDF file %s does not exist (already removed)", pdfFile)
		}
	} else {
		log.Printf("Successfully removed PDF file: %s", pdfFile)
	}

	// Also check for any other files that might be created by LilyPond (like .log files)
	logFile := filepath.Join(outputDir, documentID+".log")
	if err := os.Remove(logFile); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Failed to remove log file %s: %v", logFile, err)
	}
}

// CleanupOrphanedFiles removes LilyPond files that don't have corresponding documents
func (lh *LilyPondHandler) CleanupOrphanedFiles() {
	outputDir := lh.lilypondProcessor.GetOutputDir()

	// Get all .ly files in the output directory
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		log.Printf("Warning: Failed to read output directory %s: %v", outputDir, err)
		return
	}

	lh.mutex.RLock()
	existingDocs := make(map[string]bool)
	for docID := range lh.documents {
		existingDocs[docID] = true
	}
	lh.mutex.RUnlock()

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasSuffix(filename, ".ly") {
			continue
		}

		// Extract document ID from filename
		docID := strings.TrimSuffix(filename, ".ly")

		// Check if this document exists in our map
		if !existingDocs[docID] {
			log.Printf("Found orphaned LilyPond file: %s, removing...", filename)
			lh.cleanupDocumentFiles(docID)
		}
	}
}

// respondWithError sends an error response
func (lh *LilyPondHandler) respondWithError(w http.ResponseWriter, statusCode int, message string) {
	response := types.APIResponse[map[string]interface{}]{
		Data:    map[string]interface{}{},
		Success: false,
		Error:   message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// LoadExistingDocuments scans the output directory for existing LilyPond files and loads them into the documents map
func (lh *LilyPondHandler) LoadExistingDocuments() {
	outputDir := lh.lilypondProcessor.GetOutputDir()

	// Get all .ly files in the output directory
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		log.Printf("Warning: Failed to read output directory %s: %v", outputDir, err)
		return
	}

	loadedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasSuffix(filename, ".ly") {
			continue
		}

		// Extract document ID from filename (remove .ly extension)
		docID := strings.TrimSuffix(filename, ".ly")

		// Check if this is a lilypond_<timestamp> file
		if !strings.HasPrefix(docID, "lilypond_") {
			continue
		}

		// Read the content of the .ly file
		lyFilePath := filepath.Join(outputDir, filename)
		content, err := os.ReadFile(lyFilePath)
		if err != nil {
			log.Printf("Warning: Failed to read LilyPond file %s: %v", filename, err)
			continue
		}

		// Check if corresponding PDF exists to determine status
		pdfPath := filepath.Join(outputDir, docID+".pdf")
		status := "draft"
		if _, err := os.Stat(pdfPath); err == nil {
			status = "compiled"
		}

		// Get file info for timestamps
		fileInfo, err := entry.Info()
		if err != nil {
			log.Printf("Warning: Failed to get file info for %s: %v", filename, err)
			continue
		}

		// Create document
		doc := &LilyPondDocument{
			ID:        docID,
			Title:     fmt.Sprintf("LilyPond Document %s", docID),
			Content:   string(content),
			Status:    status,
			CreatedAt: fileInfo.ModTime(),
			UpdatedAt: fileInfo.ModTime(),
		}

		// Add to documents map
		lh.mutex.Lock()
		lh.documents[docID] = doc
		lh.mutex.Unlock()

		loadedCount++
		log.Printf("Loaded existing LilyPond document: %s (status: %s)", docID, status)
	}

	log.Printf("Loaded %d existing LilyPond documents from disk", loadedCount)
}
