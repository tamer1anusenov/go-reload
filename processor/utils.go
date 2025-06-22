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

// WordPosition represents a word with its position in text
type WordPosition struct {
	word  string
	start int
	end   int
}

// findWordsAfter finds specified number of words after the pattern position
func findWordsAfter(text string, startPos int, count int) ([]WordPosition, error) {
	if startPos >= len(text) {
		return nil, fmt.Errorf("start position out of bounds")
	}

	var result []WordPosition
	pos := startPos

	// Skip any spaces immediately after the pattern
	for pos < len(text) && unicode.IsSpace(rune(text[pos])) {
		pos++
	}

	for i := 0; i < count && pos < len(text); i++ {
		// Find the start of next word
		wordStart := pos

		// Skip any non-word characters
		for wordStart < len(text) && !isWordChar(text[wordStart]) && text[wordStart] != '\'' && text[wordStart] != '"' && text[wordStart] != '(' {
			wordStart++
		}

		if wordStart >= len(text) {
			break
		}

		// Check for quoted text
		if wordStart < len(text) && (text[wordStart] == '\'' || text[wordStart] == '"') {
			quoteChar := text[wordStart]
			endQuotePos := wordStart + 1

			// Find the closing quote
			for endQuotePos < len(text) && text[endQuotePos] != quoteChar {
				endQuotePos++
			}

			if endQuotePos < len(text) && text[endQuotePos] == quoteChar {
				// Found quoted text
				quotedWord := strings.TrimSpace(text[wordStart+1 : endQuotePos])
				result = append(result, WordPosition{
					word:  quotedWord,
					start: wordStart + 1,
					end:   endQuotePos,
				})
				pos = endQuotePos + 1
				continue
			}
		}

		// Check for parenthesized text
		if wordStart < len(text) && text[wordStart] == '(' {
			parenCount := 1
			endParenPos := wordStart + 1

			// Find the closing parenthesis, accounting for nested parentheses
			for endParenPos < len(text) && parenCount > 0 {
				if text[endParenPos] == '(' {
					parenCount++
				} else if text[endParenPos] == ')' {
					parenCount--
				}
				endParenPos++
			}

			if parenCount == 0 {
				// Found parenthesized text
				parenWord := strings.TrimSpace(text[wordStart+1 : endParenPos-1])
				result = append(result, WordPosition{
					word:  parenWord,
					start: wordStart + 1,
					end:   endParenPos - 1,
				})
				pos = endParenPos
				continue
			}
		}

		// Regular word
		wordEnd := wordStart
		for wordEnd < len(text) && isWordChar(text[wordEnd]) {
			wordEnd++
		}

		if wordEnd > wordStart {
			word := text[wordStart:wordEnd]
			result = append(result, WordPosition{
				word:  word,
				start: wordStart,
				end:   wordEnd,
			})
			pos = wordEnd
		} else {
			// If we couldn't find a word, move to the next character
			pos++
		}

		// Skip spaces after the word
		for pos < len(text) && unicode.IsSpace(rune(text[pos])) {
			pos++
		}
	}

	return result, nil
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
func bytesToRunes(b []byte) []rune {
	r := make([]rune, len(b))
	for i, v := range b {
		r[i] = rune(v)
	}
	return r
}
