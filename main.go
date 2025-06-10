package main

import (
	"fmt"
	"go-reloaded/processor"
	"os"
	"path/filepath"
)

func main() {
	// Check command line arguments
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input_file> <output_file>\n", os.Args[0])
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	// Check if input and output files are the same
	absInputPath, err := filepath.Abs(inputFile)
	if err == nil {
		absOutputPath, err := filepath.Abs(outputFile)
		if err == nil && absInputPath == absOutputPath {
			fmt.Fprintf(os.Stderr, "Error: enter different file\n")
			os.Exit(1)
		}
	}

	// Check if input file exists
	_, err = os.Stat(inputFile)
	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Input file '%s' does not exist\n", inputFile)
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
