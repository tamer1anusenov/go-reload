package transformations

import (
	"regexp"
	"strings"
)

// FixPunctuation is the main function that applies all text transformation steps.
func FixPunctuation(text string) string {
	text = FixQuotes2(text)            // Step 1: Normalize quotes
	text = FixPunctuationGroups(text)  // Step 2: Normalize punctuation clusters
	text = FixGeneralPunctuation(text) // Step 3: General punctuation cleanup
	text = FixSpacing(text)            // Step 4: Final spacing adjustments
	return text
}

// FixQuotes2 normalizes spaces inside single quotes.
func FixQuotes2(text string) string {
	re := regexp.MustCompile(`'\s*(.*?)\s*'`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		content := re.FindStringSubmatch(match)[1]
		return "'" + strings.TrimSpace(content) + "'"
	})
}

// FixPunctuationGroups handles punctuation clusters like ellipsis and interrobangs.
func FixPunctuationGroups(text string) string {
	// Correct ellipsis variations (handles trailing spaces)
	text = regexp.MustCompile(`(\.\s*){2,}\.?`).ReplaceAllString(text, "...")

	// Handle interrobang variations
	text = regexp.MustCompile(`!\s*\?`).ReplaceAllString(text, "!?")
	text = regexp.MustCompile(`\?\s*!`).ReplaceAllString(text, "?!")
	return text
}

// FixGeneralPunctuation cleans up spaces around punctuation marks.
func FixGeneralPunctuation(text string) string {
	// Remove spaces before punctuation
	text = regexp.MustCompile(`\s+([,.!?:;])`).ReplaceAllString(text, "$1")

	// Ensure spacing after ellipsis
	text = regexp.MustCompile(`(\.\.\.)(\S)`).ReplaceAllString(text, "$1 $2")

	// Ensure correct spacing after punctuation **only when needed**
	re := regexp.MustCompile(`([.!?])(\S)`)
	text = re.ReplaceAllStringFunc(text, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) == 3 && !strings.ContainsAny(parts[2], ".!?") {
			return parts[1] + " " + parts[2]
		}
		return match
	})

	// Remove spaces after sentence-ending punctuation at the end of the string
	text = regexp.MustCompile(`([.!?])\s+$`).ReplaceAllString(text, "$1")

	return strings.TrimSpace(text)
}

// FixSpacing performs final spacing adjustments and removes double spaces.
func FixSpacing(text string) string {
	// Remove double spaces
	text = regexp.MustCompile(`\s{2,}`).ReplaceAllString(text, " ")

	// Final cleanup for extra spaces
	return strings.TrimSpace(text)
}

// CleanFinalSpacing removes unnecessary spaces around punctuation while keeping spaces between words.
func CleanFinalSpacing(text string) string {
	// Step 1: Remove spaces **before** punctuation marks (, . ! ? etc.)
	text = regexp.MustCompile(`\s+([,.!?])`).ReplaceAllString(text, "$1")

	// Step 2: Remove spaces **before closing quotes** (' " ” ’)
	text = regexp.MustCompile(`\s+([’"”])`).ReplaceAllString(text, "$1")

	// Step 3: Ensure no spaces **between punctuation and quotes**
	// Example: `"Hello! "` → `"Hello!"`
	text = regexp.MustCompile(`([,.!?])\s+([’"”])`).ReplaceAllString(text, "$1$2")

	// Step 4: Add a space **after punctuation** if it's missing (except before another punctuation mark)
	text = regexp.MustCompile(`([,.!?])([^’"”\s])`).ReplaceAllString(text, "$1 $2")

	// Step 5: Trim extra spaces at the end
	text = strings.TrimSpace(text)

	return text
}
