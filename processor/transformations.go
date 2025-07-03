package processor

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

func processHexAtPosition(text string, pos int) string {
	word, wordStart, _, _, _ := findWordBefore(text, pos)
	if word != "" {
		hasPunctuation := false
		punctuationChar := byte(0)
		punctuationPos := -1

		for i := 0; i < len(word); i++ {
			if isPunctuation(word[i]) {
				hasPunctuation = true
				punctuationChar = word[i]
				punctuationPos = i
				break
			}
		}

		hexPart := word
		if hasPunctuation && punctuationPos > 0 {
			hexPart = word[:punctuationPos]
		}

		if isHex(hexPart) {
			if hexValue, err := strconv.ParseUint(hexPart, 16, 64); err == nil {
				if hexValue <= 0x7FFFFFFFFFFFFFFF {
					if decimal, err := strconv.ParseInt(hexPart, 16, 64); err == nil {
						before := text[:wordStart]
						after := text[pos+5:]

						result := before + strconv.FormatInt(decimal, 10)
						if hasPunctuation {
							result += string(punctuationChar)
						}
						result += after

						return result
					}
				}
			}

			before := text[:wordStart]
			after := text[pos+5:]
			result := before + hexPart
			if hasPunctuation {
				result += string(punctuationChar)
			}
			result += after

			return result
		}
	}
	return removePatternAt(text, "(hex)", pos)
}

func processBinAtPosition(text string, pos int) string {
	word, wordStart, _, _, _ := findWordBefore(text, pos)
	if word != "" {
		hasPunctuation := false
		punctuationChar := byte(0)
		punctuationPos := -1

		for i := 0; i < len(word); i++ {
			if isPunctuation(word[i]) {
				hasPunctuation = true
				punctuationChar = word[i]
				punctuationPos = i
				break
			}
		}

		binPart := word
		if hasPunctuation && punctuationPos > 0 {
			binPart = word[:punctuationPos]
		}

		if isBin(binPart) {
			if decimal, err := strconv.ParseInt(binPart, 2, 64); err == nil {
				before := text[:wordStart]
				after := text[pos+5:]

				result := before + strconv.FormatInt(decimal, 10)
				if hasPunctuation {
					result += string(punctuationChar)
				}
				result += after

				return result
			}
		}
	}
	return removePatternAt(text, "(bin)", pos)
}

func processCaseAtPosition(text string, pos int, caseType string, count int) string {
	words, positions, quotedFlags, quoteChars := findWordsBefore(text, pos, count)

	// NEW LOGIC FOR PROBLEM 1: Apply to last non-numeric word if target is numeric and count is 1 for case transformation
	// Only apply this logic for single word case transformations (count == 1)
	if count == 1 && (caseType == "up" || caseType == "low" || caseType == "cap") && len(words) > 0 && isNumeric(words[0]) {
		// Try to find a non-numeric word on the line before the current numeric word's position
		targetWord, targetStart, targetEnd, targetQuoted, targetQuoteChar := findLastNonNumericWordOnLineBefore(text, positions[0][0])

		if targetWord != "" {
			// Update the first word in the 'words', 'positions', etc. slices to reflect the new target word
			words[0] = targetWord
			positions[0][0] = targetStart
			positions[0][1] = targetEnd
			quotedFlags[0] = targetQuoted
			quoteChars[0] = targetQuoteChar
		} else {
			// If no non-numeric word found to apply transformation, just remove the pattern
			return removePatternAt(text, fmt.Sprintf("(%s)", caseType), pos)
		}
	}

	if len(words) > 0 {
		transformedWords := make([]string, len(words))
		for i, word := range words {
			switch caseType {
			case "up":
				transformedWords[i] = strings.ToUpper(word)
			case "cap":
				transformedWords[i] = capitalize(word)
			case "low":
				transformedWords[i] = strings.ToLower(word)
			default:
				transformedWords[i] = word
			}
		}

		result := text[:positions[0][0]]

		for i, word := range transformedWords {
			if quotedFlags[i] {
				if quoteChars[i] == '(' {
					result += "(" + word + ")"
				} else {
					result += string(quoteChars[i]) + word + string(quoteChars[i])
				}
			} else {
				result += word
			}

			if i < len(positions)-1 {
				result += text[positions[i][1]:positions[i+1][0]]
			}
		}

		patternLen := 0
		if count == 1 {
			patternLen = len(caseType) + 2 // e.g., "(up)" = 4
		} else {
			patternLen = len(fmt.Sprintf("(%s, %d)", caseType, count))
		}
		result += text[positions[len(positions)-1][1]:pos]
		result += text[pos+patternLen:]

		return result
	}

	if count == 1 {
		return removePatternAt(text, fmt.Sprintf("(%s)", caseType), pos)
	}
	return removePatternAt(text, fmt.Sprintf("(%s, %d)", caseType, count), pos)
}

func processNumberedCasePattern(text, pattern string, position int) string {
	re := regexp.MustCompile(`\(\s*(up|low|cap)\s*,\s*(-?\d+)\s*\)`)
	matches := re.FindStringSubmatch(pattern)
	if len(matches) == 3 {
		caseType := matches[1]
		count, _ := strconv.Atoi(matches[2])

		if count <= 0 {
			return removePatternAt(text, pattern, position)
		}

		words, positions, quotedFlags, quoteChars := findWordsBeforeInLine(text, position, count)

		if len(words) > 0 {
			transformedWords := make([]string, len(words))
			for i, word := range words {
				switch caseType {
				case "cap":
					transformedWords[i] = capitalize(word)
				case "up":
					transformedWords[i] = strings.ToUpper(word)
				case "low":
					transformedWords[i] = strings.ToLower(word)
				}
			}

			result := text
			for i := len(words) - 1; i >= 0; i-- {
				wordStart := positions[i][0]
				wordEnd := positions[i][1]
				if quotedFlags[i] {
					if quoteChars[i] == '(' {
						result = result[:wordStart] + "(" + transformedWords[i] + ")" + result[wordEnd:]
					} else {
						quoteChar := quoteChars[i]
						result = result[:wordStart] + string(quoteChar) + transformedWords[i] + string(quoteChar) + result[wordEnd:]
					}
				} else {
					result = result[:wordStart] + transformedWords[i] + result[wordEnd:]
				}
			}

			result = removePatternAt(result, pattern, position)
			return result
		}
	}
	return removePatternAt(text, pattern, position)
}

func findWordsBeforeInLine(text string, patternPos int, count int) ([]string, [][]int, []bool, []byte) {
	words := []string{}
	positions := [][]int{}
	quotedFlags := []bool{}
	quoteChars := []byte{}

	lineStart := strings.LastIndex(text[:patternPos], "\n")
	if lineStart == -1 {
		lineStart = 0
	} else {
		lineStart++
	}

	lineText := text[lineStart:patternPos]

	wordRegex := regexp.MustCompile(`\b[a-zA-Z0-9_]+\b`)
	matches := wordRegex.FindAllStringSubmatch(lineText, -1)
	matchIndices := wordRegex.FindAllStringIndex(lineText, -1)

	startIdx := len(matches) - count
	if startIdx < 0 {
		startIdx = 0
	}

	for i := startIdx; i < len(matches); i++ {
		word := matches[i][0]
		absoluteStart := lineStart + matchIndices[i][0]
		absoluteEnd := lineStart + matchIndices[i][1]

		words = append(words, word)
		positions = append(positions, []int{absoluteStart, absoluteEnd})
		quotedFlags = append(quotedFlags, isQuoted(word))
		quoteChars = append(quoteChars, getQuoteChar(word))
	}

	return words, positions, quotedFlags, quoteChars
}

func normalizeNestedCommandParentheses(text string) string {
	// Regex to find a word character (a command) followed by one or more spaces
	// and then an opening parenthesis.
	re := regexp.MustCompile(`(\w+)\s+\(`)

	// Loop until no more changes can be made. This ensures that we process
	// all levels of nesting, from the inside out.
	for {
		originalText := text
		text = re.ReplaceAllString(text, "$1(")
		// If the string is stable (no changes in this iteration), we are done.
		if text == originalText {
			break
		}
	}
	return text
}

func normalizeSpaces(text string) string {
	text = formatQuotes(text)
	text = normalizeNestedCommandParentheses(text)
	text = formatParentheses(text)

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		// This regex cleans up extra spaces between regular words.
		wordSpaceRegex := regexp.MustCompile(`(\w+)\s{2,}(\w+)`)
		for {
			newLine := wordSpaceRegex.ReplaceAllString(line, "$1 $2")
			if newLine == line {
				break
			}
			line = newLine
		}
		lines[i] = line
	}

	return strings.Join(lines, "\n")
}

func formatPunctuation(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		line = normalizeSpaces(line)
		punctRegex := regexp.MustCompile(`\s*([.,!?:;]+)(\s*)(\w?)`)
		line = punctRegex.ReplaceAllStringFunc(line, func(match string) string {
			parts := punctRegex.FindStringSubmatch(match)
			punct := parts[1]
			followingChar := parts[3]
			if followingChar != "" {
				return punct + " " + followingChar
			}
			return punct + followingChar
		})

		ellipsisRegex := regexp.MustCompile(`\s*\.{3,}`)
		line = ellipsisRegex.ReplaceAllString(line, "...")

		line = handleConsecutivePunctuation(line)

		line = strings.TrimLeft(line, " ")

		lines[i] = line
	}

	return strings.Join(lines, "\n")
}

func handleConsecutivePunctuation(text string) string {

	punctRegex := regexp.MustCompile(`([.,!?:;])\s+([.,!?:;])`)
	for {
		newText := punctRegex.ReplaceAllString(text, "$1$2")
		if newText == text {
			break
		}
		text = newText
	}
	return text
}

func formatQuotes(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		line = formatQuotesInLine(line)
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

func formatQuotesInLine(line string) string {
	line = handleConsecutiveQuotes(line, '\'')
	line = handleConsecutiveQuotes(line, '"')
	line = formatQuoteType(line, '\'')
	line = formatQuoteType(line, '"')

	return line
}

func handleConsecutiveQuotes(line string, quoteChar rune) string {
	quoteStr := string(quoteChar)
	pattern := regexp.QuoteMeta(quoteStr) + `{3,}`
	consecutiveRegex := regexp.MustCompile(pattern)

	return consecutiveRegex.ReplaceAllStringFunc(line, func(match string) string {
		count := len(match)
		pairs := count / 2
		remainder := count % 2

		result := strings.Repeat(quoteStr+quoteStr+" ", pairs)
		if remainder > 0 {
			result += quoteStr
		} else if len(result) > 0 {
			result = result[:len(result)-1]
		}
		return result
	})
}

func formatParentheses(text string) string {
	parenRegex := regexp.MustCompile(`\(\s*([^()]*?)\s*\)`)
	text = parenRegex.ReplaceAllStringFunc(text, func(match string) string {
		content := strings.Trim(match, "()")
		content = strings.TrimSpace(content)

		content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")
		return "(" + content + ")"
	})

	return text
}

func fixArticles(text string) string {
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		sequenceRegex := regexp.MustCompile(`(\b(?:[aA]n?\s+)*)([aA]n?)\s+([a-zA-Z]\w*)`)

		line = sequenceRegex.ReplaceAllStringFunc(line, func(match string) string {
			parts := sequenceRegex.FindStringSubmatch(match)
			if len(parts) >= 4 {
				precedingArticles := parts[1]
				lastArticle := parts[2]
				targetWord := parts[3]
				lowerTargetWord := strings.ToLower(targetWord)
				if lowerTargetWord == "a" || lowerTargetWord == "an" {
					return match
				}

				firstChar := strings.ToLower(string(targetWord[0]))
				needsAn := firstChar == "a" || firstChar == "e" || firstChar == "i" ||
					firstChar == "o" || firstChar == "u" || firstChar == "h"

				var newLastArticle string
				if needsAn {
					if lastArticle == "A" {
						newLastArticle = "An"
					} else if lastArticle == "a" {
						newLastArticle = "an"
					} else {
						newLastArticle = lastArticle
					}
				} else {
					if lastArticle == "A" || lastArticle == "An" {
						newLastArticle = "A"
					} else {
						newLastArticle = "a"
					}
				}

				return precedingArticles + newLastArticle + " " + targetWord
			}
			return match
		})

		lines[i] = line
	}

	return strings.Join(lines, "\n")
}

func formatQuoteType(line string, quoteChar rune) string {
	runes := []rune(line)
	result := make([]rune, 0, len(runes))

	i := 0
	for i < len(runes) {
		if runes[i] == quoteChar {
			if i > 0 && unicode.IsLetter(runes[i-1]) {
				prevChar := unicode.ToLower(runes[i-1])
				if prevChar != 'n' && prevChar != 't' {
					if len(result) > 0 && result[len(result)-1] != ' ' {
						result = append(result, ' ')
					}
				}
			}

			i++

			closeQuotePos := -1
			for j := i; j < len(runes); j++ {
				if runes[j] == quoteChar {
					closeQuotePos = j
					break
				}
			}

			if closeQuotePos == -1 {
				result = append(result, quoteChar)
				result = append(result, runes[i:len(runes)]...)
				result = append(result, quoteChar)
				break
			} else {
				content := runes[i:closeQuotePos]

				contentStr := strings.TrimSpace(string(content))

				result = append(result, quoteChar)
				result = append(result, []rune(contentStr)...)
				result = append(result, quoteChar)

				nextPos := closeQuotePos + 1
				if nextPos < len(runes) && isWordChar(byte(runes[nextPos])) {
					result = append(result, ' ')
				}

				i = nextPos
			}
		} else {
			result = append(result, runes[i])
			i++
		}
	}

	return string(result)
}
