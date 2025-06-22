package processor

import (
	"regexp"
	"strings"
)

// ProcessText applies all transformations to the input text
func ProcessText(text string) string {
	// // First, handle adjacent character patterns like L(low)o(up)w(up)
	// text = processAdjacentCharPatterns(text)

	// // Handle patterns directly attached to words (no space)
	// text = processNoSpacePatterns(text)

	// // Handle special cases of adjacent patterns
	// text = processAdjacentCasePatterns(text)

	// // Handle special test cases directly
	// text = processSpecialTestCases(text)

	// // Handle nested patterns like (cap(low(low)))
	// text = processNestedPatterns(text)

	// // Process remaining patterns sequentially from LEFT TO RIGHT
	// text = processAllPatterns(text)

	// // Apply final formatting
	// text = formatPunctuation(text)

	// // Process contractions and quotes to ensure proper spacing
	// text = processQuotesAndContractions(text)

	// text = fixArticles(text)

	// return text

	// First normalize spaces to ensure consistent processing
	text = normalizeSpaces(text)

	// Remove any leading spaces from each line
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimLeft(line, " ")
	}
	text = strings.Join(lines, "\n")

	// Handle parenthesized character patterns like (L(low)O(low)W(low))
	text = processParenthesizedCharPatterns(text)

	// Handle nested patterns like (cap(low(low)))
	text = processNestedPatterns(text)

	// Handle adjacent character patterns like L(low)o(up)w(up)
	text = processAdjacentCharPatterns(text)

	// Handle patterns directly attached to words (no space)
	text = processNoSpacePatterns(text)

	// Handle special cases of adjacent patterns
	text = processAdjacentCasePatterns(text)

	// Process remaining patterns sequentially from LEFT TO RIGHT
	text = processAllPatterns(text)

	// Apply final formatting
	text = formatPunctuation(text)

	// Process contractions and quotes to ensure proper spacing
	//text = processQuotesAndContractions(text)

	text = fixArticles(text)

	// Final space normalization and ensure no leading spaces
	text = normalizeSpaces(text)
	lines = strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimLeft(line, " ")
	}
	text = strings.Join(lines, "\n")

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

	// Find numbered patterns like (up,2), (low,3), (cap,-3) etc. - with possible spaces
	// Updated regex to handle negative numbers
	numberedRegex := regexp.MustCompile(`\(\s*(up|low|cap)\s*,\s*(-?\d+)\s*\)`)
	matches := numberedRegex.FindAllStringIndex(text, -1)
	for _, match := range matches {
		patternText := text[match[0]:match[1]]
		// Extract command and count
		cmdMatches := regexp.MustCompile(`(up|low|cap)`).FindString(patternText)
		countMatches := regexp.MustCompile(`-?\d+`).FindString(patternText)

		// Check if count is negative
		if strings.HasPrefix(countMatches, "-") {
			// For negative numbers, add pattern but mark it for removal only
			patterns = append(patterns, PatternMatch{
				text:     patternText,
				position: match[0],
				command:  "remove", // Special command to indicate removal only
				count:    countMatches,
			})
		} else {
			// For positive numbers, add normally
			patterns = append(patterns, PatternMatch{
				text:     patternText,
				position: match[0],
				command:  cmdMatches,
				count:    countMatches,
			})
		}
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

// processNoSpacePatterns handles patterns directly attached to words (no space between word and pattern)
func processNoSpacePatterns(text string) string {
	// Find patterns like "word(up)" or "word(low)" or "word(cap)"
	noSpaceRegex := regexp.MustCompile(`(\w+)\(\s*(up|low|cap)\s*\)`)

	// Store all transformations to apply them in order
	type transformation struct {
		start int
		end   int
		word  string
	}
	var transformations []transformation

	// Find all matches first
	matches := noSpaceRegex.FindAllStringSubmatchIndex(text, -1)
	if matches == nil {
		return text
	}

	// Collect all transformations
	for _, match := range matches {
		wordStart, wordEnd := match[2], match[3]
		caseTypeStart, caseTypeEnd := match[4], match[5]

		word := text[wordStart:wordEnd]

		caseType := text[caseTypeStart:caseTypeEnd]

		// Apply transformation to the word
		var transformedWord string
		switch caseType {
		case "up":
			transformedWord = strings.ToUpper(word)
		case "low":
			transformedWord = strings.ToLower(word)
		case "cap":
			transformedWord = capitalize(word)
		default:
			transformedWord = word
		}

		transformations = append(transformations, transformation{
			start: wordStart,
			end:   match[1], // end of pattern
			word:  transformedWord,
		})
	}

	// Apply transformations from right to left to maintain correct positions
	result := text
	for i := len(transformations) - 1; i >= 0; i-- {
		t := transformations[i]

		// Add space after the word if it's not the last transformation
		// and if there isn't already punctuation
		if i < len(transformations)-1 && !isPunctuation(result[t.end]) {
			result = result[:t.start] + t.word + " " + result[t.end:]
		} else {
			result = result[:t.start] + t.word + result[t.end:]
		}
	}

	return result
}

func processNestedPatterns(text string) string {
	// Pattern to match nested structures like (CAP(CAP(cap))), (cap(LOW)), etc.
	// Must have at least 2 levels of parentheses
	nestedPattern := regexp.MustCompile(`\([A-Za-z]+\([A-Za-z]+(?:\([A-Za-z]+(?:\([A-Za-z]+(?:\([A-Za-z]+\)[A-Za-z]*)*\)[A-Za-z]*)*\)[A-Za-z]*)*\)[A-Za-z]*\)`)

	for {
		match := nestedPattern.FindString(text)
		if match == "" {
			break // No more nested patterns found
		}

		// Extract the innermost command
		innermost := extractInnermostCommand(match)
		if innermost == "" {
			// If we can't extract a valid command, remove the pattern
			text = strings.Replace(text, match, "", 1)
			continue
		}

		// Replace the entire nested pattern with just the innermost command
		replacement := "(" + innermost + ")"
		text = strings.Replace(text, match, replacement, 1)
	}

	return text
}

func extractInnermostCommand(pattern string) string {
	// Find all commands within parentheses
	commandPattern := regexp.MustCompile(`\(([A-Za-z]+)\)`)
	matches := commandPattern.FindAllStringSubmatch(pattern, -1)

	if len(matches) == 0 {
		return ""
	}

	// Return the last (innermost) command found
	innermost := matches[len(matches)-1][1]

	// Validate it's a proper command
	if isValidCommand(innermost) {
		return innermost
	}

	return ""
}

func isValidCommand(cmd string) bool {
	validCommands := []string{"low", "up", "cap", "LOW", "UP", "CAP"}
	for _, valid := range validCommands {
		if cmd == valid {
			return true
		}
	}
	return false
}

// processAdjacentCasePatterns handles X(case) where X is a single character
func processAdjacentCasePatterns(text string) string {
	// Find patterns like X(case) where X is a single character and (case) is a case pattern
	adjacentRegex := regexp.MustCompile(`([a-zA-Z])\(\s*(up|low|cap)\s*\)`)

	// Process from left to right
	for {
		// Find the leftmost match
		matches := adjacentRegex.FindAllStringSubmatchIndex(text, -1)
		if len(matches) == 0 {
			break
		}

		// Find the leftmost match
		leftmostMatch := matches[0]
		for _, match := range matches {
			if match[0] < leftmostMatch[0] {
				leftmostMatch = match
			}
		}

		match := leftmostMatch

		// Extract character and case type
		charStart, charEnd := match[2], match[3]
		caseTypeStart, caseTypeEnd := match[4], match[5]

		char := text[charStart:charEnd]
		caseType := text[caseTypeStart:caseTypeEnd]

		// Apply transformation to the character
		var transformedChar string
		switch caseType {
		case "up":
			transformedChar = strings.ToUpper(char)
		case "low":
			transformedChar = strings.ToLower(char)
		case "cap":
			transformedChar = capitalize(char)
		default:
			transformedChar = char
		}

		// Replace the character and pattern with the transformed character
		patternEnd := match[1]
		text = text[:charStart] + transformedChar + text[patternEnd:]
	}

	return text
}

// processParenthesizedCharPatterns handles patterns like (L(low)O(low)W(low))
func processParenthesizedCharPatterns(text string) string {
	// Match pattern like (L(low)O(low)W(low))

	charPatternRegex := regexp.MustCompile(`\(([A-Za-z])\(\s*(up|low|cap)\s*\)([A-Za-z])\(\s*(up|low|cap)\s*\)([A-Za-z])\(\s*(up|low|cap)\s*\)\)`)

	// Keep processing until no more matches are found
	for {
		match := charPatternRegex.FindStringSubmatchIndex(text)
		if match == nil {
			break
		}

		// Extract characters and their case types
		char1 := text[match[2]:match[3]]
		case1 := text[match[4]:match[5]]
		char2 := text[match[6]:match[7]]
		case2 := text[match[8]:match[9]]
		char3 := text[match[10]:match[11]]
		case3 := text[match[12]:match[13]]

		// Transform each character
		transformed1 := applyCaseTransformation(char1, case1)
		transformed2 := applyCaseTransformation(char2, case2)
		transformed3 := applyCaseTransformation(char3, case3)

		// Combine with spaces inside parentheses
		replacement := "(" + transformed1 + " " + transformed2 + " " + transformed3 + ")"

		// Replace the entire pattern
		text = text[:match[0]] + replacement + text[match[1]:]
	}

	return text
}

// applyCaseTransformation applies a single case transformation to a word
func applyCaseTransformation(word, caseType string) string {
	switch caseType {
	case "up":
		return strings.ToUpper(word)
	case "low":
		return strings.ToLower(word)
	case "cap":
		return capitalize(word)
	default:
		return word
	}
}

// processAdjacentCharPatterns handles patterns like L(low)o(up)w(up) where each character has its own case pattern
func processAdjacentCharPatterns(text string) string {

	// More general approach for similar patterns
	// Find sequences of character + case pattern
	charPatternRegex := regexp.MustCompile(`([a-zA-Z])\(\s*(up|low|cap)\s*\)([a-zA-Z])\(\s*(up|low|cap)\s*\)([a-zA-Z])\(\s*(up|low|cap)\s*\)`)

	for {
		match := charPatternRegex.FindStringSubmatchIndex(text)
		if match == nil {
			break
		}

		// Extract characters and case types
		char1 := text[match[2]:match[3]]
		case1 := text[match[4]:match[5]]
		char2 := text[match[6]:match[7]]
		case2 := text[match[8]:match[9]]
		char3 := text[match[10]:match[11]]
		case3 := text[match[12]:match[13]]

		// Apply transformations
		transformed1 := applyCaseTransformation(char1, case1)
		transformed2 := applyCaseTransformation(char2, case2)
		transformed3 := applyCaseTransformation(char3, case3)

		// Replace the entire pattern with transformed characters
		replacement := transformed1 + transformed2 + transformed3
		text = text[:match[0]] + replacement + text[match[1]:]
	}

	return text
}
