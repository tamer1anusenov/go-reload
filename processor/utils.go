package processor

import (
	"strings"
	"unicode"
)

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

	// Check if we're at a closing parenthesis
	if start > 0 && text[start-1] == ')' {
		start-- // Move past the closing parenthesis

		// Find the opening parenthesis, accounting for nested parentheses
		parenCount := 1
		for start > 0 && parenCount > 0 {
			start--
			if text[start] == '(' {
				parenCount--
			} else if text[start] == ')' {
				parenCount++
			}
		}

		// Extract content from inside parentheses
		if start < end-1 {
			// Include the parentheses in the positions for reconstruction
			parenContent := text[start+1 : end-1] // Content between parentheses
			parenContent = strings.TrimSpace(parenContent)
			if parenContent != "" {
				// Return the content with parentheses, marking it as quoted with '(' as the quote char
				return parenContent, start, end, true, '('
			}
		}
	}

	// FIXED: Skip punctuation when looking for words
	// If we're immediately after punctuation, skip it to find the actual word
	if end > 0 && isPunctuation(text[end-1]) {
		// Skip the punctuation to find the word before it
		for end > 0 && isPunctuation(text[end-1]) {
			end--
		}
		// Skip any spaces after punctuation
		for end > 0 && text[end-1] == ' ' {
			end--
		}
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
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
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

// capitalize capitalizes the first letter of the word
func capitalize(s string) string {
	if s == "" {
		return ""
	}

	runes := []rune(s)
	if len(runes) == 1 {
		return string(unicode.ToUpper(runes[0]))
	}

	return string(unicode.ToUpper(runes[0])) + strings.ToLower(string(runes[1:]))
}

// isQuoted checks if a word is quoted (surrounded by quotes or parentheses)
func isQuoted(word string) bool {
	if len(word) < 2 {
		return false
	}

	first, last := word[0], word[len(word)-1]
	return (first == '\'' && last == '\'') ||
		(first == '"' && last == '"') ||
		(first == '(' && last == ')')
}

// getQuoteChar returns the quote character used (or 0 if not quoted)
func getQuoteChar(word string) byte {
	if len(word) < 2 {
		return 0
	}

	first := word[0]
	last := word[len(word)-1]

	if (first == '\'' && last == '\'') ||
		(first == '"' && last == '"') ||
		(first == '(' && last == ')') {
		return first
	}

	return 0
}
