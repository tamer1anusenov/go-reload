package process

import (
	"as/transformations"
	"strings"
)

// ProcessText applies all transformations to the input text in the correct order
// ProcessText applies transformations in correct order
func ProcessText(text string) string {
	// Step 1: Handle quotes first
	text = transformations.FixQuotes2(text)

	// Step 2: Fix punctuation groups
	text = transformations.FixPunctuationGroups(text)

	// Step 3: Apply case transformations
	text = transformations.TransformCase(text)

	// Step 4: Fix general punctuation spacing
	text = transformations.FixGeneralPunctuation(text)

	text = transformations.FixSpacing(text)

	text = transformations.CleanFinalSpacing(text)

	// Final cleanup: remove extra spaces
	return strings.Join(strings.Fields(text), " ")
}
