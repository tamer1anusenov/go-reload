package processor

import (
	"fmt"
	"strings"
	"unicode"
)

// isVowel checks if a character is a vowel (case-insensitive).
func isVowel(char rune) bool {
	switch unicode.ToLower(char) {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	default:
		return false
	}
}

// isHex validates if the string is a valid hexadecimal (0-9, a-f, A-F).
func isHex(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') ||
			(c >= 'a' && c <= 'f') ||
			(c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// isBin validates if the string is a valid binary (only 0 or 1).
func isBin(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c != '0' && c != '1' {
			return false
		}
	}
	return true
}

// tokenize splits the text into tokens (words and punctuation).
func tokenize(text string) []string {
	var tokens []string
	var token strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			token.WriteRune(r)
		} else {
			if token.Len() > 0 {
				tokens = append(tokens, token.String())
				token.Reset()
			}
			if !unicode.IsSpace(r) {
				tokens = append(tokens, string(r))
			}
		}
	}
	if token.Len() > 0 {
		tokens = append(tokens, token.String())
	}
	return tokens
}

// reconstructText rebuilds text from tokens, adding spaces where appropriate.
func reconstructText(tokens []string) string {
	var sb strings.Builder
	for i, tok := range tokens {
		if i > 0 {
			// Add space if previous token and current token are both alphanumeric
			if isAlnum(tokens[i-1]) && isAlnum(tok) {
				sb.WriteRune(' ')
			}
		}
		sb.WriteString(tok)
	}
	return sb.String()
}

// isAlnum checks if the string is alphanumeric.
func isAlnum(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0
}

// findWordBefore finds the word immediately before the given position
func findWordBefore(text string, patternPos int) (word string, start, end int) {
	end = patternPos
	// Skip spaces before pattern
	for end > 0 && text[end-1] == ' ' {
		end--
	}

	start = end

	// Check if we're at a closing quote
	if start > 0 && (text[start-1] == '\'' || text[start-1] == '"') {
		quote := text[start-1]
		start-- // Move past the closing quote

		// Find the opening quote
		for start > 0 && text[start-1] != quote {
			start--
		}
		if start > 0 && text[start-1] == quote {
			start-- // Move past the opening quote
		}

		// Extract word from inside quotes
		if start+1 < end-1 {
			quotedContent := text[start+1 : end-1] // Content between quotes
			quotedContent = strings.TrimSpace(quotedContent)
			if quotedContent != "" {
				// Return the word inside quotes, but keep original positions for reconstruction
				return quotedContent, start, end
			}
		}
	}

	// Check if we're at a closing parenthesis
	if start > 0 && text[start-1] == ')' {
		start-- // Move past the closing parenthesis

		// Find the opening parenthesis
		parenCount := 1
		for start > 0 && parenCount > 0 {
			start--
			if text[start] == '(' {
				parenCount--
			} else if text[start] == ')' {
				parenCount++
			}
		}

		// Extract word from inside parentheses
		if start < end-1 {
			parenContent := text[start+1 : end-1] // Content between parentheses
			parenContent = strings.TrimSpace(parenContent)
			if parenContent != "" {
				// Return the word inside parentheses, but keep original positions for reconstruction
				return parenContent, start, end
			}
		}
	}

	// Normal word finding (no quotes or parentheses)
	start = end
	for start > 0 && isWordChar(text[start-1]) {
		start--
	}

	if start < end {
		return text[start:end], start, end
	}
	return "", -1, -1
}

func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// findWordsBefore finds specified number of words before the pattern
func findWordsBefore(text string, patternPos int, count int) ([]string, [][]int) {
	words := []string{}
	positions := [][]int{}
	wordCount := 0

	// First find all non-quoted words and quoted groups before the pattern
	var allWords []string
	var allPositions [][]int

	pos := patternPos
	for pos > 0 && wordCount < count {
		word, start, end := findWordBefore(text, pos)
		if word == "" {
			break
		}

		// Save the word/group and its position
		allWords = append([]string{word}, allWords...)
		allPositions = append([][]int{{start, end}}, allPositions...)

		// Count words in this segment
		if strings.Contains(word, " ") {
			// If it's a quoted/parenthesized group with spaces, count each word inside
			subWords := strings.Fields(word)
			wordCount += len(subWords)
		} else {
			// Single word
			wordCount++
		}

		// If we've found enough words, stop searching
		if wordCount >= count {
			break
		}

		pos = start
	}

	// Now take exactly 'count' words from the end
	remainingCount := count

	// Process words and positions from right to left (most recent first)
	for i := len(allWords) - 1; i >= 0 && remainingCount > 0; i-- {
		word := allWords[i]
		position := allPositions[i]

		if strings.Contains(word, " ") {
			// For quoted groups, split into individual words
			subWords := strings.Fields(word)

			// Take words from right to left
			for j := len(subWords) - 1; j >= 0 && remainingCount > 0; j-- {
				words = append([]string{subWords[j]}, words...)
				positions = append([][]int{position}, positions...)
				remainingCount--
			}
		} else {
			// Single word
			words = append([]string{word}, words...)
			positions = append([][]int{position}, positions...)
			remainingCount--
		}
	}

	return words, positions
}

// Helper function for minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// removePatternAt removes pattern at specific position
func removePatternAt(text, pattern string, pos int) string {
	return text[:pos] + text[pos+len(pattern):]
}

// getPatternLength returns the length of a case pattern
func getPatternLength(caseType string, count int) int {
	if count == 1 {
		return len(caseType) + 2 // e.g., "(up)" = 4
	}
	// For patterns like "(up, 2)" - calculate actual length
	return len(fmt.Sprintf("(%s, %d)", caseType, count))
}

// reconstructTextWithTransformedWords rebuilds text with transformed words
func reconstructTextWithTransformedWords(text string, positions [][]int, transformedWords []string, patternPos, patternLen int) string {
	if len(positions) == 0 {
		return text
	}

	result := text[:positions[0][0]] // Text before first word

	// Add transformed words with original spacing
	for i, word := range transformedWords {
		result += word
		if i < len(positions)-1 {
			// Add text between words
			result += text[positions[i][1]:positions[i+1][0]]
		}
	}

	// Add text after last word until pattern, then skip pattern
	result += text[positions[len(positions)-1][1]:patternPos]
	result += text[patternPos+patternLen:]

	return result
}
