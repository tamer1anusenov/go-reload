package main

import (
	"fmt"
	"go-reloaded/processor"
	"os"
	"path/filepath"
	"strings"
)

// isValidTxtFile checks if the filename has a valid .txt extension
func isValidTxtFile(filename string) bool {
	// Check if file has .txt extension
	if !strings.HasSuffix(strings.ToLower(filename), ".txt") {
		return false
	}

	// Check if there's a filename before the extension
	base := filepath.Base(filename)
	if base == ".txt" || strings.HasPrefix(base, ".") {
		return false
	}

	return true
}

func main() {
	// Check command line arguments
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input_file> <output_file>\n", os.Args[0])
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	// Validate file extensions
	if !isValidTxtFile(inputFile) {
		fmt.Fprintf(os.Stderr, "Error: Input file must have .txt extension\n")
		os.Exit(1)
	}
	if !isValidTxtFile(outputFile) {
		fmt.Fprintf(os.Stderr, "Error: Output file must have .txt extension\n")
		os.Exit(1)
	}

	// Check if input and output files are the same
	absInputPath, err := filepath.Abs(inputFile)
	if err == nil {
		absOutputPath, err := filepath.Abs(outputFile)
		if err == nil && absInputPath == absOutputPath {
			fmt.Fprintf(os.Stderr, "Error: Input and output files cannot be the same\n")
			os.Exit(1)
		}
	}

	// Check if input file exists
	_, err = os.Stat(inputFile)
	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Input file '%s' does not exist\n", inputFile)
		os.Exit(1)
	}

	// Check if output file is writable
	outputDir := filepath.Dir(outputFile)
	if outputDir == "." {
		outputDir = "."
	}
	info, err := os.Stat(outputDir)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: Output directory '%s' does not exist\n", outputDir)
		os.Exit(1)
	}

	// Read input file
	content, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// Process the content using the processor package
	processedText := processor.ProcessText(string(content))

	// Write to output file
	err = os.WriteFile(outputFile, []byte(processedText), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}
}
