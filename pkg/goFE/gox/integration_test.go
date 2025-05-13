package gox

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegration(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "gox-integration-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test component files
	testFiles := map[string]string{
		"Button.gox":  "testdata/button.gox",
		"Counter.gox": "testdata/counter.gox",
	}

	// Write test files
	for name, sourcePath := range testFiles {
		// Load source content
		content, err := LoadGoxFromFile(sourcePath)
		if err != nil {
			t.Fatalf("Failed to load test file %s: %v", sourcePath, err)
		}

		// Write to temp directory
		path := filepath.Join(tempDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file %s: %v", name, err)
		}
	}

	// Process each file
	for name := range testFiles {
		path := filepath.Join(tempDir, name)

		// Parse the file
		file, err := ParseFile(path)
		if err != nil {
			t.Errorf("Failed to parse %s: %v", name, err)
			continue
		}

		// Generate Go code
		code, err := GenerateGoCode(file)
		if err != nil {
			t.Errorf("Failed to generate code for %s: %v", name, err)
			continue
		}

		// Verify the generated code
		if !strings.Contains(code, "package components") {
			t.Errorf("Generated code for %s missing package declaration", name)
		}

		componentName := strings.TrimSuffix(name, ".gox")
		if !strings.Contains(code, "type "+componentName+" struct") {
			t.Errorf("Generated code for %s missing component struct", name)
		}

		if !strings.Contains(code, "func New"+componentName) {
			t.Errorf("Generated code for %s missing constructor", name)
		}

		// Verify interface implementation
		interfaceMethods := []string{
			"GetID() uuid.UUID",
			"GetChildren() []goFE.Component",
			"InitEventListeners()",
			"Render() string",
		}

		for _, method := range interfaceMethods {
			if !strings.Contains(code, method) {
				t.Errorf("Generated code for %s missing interface method: %s", name, method)
			}
		}
	}
}
