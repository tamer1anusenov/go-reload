package processor

import (
	"regexp"
	"strings"
	"unicode"
)

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

func findWordBefore(text string, patternPos int) (word string, start, end int, quoted bool, quoteChar byte) {
	end = patternPos
	for end > 0 && text[end-1] == ' ' {
		end--
	}

	start = end

	if start > 0 && (text[start-1] == '\'' || text[start-1] == '"') {
		quote := text[start-1]
		start--

		for start > 0 && text[start-1] != quote {
			start--
		}
		if start > 0 && text[start-1] == quote {
			start--
		}

		quotedContent := text[start+1 : end-1]
		quotedContent = strings.TrimSpace(quotedContent)
		if quotedContent != "" {
			return quotedContent, start, end, true, quote
		}
	}

	if start > 0 && text[start-1] == ')' {
		start--

		parenCount := 1
		for start > 0 && parenCount > 0 {
			start--
			if text[start] == '(' {
				parenCount--
			} else if text[start] == ')' {
				parenCount++
			}
		}

		if start < end-1 {
			parenContent := text[start+1 : end-1]
			parenContent = strings.TrimSpace(parenContent)
			if parenContent != "" {
				return parenContent, start, end, true, '('
			}
		}
	}

	if end > 0 && isPunctuation(text[end-1]) {
		for end > 0 && isPunctuation(text[end-1]) {
			end--
		}
		for end > 0 && text[end-1] == ' ' {
			end--
		}
	}

	start = end
	for start > 0 && isWordChar(text[start-1]) {
		start--
	}

	if start < end {
		return text[start:end], start, end, false, 0
	}
	return "", -1, -1, false, 0
}

func isPunctuation(c byte) bool {
	return c == '.' || c == ',' || c == '!' || c == '?' || c == ':' || c == ';'
}

func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

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

		words = append([]string{word}, words...)
		positions = append([][]int{{start, end}}, positions...)
		quotedFlags = append([]bool{quoted}, quotedFlags...)
		quoteChars = append([]byte{quoteChar}, quoteChars...)
		pos = start
	}

	return words, positions, quotedFlags, quoteChars
}

func removePatternAt(text, pattern string, pos int) string {
	return text[:pos] + text[pos+len(pattern):]
}

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

func isQuoted(word string) bool {
	if len(word) < 2 {
		return false
	}

	first, last := word[0], word[len(word)-1]
	return (first == '\'' && last == '\'') ||
		(first == '"' && last == '"') ||
		(first == '(' && last == ')')
}

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

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func findLastNonNumericWordOnLineBefore(text string, searchEndPos int) (word string, start, end int, quoted bool, quoteChar byte) {
	lineStart := strings.LastIndex(text[:searchEndPos], "\n")
	if lineStart == -1 {
		lineStart = 0
	} else {
		lineStart++
	}

	lineText := text[lineStart:searchEndPos]

	wordRegex := regexp.MustCompile(`\b[a-zA-Z0-9_]+\b`)
	matches := wordRegex.FindAllStringSubmatch(lineText, -1)
	matchIndices := wordRegex.FindAllStringIndex(lineText, -1)

	for i := len(matches) - 1; i >= 0; i-- {
		word := matches[i][0]
		if !isNumeric(word) { // Found the last non-numeric word
			absoluteStart := lineStart + matchIndices[i][0]
			absoluteEnd := lineStart + matchIndices[i][1]
			return word, absoluteStart, absoluteEnd, isQuoted(word), getQuoteChar(word)
		}
	}
	return "", -1, -1, false, 0 // No non-numeric word found
}
