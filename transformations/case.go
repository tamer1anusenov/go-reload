package transformations

import (
	"regexp"
	"strconv"
	"strings"
)

func protectEllipsis(text string) (string, string) {
	placeholder := "%%ELLIPSIS%%"
	// Replace proper ellipsis with placeholder
	protected := strings.ReplaceAll(text, "...", placeholder)
	return protected, placeholder
}

func unprotectEllipsis(text string, placeholder string) string {
	return strings.ReplaceAll(text, placeholder, "...")
}

// capitalize returns the word with first letter uppercase and rest lowercase
func capitalize(word string) string {
	if len(word) == 0 {
		return word
	}
	return strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
}

// fixQuotes ensures correct placement of single quotes around words/phrases
func fixQuotes(text string) string {
	return regexp.MustCompile(`'\s*(.*?)\s*'`).ReplaceAllString(text, "'$1'")
}

// TransformCase processes all (up), (low), and (cap) markers in the text
func TransformCase(text string) string {
	// Ensure quotes are normalized
	text = fixQuotes(text)

	// Protect ellipsis so they are not split during tokenization
	protectedText, placeholder := protectEllipsis(text)

	markerRegex := regexp.MustCompile(`\((cap|up|low)(?:,\s*(\d+))?\)`)
	tokens := tokenizeText(protectedText, `\((cap|up|low)(?:,\s*\d+)?\)`)

	var result []string
	for _, token := range tokens {
		if markerRegex.MatchString(token) {
			matches := markerRegex.FindStringSubmatch(token)
			transformType := matches[1]
			num := 1
			if matches[2] != "" {
				n, err := strconv.Atoi(matches[2])
				if err == nil {
					num = n
				}
			}
			start := len(result) - num
			if start < 0 {
				start = 0
			}
			for i := start; i < len(result); i++ {
				// If token starts with a quote, ensure we capitalize the first letter inside the quotes.
				if strings.HasPrefix(result[i], "'") && len(result[i]) > 1 {
					// Remove the quote, capitalize, then re-add it.
					withoutQuote := result[i][1:]
					result[i] = "'" + capitalize(withoutQuote)
				} else {
					switch transformType {
					case "cap":
						result[i] = capitalize(result[i])
					case "up":
						result[i] = strings.ToUpper(result[i])
					case "low":
						result[i] = strings.ToLower(result[i])
					}
				}
			}
			continue
		}
		result = append(result, token)
	}

	// Rejoin tokens and then restore any ellipsis placeholders
	joined := strings.Join(result, " ")
	joined = unprotectEllipsis(joined, placeholder)
	return joined
}
