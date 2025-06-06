package processor

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// processHexAtPosition converts hex number before position and removes pattern
// processHexAtPosition converts hex number before position and removes pattern
func processHexAtPosition(text string, pos int) string {
	word, wordStart, _ := findWordBefore(text, pos)
	if word != "" {
		if decimal, err := strconv.ParseInt(word, 16, 64); err == nil {
			// Replace word and ALL consecutive (hex) patterns with single decimal
			before := text[:wordStart]
			after := text[pos+5:] // +5 for "(hex)"
			// Remove any consecutive (hex) patterns that follow
			after = regexp.MustCompile(`^(\s*\(hex\))*`).ReplaceAllString(after, "")
			return before + strconv.FormatInt(decimal, 10) + after
		}
	}
	// If conversion failed, just remove (hex)
	return removePatternAt(text, "(hex)", pos)
}

// processBinAtPosition converts binary number before position and removes pattern
func processBinAtPosition(text string, pos int) string {
	word, wordStart, _ := findWordBefore(text, pos)
	if word != "" {
		if decimal, err := strconv.ParseInt(word, 2, 64); err == nil {
			// Replace word and pattern with decimal
			before := text[:wordStart]
			after := text[pos+5:] // +5 for "(bin)"
			return before + strconv.FormatInt(decimal, 10) + after
		}
	}
	// If conversion failed, just remove (bin)
	return removePatternAt(text, "(bin)", pos)
}

// processCaseAtPosition applies case transformation and removes pattern
func processCaseAtPosition(text string, pos int, caseType string, count int) string {
	words, positions := findWordsBefore(text, pos, count)
	if len(words) > 0 {
		// Apply transformation to words
		transformedWords := make([]string, len(words))
		for i, word := range words {
			switch caseType {
			case "up":
				transformedWords[i] = strings.ToUpper(word)
			case "cap":
				caser := cases.Title(language.Und)
				transformedWords[i] = caser.String(strings.ToLower(word))
			case "low":
				transformedWords[i] = strings.ToLower(word)
			default:
				transformedWords[i] = word
			}
		}

		// Group words by their positions to handle multiple words with the same position
		// (words inside the same quotes or parentheses)
		positionGroups := make(map[string][]int) // key: "start,end", value: slice of indices
		for i, pos := range positions {
			key := fmt.Sprintf("%d,%d", pos[0], pos[1])
			positionGroups[key] = append(positionGroups[key], i)
		}

		// Create a copy of the text that we'll modify
		result := text

		// Process each position group (words at the same position)
		for posKey, indices := range positionGroups {
			var start, end int
			fmt.Sscanf(posKey, "%d,%d", &start, &end)

			// Get the original segment
			originalSegment := text[start:end]

			// Check if it's a quoted or parenthesized segment with multiple words
			if (start < end && len(originalSegment) > 0) &&
				((originalSegment[0] == '\'' && originalSegment[len(originalSegment)-1] == '\'') ||
					(originalSegment[0] == '"' && originalSegment[len(originalSegment)-1] == '"') ||
					(originalSegment[0] == '(' && originalSegment[len(originalSegment)-1] == ')')) {

				// Get content inside the delimiters
				originalContent := originalSegment[1 : len(originalSegment)-1]
				transformedContent := originalContent

				// Process each word in this group, in the correct order
				for _, idx := range indices {
					word := words[idx]
					transformedWord := transformedWords[idx]

					// Ensure we're only replacing whole words by using word boundary anchors
					wordRegex := regexp.MustCompile(`\b` + regexp.QuoteMeta(word) + `\b`)
					transformedContent = wordRegex.ReplaceAllString(transformedContent, transformedWord)
				}

				// Rebuild the segment with the transformed content
				newSegment := string(originalSegment[0]) + transformedContent + string(originalSegment[len(originalSegment)-1])

				// Replace in the result
				result = strings.Replace(result, originalSegment, newSegment, 1)
			} else {
				// For regular (non-quoted) words, we replace them one by one
				for _, idx := range indices {
					word := words[idx]
					transformedWord := transformedWords[idx]

					// Find the exact position of this word in the text
					// This is more accurate than using string.Replace which might replace multiple occurrences
					wordRegex := regexp.MustCompile(`\b` + regexp.QuoteMeta(word) + `\b`)
					result = wordRegex.ReplaceAllString(result, transformedWord)
				}
			}
		}

		// Remove the pattern
		patternLen := 0
		if count == 1 {
			patternLen = len(caseType) + 2 // e.g., "(up)" = 4
		} else {
			patternLen = len(fmt.Sprintf("(%s, %d)", caseType, count))
		}

		// The pattern position might have changed if we modified text before it
		// So we need to find it again in the transformed text
		patternPos := strings.Index(result, text[pos:pos+patternLen])
		if patternPos != -1 {
			result = result[:patternPos] + result[patternPos+patternLen:]
		} else {
			// Fallback: just remove from the original position
			// This shouldn't happen but it's a safety measure
			if pos < len(result) && pos+patternLen <= len(result) {
				result = result[:pos] + result[pos+patternLen:]
			}
		}

		return result
	}

	// If no words found, just remove pattern
	if count == 1 {
		return removePatternAt(text, fmt.Sprintf("(%s)", caseType), pos)
	}
	return removePatternAt(text, fmt.Sprintf("(%s, %d)", caseType, count), pos)
}

func processNumberedCasePattern(text, pattern string, position int) string {
	re := regexp.MustCompile(`\(\s*(up|low|cap)\s*,\s*(\d+)\s*\)`)
	matches := re.FindStringSubmatch(pattern)
	if len(matches) == 3 {
		caseType := matches[1]
		count, _ := strconv.Atoi(matches[2])

		// Apply transformation but don't let processCaseAtPosition remove pattern
		words, positions := findWordsBefore(text, position, count)
		if len(words) > 0 {
			// Apply transformation to words
			transformedWords := make([]string, len(words))
			for i, word := range words {
				switch caseType {
				case "cap":
					caser := cases.Title(language.Und)
					transformedWords[i] = caser.String(strings.ToLower(word))
				case "up":
					transformedWords[i] = strings.ToUpper(word)
				case "low":
					transformedWords[i] = strings.ToLower(word)
				}
			}

			// Group words by their positions to handle multiple words with the same position
			// (words inside the same quotes or parentheses)
			positionGroups := make(map[string][]int) // key: "start,end", value: slice of indices
			for i, pos := range positions {
				key := fmt.Sprintf("%d,%d", pos[0], pos[1])
				positionGroups[key] = append(positionGroups[key], i)
			}

			// Create a copy of the text that we'll modify
			result := text

			// Process each position group (words at the same position)
			for posKey, indices := range positionGroups {
				var start, end int
				fmt.Sscanf(posKey, "%d,%d", &start, &end)

				// Get the original segment
				originalSegment := text[start:end]

				// Check if it's a quoted or parenthesized segment with multiple words
				if (start < end && len(originalSegment) > 0) &&
					((originalSegment[0] == '\'' && originalSegment[len(originalSegment)-1] == '\'') ||
						(originalSegment[0] == '"' && originalSegment[len(originalSegment)-1] == '"') ||
						(originalSegment[0] == '(' && originalSegment[len(originalSegment)-1] == ')')) {

					// Get content inside the delimiters
					originalContent := originalSegment[1 : len(originalSegment)-1]
					transformedContent := originalContent

					// Process each word in this group, in the correct order
					for _, idx := range indices {
						word := words[idx]
						transformedWord := transformedWords[idx]

						// Ensure we're only replacing whole words by using word boundary anchors
						wordRegex := regexp.MustCompile(`\b` + regexp.QuoteMeta(word) + `\b`)
						transformedContent = wordRegex.ReplaceAllString(transformedContent, transformedWord)
					}

					// Rebuild the segment with the transformed content
					newSegment := string(originalSegment[0]) + transformedContent + string(originalSegment[len(originalSegment)-1])

					// Replace in the result
					result = strings.Replace(result, originalSegment, newSegment, 1)
				} else {
					// For regular (non-quoted) words, we replace them one by one
					for _, idx := range indices {
						word := words[idx]
						transformedWord := transformedWords[idx]

						// Find the exact position of this word in the text
						// This is more accurate than using string.Replace which might replace multiple occurrences
						wordRegex := regexp.MustCompile(`\b` + regexp.QuoteMeta(word) + `\b`)
						result = wordRegex.ReplaceAllString(result, transformedWord)
					}
				}
			}

			// Now remove the numbered pattern
			result = removePatternAt(result, pattern, position)
			return result
		}
	}
	return removePatternAt(text, pattern, position)
}

// formatPunctuation handles spacing around punctuation marks
func formatPunctuation(text string) string {
	// Handle regular punctuation: .,!?:;
	punctRegex := regexp.MustCompile(`\s+([.,!?:;])`)
	text = punctRegex.ReplaceAllString(text, "$1")

	// Add space after punctuation if not already present
	spaceRegex := regexp.MustCompile(`([.,!?:;])([^\s.,!?:;])`)
	text = spaceRegex.ReplaceAllString(text, "$1 $2")

	// Handle grouped punctuation like ... or !?
	groupRegex := regexp.MustCompile(`([.!?])\s+([.!?])`)
	text = groupRegex.ReplaceAllString(text, "$1$2")

	// Handle quotes
	text = formatQuotes(text)

	// Handle parentheses
	text = formatParentheses(text)

	return text
}

// formatQuotes handles single quote formatting
func formatQuotes(text string) string {
	// First handle double quotes and trim spaces inside them
	doubleQuoteRegex := regexp.MustCompile(`"\s*([^"]*?)\s*"`)
	text = doubleQuoteRegex.ReplaceAllStringFunc(text, func(match string) string {
		content := strings.Trim(match, "\"")
		content = strings.TrimSpace(content)
		return "\"" + content + "\""
	})

	// Then handle single quotes and trim spaces inside them
	singleQuoteRegex := regexp.MustCompile(`'\s*([^']*?)\s*'`)
	text = singleQuoteRegex.ReplaceAllStringFunc(text, func(match string) string {
		content := strings.Trim(match, "'")
		content = strings.TrimSpace(content)
		return "'" + content + "'"
	})

	return text
}

// formatParentheses handles spacing inside parentheses
func formatParentheses(text string) string {
	// Find text inside parentheses and trim excess spaces
	parenRegex := regexp.MustCompile(`\(\s*([^()]*?)\s*\)`)
	text = parenRegex.ReplaceAllStringFunc(text, func(match string) string {
		// Extract content between parentheses
		content := strings.Trim(match, "()")
		content = strings.TrimSpace(content)
		// Add a single space between words if needed
		content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")
		return "(" + content + ")"
	})

	return text
}

// fixArticles converts "a" to "an" before vowels and 'h'
func fixArticles(text string) string {
	// Convert "a/A" to "an/AN" before vowel sounds (a, e, i, o, u, h)
	aToAnRegex := regexp.MustCompile(`\b([aA])\s+([aeiouAEIOUhH])`)
	text = aToAnRegex.ReplaceAllStringFunc(text, func(match string) string {
		parts := regexp.MustCompile(`\b([aA])\s+([aeiouAEIOUhH])`).FindStringSubmatch(match)
		if len(parts) == 3 {
			article := parts[1]
			nextChar := parts[2]
			if article == "A" {
				return "AN " + nextChar
			} else {
				return "an " + nextChar
			}
		}
		return match
	})

	// Convert "an/AN" to "a/A" before consonant sounds (everything except a,e,i,o,u,h)
	anToARegex := regexp.MustCompile(`\b([aA][nN])\s+([bcdfgjklmnpqrstvwxyzBCDFGJKLMNPQRSTVWXYZ])`)
	text = anToARegex.ReplaceAllStringFunc(text, func(match string) string {
		parts := regexp.MustCompile(`\b([aA][nN])\s+([bcdfgjklmnpqrstvwxyzBCDFGJKLMNPQRSTVWXYZ])`).FindStringSubmatch(match)
		if len(parts) == 3 {
			article := parts[1]
			nextChar := parts[2]
			if strings.ToUpper(article) == "AN" {
				return "A " + nextChar
			} else {
				return "a " + nextChar
			}
		}
		return match
	})

	return text
}
