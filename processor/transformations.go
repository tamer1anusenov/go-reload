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
	// If conversion failed, just remove (hex)
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

		// Handle negative count (count words after the pattern)
		if count < 0 {
			// Convert to positive for processing
			absCount := -count

			// Find the words after the pattern
			wordsAfter, err := findWordsAfter(text, position+len(pattern), absCount)
			if err == nil && len(wordsAfter) > 0 {
				// Apply transformation to words
				for i := 0; i < len(wordsAfter); i++ {
					word := wordsAfter[i].word
					transformedWord := ""

					switch caseType {
					case "cap":
						caser := cases.Title(language.Und)
						transformedWord = caser.String(strings.ToLower(word))
					case "up":
						transformedWord = strings.ToUpper(word)
					case "low":
						transformedWord = strings.ToLower(word)
					default:
						transformedWord = word
					}

					// Replace the word with transformed word
					text = text[:wordsAfter[i].start] + transformedWord + text[wordsAfter[i].end:]

					// Adjust positions of subsequent words based on length change
					lengthDiff := len(transformedWord) - len(word)
					for j := i + 1; j < len(wordsAfter); j++ {
						wordsAfter[j].start += lengthDiff
						wordsAfter[j].end += lengthDiff
					}
				}

				// Remove the pattern
				return removePatternAt(text, pattern, position)
			}

			// If no words found after, just remove the pattern
			return removePatternAt(text, pattern, position)
		}

		// Handle positive count (original behavior - words before the pattern)
		words, positions, quotedFlags, quoteChars := findWordsBefore(text, position, count)
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

// formatPunctuation handles spacing around punctuation marks
func formatPunctuation(text string) string {
	// First normalize spaces throughout the text
	text = normalizeSpaces(text)

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
		// Convert "a/A" to "an/An" before vowel sounds (a, e, i, o, u, h)
		// Also handle quotes - match 'a' before quoted words starting with vowels
		aToAnRegex := regexp.MustCompile(`\b([aA])\s+((['"]?)([aeiouAEIOUhH]))`)
		line = aToAnRegex.ReplaceAllStringFunc(line, func(match string) string {
			parts := regexp.MustCompile(`\b([aA])\s+((['"]?)([aeiouAEIOUhH]))`).FindStringSubmatch(match)
			if len(parts) == 5 {
				article := parts[1]
				quote := parts[3]    // Capture the quote if present
				nextChar := parts[4] // The actual first character

				if article == "A" {
					return "An " + quote + nextChar // Use "An" instead of "AN"
				} else {
					return "an " + quote + nextChar
				}
			}
			return match
		})

		// Convert "an/AN" to "a/A" before consonant sounds (everything except a,e,i,o,u,h)
		// Also handle quotes - match 'an' before quoted words starting with consonants
		anToARegex := regexp.MustCompile(`\b([aA][nN])\s+((['"]?)([bcdfgjklmnpqrstvwxyzBCDFGJKLMNPQRSTVWXYZ]))`)
		line = anToARegex.ReplaceAllStringFunc(line, func(match string) string {
			parts := regexp.MustCompile(`\b([aA][nN])\s+((['"]?)([bcdfgjklmnpqrstvwxyzBCDFGJKLMNPQRSTVWXYZ]))`).FindStringSubmatch(match)
			if len(parts) == 5 {
				article := parts[1]
				quote := parts[3]    // Capture the quote if present
				nextChar := parts[4] // The actual first character

				if strings.ToUpper(article) == "AN" {
					if article == "AN" {
						return "A " + quote + nextChar
					} else {
						return "a " + quote + nextChar
					}
				}
			}
			return match
		})

		lines[i] = line
	}

	// Join the lines back with newlines
	return strings.Join(lines, "\n")
}
