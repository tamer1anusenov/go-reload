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
						// Replace word and ALL consecutive (hex) patterns with single decimal
						before := text[:wordStart]
						after := text[pos+5:] // +5 for "(hex)"

						// Remove any consecutive (hex) patterns that follow
						after = regexp.MustCompile(`^(\s*\(hex\))*`).ReplaceAllString(after, "")

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

			// Remove any consecutive (hex) patterns that follow
			after = regexp.MustCompile(`^(\s*\(hex\))*`).ReplaceAllString(after, "")

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
				// Replace word and pattern with decimal
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
				caser := cases.Title(language.Und)
				transformedWords[i] = caser.String(strings.ToLower(word))
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
		words, positions := findWordsBeforeInLine(text, position, count)

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

			// Reconstruct text with transformed words (work backwards to preserve positions)
			result := text
			for i := len(words) - 1; i >= 0; i-- {
				wordStart := positions[i][0]
				wordEnd := positions[i][1]
				result = result[:wordStart] + transformedWords[i] + result[wordEnd:]
			}

			// Now remove the numbered pattern
			result = removePatternAt(result, pattern, position)
			return result
		}
	}
	return removePatternAt(text, pattern, position)
}

// findWordsBeforeInLine finds specified number of actual words before the pattern, ignoring parentheses
func findWordsBeforeInLine(text string, patternPos int, count int) ([]string, [][]int) {
	words := []string{}
	positions := [][]int{}

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
	}

	return words, positions
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

// formatQuotes handles single quote formatting
func formatQuotes(text string) string {
	// Process the text line by line to preserve newlines
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		// Handle consecutive quotes first (like '''''')
		// Count consecutive single quotes and format them as pairs with spaces
		singleQuoteConsecutiveRegex := regexp.MustCompile(`'{3,}`)
		line = singleQuoteConsecutiveRegex.ReplaceAllStringFunc(line, func(match string) string {
			count := len(match)
			pairs := count / 2
			remainder := count % 2

			result := strings.Repeat("'' ", pairs)
			if remainder > 0 {
				result += "'"
			} else if len(result) > 0 {
				// Remove trailing space from last pair
				result = result[:len(result)-1]
			}

			return result
		})

		// Handle consecutive double quotes similarly
		doubleQuoteConsecutiveRegex := regexp.MustCompile(`"{3,}`)
		line = doubleQuoteConsecutiveRegex.ReplaceAllStringFunc(line, func(match string) string {
			count := len(match)
			pairs := count / 2
			remainder := count % 2

			result := strings.Repeat("\"\" ", pairs)
			if remainder > 0 {
				result += "\""
			} else if len(result) > 0 {
				// Remove trailing space from last pair
				result = result[:len(result)-1]
			}

			return result
		})

		// Handle double quotes in this line
		doubleQuoteRegex := regexp.MustCompile(`"(\s*)([^"]*?)(\s*)"(\s*)`)
		line = doubleQuoteRegex.ReplaceAllStringFunc(line, func(match string) string {
			parts := regexp.MustCompile(`"(\s*)([^"]*?)(\s*)"(\s*)`).FindStringSubmatch(match)
			if len(parts) == 5 {
				content := parts[2]
				spaceAfter := parts[4]

				// If there was a space after the closing quote, preserve it
				if spaceAfter != "" {
					return "\"" + content + "\" "
				}

				// Check if the next character after the quote is alphanumeric
				pos := strings.Index(line, match) + len(match)
				if pos < len(line) && isWordChar(line[pos]) {
					return "\"" + content + "\" "
				}
				return "\"" + content + "\""
			}
			return match
		})

		// Handle single quotes in this line
		singleQuoteRegex := regexp.MustCompile(`'(\s*)([^']*?)(\s*)'(\s*)`)
		line = singleQuoteRegex.ReplaceAllStringFunc(line, func(match string) string {
			parts := regexp.MustCompile(`'(\s*)([^']*?)(\s*)'(\s*)`).FindStringSubmatch(match)
			if len(parts) == 5 {
				content := parts[2]
				spaceAfter := parts[4]

				// If there was a space after the closing quote, preserve it
				if spaceAfter != "" {
					return "'" + content + "' "
				}

				// Check if the next character after the quote is alphanumeric
				pos := strings.Index(line, match) + len(match)
				if pos < len(line) && isWordChar(line[pos]) {
					return "'" + content + "' "
				}
				return "'" + content + "'"
			}
			return match
		})

		lines[i] = line
	}

	// Join the lines back with newlines
	return strings.Join(lines, "\n")
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
		// Find all words in the line
		wordRegex := regexp.MustCompile(`\b\w+\b`)
		matches := wordRegex.FindAllStringSubmatch(line, -1)
		matchIndices := wordRegex.FindAllStringIndex(line, -1)

		// Process each word from right to left to avoid position shifts
		for j := len(matches) - 1; j >= 0; j-- {
			word := matches[j][0]
			wordStart := matchIndices[j][0]

			// Skip if this word is an article itself
			if strings.ToLower(word) == "a" || strings.ToLower(word) == "an" {
				continue
			}

			// Find the article immediately before this word (ignoring other articles)
			beforeWord := line[:wordStart]

			// Look for the last article before this word, skipping intervening articles
			articleRegex := regexp.MustCompile(`\b([aA]|[aA][nN])\s+(?:(?:[aA]|[aA][nN])\s+)*$`)
			articleMatch := articleRegex.FindStringSubmatch(beforeWord)

			if len(articleMatch) >= 2 {
				lastArticle := articleMatch[1]

				// Check if word starts with vowel sound
				firstChar := strings.ToLower(string(word[0]))
				needsAn := firstChar == "a" || firstChar == "e" || firstChar == "i" ||
					firstChar == "o" || firstChar == "u" || firstChar == "h"

				// Find the position of this article to replace
				articlePos := strings.LastIndex(beforeWord, lastArticle)
				if articlePos != -1 {
					var newArticle string
					if needsAn {
						if lastArticle == "A" {
							newArticle = "An"
						} else {
							newArticle = "an"
						}
					} else {
						if strings.ToUpper(lastArticle) == lastArticle {
							newArticle = "A"
						} else {
							newArticle = "a"
						}
					}

					// Only replace if it's different
					if lastArticle != newArticle {
						line = line[:articlePos] + newArticle + line[articlePos+len(lastArticle):]
					}
				}
			}
		}

		lines[i] = line
	}

	// Join the lines back with newlines
	return strings.Join(lines, "\n")
}
