package processor

import (
	"regexp"
	"strings"
)

// ProcessText applies all transformations to the input text
func ProcessText(text string) string {
	// Process patterns sequentially from LEFT TO RIGHT
	text = processAllPatterns(text)

	// Apply final formatting
	text = formatPunctuation(text)
	text = fixArticles(text)

	return text
}

// processAllPatterns handles all transformation patterns from left to right
func processAllPatterns(text string) string {
	// Process patterns until no more are found
	for {
		// Find the leftmost pattern
		pattern, position := findLeftmostPattern(text)
		if position == -1 {
			// No more patterns found
			break
		}

		// Apply the transformation and remove the pattern
		text = applyAndRemovePattern(text, pattern, position)
	}

	return text
}

// findLeftmostPattern finds the leftmost transformation pattern in the text
func findLeftmostPattern(text string) (string, int) {
	patterns := findAllPatterns(text)
	if len(patterns) == 0 {
		return "", -1
	}

	// Find leftmost pattern
	leftmost := patterns[0]
	for _, p := range patterns {
		if p.position < leftmost.position {
			leftmost = p
		}
	}

	return leftmost.text, leftmost.position
}

// PatternMatch represents a pattern found in text
type PatternMatch struct {
	text     string
	position int
	command  string
	count    string
}

// findAllPatterns finds all transformation patterns in the text
func findAllPatterns(text string) []PatternMatch {
	var patterns []PatternMatch

	// Find simple patterns (hex), (bin), (up), (low), (cap) - with possible spaces inside
	simplePatternRegex := regexp.MustCompile(`\(\s*(hex|bin|up|low|cap)\s*\)`)
	simpleMatches := simplePatternRegex.FindAllStringIndex(text, -1)
	for _, match := range simpleMatches {
		patternText := text[match[0]:match[1]]
		// Extract the actual command (hex, bin, up, low, cap)
		command := regexp.MustCompile(`(hex|bin|up|low|cap)`).FindString(patternText)
		// Store the original pattern with spaces
		patterns = append(patterns, PatternMatch{
			text:     patternText,
			position: match[0],
			command:  command,
		})
	}

	// Find numbered patterns like (up,2), (low,3), etc. - with possible spaces
	numberedRegex := regexp.MustCompile(`\(\s*(up|low|cap)\s*,\s*(\d+)\s*\)`)
	matches := numberedRegex.FindAllStringIndex(text, -1)
	for _, match := range matches {
		patternText := text[match[0]:match[1]]
		// Extract command and count
		cmdMatches := regexp.MustCompile(`(up|low|cap)`).FindString(patternText)
		countMatches := regexp.MustCompile(`\d+`).FindString(patternText)

		patterns = append(patterns, PatternMatch{
			text:     patternText,
			position: match[0],
			command:  cmdMatches,
			count:    countMatches,
		})
	}

	return patterns
}

// applyAndRemovePattern applies transformation and removes the pattern
func applyAndRemovePattern(text, pattern string, position int) string {
	var result string

	switch {
	case pattern == "(hex)" || strings.Contains(pattern, "hex"):
		result = processHexAtPosition(text, position)
	case pattern == "(bin)" || strings.Contains(pattern, "bin"):
		result = processBinAtPosition(text, position)
	case pattern == "(up)" || (strings.Contains(pattern, "up") && !strings.Contains(pattern, ",")):
		result = processCaseAtPosition(text, position, "up", 1)
	case pattern == "(low)" || (strings.Contains(pattern, "low") && !strings.Contains(pattern, ",")):
		result = processCaseAtPosition(text, position, "low", 1)
	case pattern == "(cap)" || (strings.Contains(pattern, "cap") && !strings.Contains(pattern, ",")):
		result = processCaseAtPosition(text, position, "cap", 1)
	case strings.Contains(pattern, ","):
		result = processNumberedCasePattern(text, pattern, position)
	default:
		// If no transformation applied, just remove the pattern
		result = removePatternAt(text, pattern, position)
	}

	// Apply formatQuotes to handle spaces inside quotes
	result = formatQuotes(result)

	return result
}
