package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cstevenson98/goFE/pkg/goFE/gox"
)

func main() {
	// Parse command line flags
	inputDir := flag.String("input", ".", "Input directory containing .gox files")
	outputDir := flag.String("output", ".", "Output directory for generated .go files")
	flag.Parse()

	// Walk through the input directory
	err := filepath.Walk(*inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Process only .gox files
		if !strings.HasSuffix(path, ".gox") {
			return nil
		}

		// Parse the .gox file
		file, err := gox.ParseFile(path)
		if err != nil {
			return fmt.Errorf("error parsing %s: %v", path, err)
		}

		// Generate Go code
		code, err := gox.GenerateGoCode(file)
		if err != nil {
			return fmt.Errorf("error generating code for %s: %v", path, err)
		}

		// Create output file path
		relPath, err := filepath.Rel(*inputDir, path)
		if err != nil {
			return fmt.Errorf("error getting relative path for %s: %v", path, err)
		}
		outputPath := filepath.Join(*outputDir, strings.TrimSuffix(relPath, ".gox")+".go")

		// Create output directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return fmt.Errorf("error creating output directory for %s: %v", outputPath, err)
		}

		// Write the generated code to the output file
		if err := os.WriteFile(outputPath, []byte(code), 0644); err != nil {
			return fmt.Errorf("error writing %s: %v", outputPath, err)
		}

		fmt.Printf("Generated %s\n", outputPath)
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
} 