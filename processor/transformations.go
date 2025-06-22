package processor

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// processHexAtPosition converts hex number before position and removes pattern
func processHexAtPosition(text string, pos int) string {
	word, wordStart, _, _, _ := findWordBefore(text, pos)
	if word != "" {
		// Check if the word contains punctuation
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

		// If word has punctuation, split it and process only the hex part
		hexPart := word
		if hasPunctuation && punctuationPos > 0 {
			hexPart = word[:punctuationPos]
		}

		if isHex(hexPart) {
			// First check if hex value is within valid range using ParseUint
			if hexValue, err := strconv.ParseUint(hexPart, 16, 64); err == nil {
				// Check if it's within the valid range (not larger than 7FFFFFFFFFFFFFFF)
				if hexValue <= 0x7FFFFFFFFFFFFFFF {
					// Convert to decimal since it's within range
					if decimal, err := strconv.ParseInt(hexPart, 16, 64); err == nil {
						// Replace word with decimal but keep the pattern for now
						before := text[:wordStart]
						after := text[pos+5:] // +5 for "(hex)"

						// Remove the current (hex) pattern but NOT consecutive ones
						// This allows successive transformations to work

						// Add back the punctuation if it was present
						result := before + strconv.FormatInt(decimal, 10)
						if hasPunctuation {
							result += string(punctuationChar)
						}
						result += after

						return result
					}
				}
			}

			// If we reach here, either:
			// 1. Hex value is too large (> 7FFFFFFFFFFFFFFF)
			// 2. Hex value couldn't be parsed (malformed but passed isHex check)
			// In both cases, just remove the (hex) pattern and keep original hex
			before := text[:wordStart]
			after := text[pos+5:] // +5 for "(hex)"

			// Keep the original hex value
			result := before + hexPart
			if hasPunctuation {
				result += string(punctuationChar)
			}
			result += after

			return result
		}
	}
	// If conversion failed or word not found, just remove (hex)
	return removePatternAt(text, "(hex)", pos)
}

// processBinAtPosition converts binary number before position and removes pattern
func processBinAtPosition(text string, pos int) string {
	word, wordStart, _, _, _ := findWordBefore(text, pos)
	if word != "" {
		// Check if the word contains punctuation
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

		// If word has punctuation, split it and process only the bin part
		binPart := word
		if hasPunctuation && punctuationPos > 0 {
			binPart = word[:punctuationPos]
		}

		if isBin(binPart) {
			if decimal, err := strconv.ParseInt(binPart, 2, 64); err == nil {
				// Replace word with decimal but keep additional patterns
				before := text[:wordStart]
				after := text[pos+5:] // +5 for "(bin)"

				// Add back the punctuation if it was present
				result := before + strconv.FormatInt(decimal, 10)
				if hasPunctuation {
					result += string(punctuationChar)
				}
				result += after

				return result
			}
		}
	}
	// If conversion failed, just remove (bin)
	return removePatternAt(text, "(bin)", pos)
}

// processCaseAtPosition applies case transformation and removes pattern
func processCaseAtPosition(text string, pos int, caseType string, count int) string {
	words, positions, quotedFlags, quoteChars := findWordsBefore(text, pos, count)
	if len(words) > 0 {
		// Apply transformation to words
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

		// Reconstruct text with transformed words
		result := text[:positions[0][0]] // Text before first word

		// Add transformed words with original spacing and quotes if needed
		for i, word := range transformedWords {
			if quotedFlags[i] {
				if quoteChars[i] == '(' {
					// Handle parenthesized text
					result += "(" + word + ")"
				} else {
					// Handle quoted text
					result += string(quoteChars[i]) + word + string(quoteChars[i])
				}
			} else {
				result += word
			}

			if i < len(positions)-1 {
				// Add text between words
				result += text[positions[i][1]:positions[i+1][0]]
			}
		}

		// Add text after last word until pattern, then skip pattern
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

	// If no words found, just remove pattern
	if count == 1 {
		return removePatternAt(text, fmt.Sprintf("(%s)", caseType), pos)
	}
	return removePatternAt(text, fmt.Sprintf("(%s, %d)", caseType, count), pos)
}

func processNumberedCasePattern(text, pattern string, position int) string {
	// Updated regex to handle negative numbers
	re := regexp.MustCompile(`\(\s*(up|low|cap)\s*,\s*(-?\d+)\s*\)`)
	matches := re.FindStringSubmatch(pattern)
	if len(matches) == 3 {
		caseType := matches[1]
		count, _ := strconv.Atoi(matches[2])

		// If count is negative or zero, just remove the pattern and return the text unchanged
		if count <= 0 {
			return removePatternAt(text, pattern, position)
		}

		// Handle positive count - but limit to words in the current line only
		words, positions, quotedFlags, quoteChars := findWordsBeforeInLine(text, position, count)

		if len(words) > 0 {
			// Apply transformation to words
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

			// Reconstruct text with transformed words
			result := text
			for i := len(words) - 1; i >= 0; i-- {
				wordStart := positions[i][0]
				wordEnd := positions[i][1]
				// Handle quoted words and parenthesized text
				if quotedFlags[i] {
					if quoteChars[i] == '(' {
						// Handle parenthesized text
						result = result[:wordStart] + "(" + transformedWords[i] + ")" + result[wordEnd:]
					} else {
						// Handle quoted text
						quoteChar := quoteChars[i]
						result = result[:wordStart] + string(quoteChar) + transformedWords[i] + string(quoteChar) + result[wordEnd:]
					}
				} else {
					// Replace with transformed word
					result = result[:wordStart] + transformedWords[i] + result[wordEnd:]
				}
			}

			// Now remove the numbered pattern
			result = removePatternAt(result, pattern, position)
			return result
		}
	}
	return removePatternAt(text, pattern, position)
}

// findWordsBeforeInLine finds specified number of actual words before the pattern, ignoring parentheses
func findWordsBeforeInLine(text string, patternPos int, count int) ([]string, [][]int, []bool, []byte) {
	words := []string{}
	positions := [][]int{}
	quotedFlags := []bool{}
	quoteChars := []byte{}

	// Find the start of the current line
	lineStart := strings.LastIndex(text[:patternPos], "\n")
	if lineStart == -1 {
		lineStart = 0
	} else {
		lineStart++ // Move past the newline
	}

	// Extract text from line start to pattern position
	lineText := text[lineStart:patternPos]

	// Use regex to find all words (sequences of letters, numbers, underscores, etc.)
	wordRegex := regexp.MustCompile(`\b[a-zA-Z0-9_]+\b`)
	matches := wordRegex.FindAllStringSubmatch(lineText, -1)
	matchIndices := wordRegex.FindAllStringIndex(lineText, -1)

	// Take the last 'count' words
	startIdx := len(matches) - count
	if startIdx < 0 {
		startIdx = 0
	}

	for i := startIdx; i < len(matches); i++ {
		word := matches[i][0]
		// Adjust positions to be relative to the full text
		absoluteStart := lineStart + matchIndices[i][0]
		absoluteEnd := lineStart + matchIndices[i][1]

		words = append(words, word)
		positions = append(positions, []int{absoluteStart, absoluteEnd})
		quotedFlags = append(quotedFlags, isQuoted(word))
		quoteChars = append(quoteChars, getQuoteChar(word))
	}

	return words, positions, quotedFlags, quoteChars
}

// normalizeSpaces reduces multiple spaces to a single space throughout the text
func normalizeSpaces(text string) string {
	// First, handle spaces inside quotes and parentheses
	text = formatQuotes(text)
	text = formatParentheses(text)

	// Split text by newlines to preserve them
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		// Handle multiple spaces between words
		// Keep applying the regex until no more changes are made
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

	// Join lines back with newlines
	return strings.Join(lines, "\n")
}

func formatPunctuation(text string) string {
	// Process the text line by line to preserve newlines
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		// First normalize spaces within the line
		line = normalizeSpaces(line)

		// Handle regular punctuation: .,!?:;
		// Remove spaces before punctuation and ensure space after if followed by word
		// BUT ONLY within the same line - don't cross newline boundaries
		punctRegex := regexp.MustCompile(`\s*([.,!?:;]+)(\s*)(\w?)`)
		line = punctRegex.ReplaceAllStringFunc(line, func(match string) string {
			parts := punctRegex.FindStringSubmatch(match)
			punct := parts[1]
			followingChar := parts[3]
			if followingChar != "" {
				// If punctuation is followed by a word character, add a space
				return punct + " " + followingChar
			}
			return punct + followingChar
		})

		// Handle ellipsis (...)
		ellipsisRegex := regexp.MustCompile(`\s*\.{3,}`)
		line = ellipsisRegex.ReplaceAllString(line, "...")

		// Handle consecutive punctuation marks
		line = handleConsecutivePunctuation(line)

		// Remove any leading spaces that might have been added
		line = strings.TrimLeft(line, " ")

		lines[i] = line
	}

	// Join the lines back with newlines - this preserves the original newline structure
	return strings.Join(lines, "\n")
}

// handleConsecutivePunctuation handles consecutive punctuation marks
func handleConsecutivePunctuation(text string) string {
	// Handle consecutive punctuation marks (e.g., !!, ??, !?, etc.)
	// Keep them together without spaces
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

// formatQuotes handles single and double quote formatting
func formatQuotes(text string) string {
	// Process the text line by line to preserve newlines
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		line = formatQuotesInLine(line)
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

func formatQuotesInLine(line string) string {
	// First handle consecutive quotes (3 or more)
	line = handleConsecutiveQuotes(line, '\'')
	line = handleConsecutiveQuotes(line, '"')

	// Then handle quote pairing and formatting
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
			// Remove trailing space from last pair
			result = result[:len(result)-1]
		}
		return result
	})
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
	// Process the text line by line to preserve newlines
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		// Find pattern: (any articles and spaces)(last article)(space)(non-article word)
		// This regex captures everything before the last article, the last article, and the target word
		sequenceRegex := regexp.MustCompile(`(\b(?:[aA]n?\s+)*)([aA]n?)\s+([a-zA-Z]\w*)`)

		line = sequenceRegex.ReplaceAllStringFunc(line, func(match string) string {
			parts := sequenceRegex.FindStringSubmatch(match)
			if len(parts) >= 4 {
				precedingArticles := parts[1] // All articles before the last one (keep unchanged)
				lastArticle := parts[2]       // The article immediately before the word (change this)
				targetWord := parts[3]        // The actual word that determines a/an

				// Skip if the "target word" is actually an article
				lowerTargetWord := strings.ToLower(targetWord)
				if lowerTargetWord == "a" || lowerTargetWord == "an" {
					return match // Don't change anything
				}

				// Check if target word starts with vowel sound
				firstChar := strings.ToLower(string(targetWord[0]))
				needsAn := firstChar == "a" || firstChar == "e" || firstChar == "i" ||
					firstChar == "o" || firstChar == "u" || firstChar == "h"

				// Determine the correct article, preserving the case of the last article
				var newLastArticle string
				if needsAn {
					if lastArticle == "A" {
						newLastArticle = "An"
					} else if lastArticle == "a" {
						newLastArticle = "an"
					} else { // already "An" or "an"
						newLastArticle = lastArticle
					}
				} else {
					if lastArticle == "A" || lastArticle == "An" {
						newLastArticle = "A"
					} else { // "a" or "an"
						newLastArticle = "a"
					}
				}

				return precedingArticles + newLastArticle + " " + targetWord
			}
			return match
		})

		lines[i] = line
	}

	// Join the lines back with newlines
	return strings.Join(lines, "\n")
}

func formatQuoteType(line string, quoteChar rune) string {
	runes := []rune(line)
	result := make([]rune, 0, len(runes))

	i := 0
	for i < len(runes) {
		if runes[i] == quoteChar {
			// Check if we need to add space before the quote
			if i > 0 && unicode.IsLetter(runes[i-1]) {
				// Don't add space if the previous letter is 'n' or 't' (for contractions like don't, can't)
				prevChar := unicode.ToLower(runes[i-1])
				if prevChar != 'n' && prevChar != 't' {
					// Add space before the quote if not already there
					if len(result) > 0 && result[len(result)-1] != ' ' {
						result = append(result, ' ')
					}
				}
			}

			// Found a quote
			i++ // Move past the opening quote

			// Find the closing quote
			closeQuotePos := -1
			for j := i; j < len(runes); j++ {
				if runes[j] == quoteChar {
					closeQuotePos = j
					break
				}
			}

			if closeQuotePos == -1 {
				// No closing quote found, add one at the end
				// Add the opening quote and content
				result = append(result, quoteChar)
				result = append(result, runes[i:len(runes)]...)
				result = append(result, quoteChar)
				break
			} else {
				// Found matching quote pair
				content := runes[i:closeQuotePos]

				// Clean up content - remove leading/trailing spaces
				contentStr := strings.TrimSpace(string(content))

				// Add formatted quote pair
				result = append(result, quoteChar)
				result = append(result, []rune(contentStr)...)
				result = append(result, quoteChar)

				// Check if we need a space after the closing quote
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
