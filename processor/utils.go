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
func findWordBefore(text string, patternPos int) (word string, start, end int, quoted bool, quoteChar byte) {
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
		quotedContent := text[start+1 : end-1] // Content between quotes
		quotedContent = strings.TrimSpace(quotedContent)
		if quotedContent != "" {
			// Return the word inside quotes, but keep original positions for reconstruction
			return quotedContent, start, end, true, quote
		}
	}

	// Normal word finding (no quotes)
	// First check if we're immediately after punctuation
	if end > 0 && isPunctuation(text[end-1]) {
		// Include the punctuation as part of the "word" for conversion patterns
		end--
		start = end
		// Continue finding the actual word before punctuation
		for start > 0 && isWordChar(text[start-1]) {
			start--
		}
		// If we found a word + punctuation, return it
		if start < end {
			return text[start : end+1], start, end + 1, false, 0
		}
		// If just punctuation, reset and try normal word finding
		start = end + 1
		end = start
	}

	// Standard word finding
	start = end
	for start > 0 && isWordChar(text[start-1]) {
		start--
	}

	if start < end {
		return text[start:end], start, end, false, 0
	}
	return "", -1, -1, false, 0
}

// isPunctuation checks if a character is punctuation
func isPunctuation(c byte) bool {
	return c == '.' || c == ',' || c == '!' || c == '?' || c == ':' || c == ';'
}

func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// findWordsBefore finds specified number of words before the pattern
func findWordsBefore(text string, patternPos int, count int) ([]string, [][]int, []bool, []byte) {
	words := []string{}
	positions := [][]int{}
	quotedFlags := []bool{}
	quoteChars := []byte{}

	pos := patternPos
	for i := 0; i < count && pos > 0; i++ {
		word, start, end, quoted, quoteChar := findWordBefore(text, pos)
		if word == "" {
			break
		}

		words = append([]string{word}, words...) // Prepend to maintain order
		positions = append([][]int{{start, end}}, positions...)
		quotedFlags = append([]bool{quoted}, quotedFlags...)
		quoteChars = append([]byte{quoteChar}, quoteChars...)
		pos = start
	}

	return words, positions, quotedFlags, quoteChars
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
