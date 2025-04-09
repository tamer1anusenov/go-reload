package transformations

import (
	"regexp"
	"strconv"
	"strings"
)

// ConvertHexToDec converts a hexadecimal string to decimal
func ConvertHexToDec(hexStr string) (string, error) {
	dec, err := strconv.ParseInt(hexStr, 16, 64)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(dec, 10), nil
}

// ConvertBinToDec converts a binary string to decimal
func ConvertBinToDec(binStr string) (string, error) {
	dec, err := strconv.ParseInt(binStr, 2, 64)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(dec, 10), nil
}

// TransformHexBin processes all (hex) and (bin) markers in the text
func TransformHexBin(text string) string {
	markerRegex := regexp.MustCompile(`\((hex|bin)\)`)
	tokens := tokenizeText(text, `\((hex|bin)\)`)

	var result []string
	for i, token := range tokens {
		if markerRegex.MatchString(token) && i > 0 {
			// Get the previous word
			prevWord := result[len(result)-1]

			// Determine conversion type
			var converted string
			var err error

			if strings.Contains(token, "hex") {
				converted, err = ConvertHexToDec(prevWord)
			} else {
				converted, err = ConvertBinToDec(prevWord)
			}

			// Replace if conversion succeeded
			if err == nil {
				result[len(result)-1] = converted
			}
			continue
		}
		result = append(result, token)
	}

	return strings.Join(result, " ")
}

// tokenizeText is a helper shared with case.go
func tokenizeText(text string, markerPattern string) []string {
	markerRegex := regexp.MustCompile(markerPattern)
	matches := markerRegex.FindAllStringSubmatchIndex(text, -1)

	var parts []string
	lastEnd := 0
	for _, m := range matches {
		start, end := m[0], m[1]
		if start > lastEnd {
			parts = append(parts, text[lastEnd:start])
		}
		parts = append(parts, text[start:end])
		lastEnd = end
	}
	if lastEnd < len(text) {
		parts = append(parts, text[lastEnd:])
	}

	var tokens []string
	for _, part := range parts {
		if markerRegex.MatchString(part) {
			tokens = append(tokens, part)
		} else {
			tokens = append(tokens, strings.Fields(part)...)
		}
	}

	return tokens
}
