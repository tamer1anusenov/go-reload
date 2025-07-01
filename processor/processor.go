package processor

import (
	"regexp"
	"strings"
)

func ProcessText(text string) string {
	text = normalizeSpaces(text)

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		line = strings.TrimLeft(line, " ")
		lines[i] = line
	}
	text = strings.Join(lines, "\n")

	// First handle nested patterns like (cap(low)) and BIN(low)
	text = processNestedPatterns(text)

	// Then handle parenthesized character patterns like (L(low)O(low)W(low))
	text = processParenthesizedCharPatterns(text)

	// Handle adjacent character patterns like L(low)o(up)w(up)
	text = processAdjacentCharPatterns(text)

	// Handle patterns directly attached to words (no space)
	text = processNoSpacePatterns(text)

	// Handle special cases of adjacent patterns
	text = processAdjacentCasePatterns(text)

	// Process remaining patterns sequentially from LEFT TO RIGHT
	text = processAllPatterns(text)

	// Apply final formatting
	text = formatPunctuation(text)

	text = fixArticles(text)

	text = normalizeSpaces(text)
	lines = strings.Split(text, "\n")
	for i, line := range lines {
		line = strings.TrimLeft(line, " ")
		lines[i] = line
	}
	text = strings.Join(lines, "\n")

	return text
}

func processAllPatterns(text string) string {
	for {
		pattern, position := findLeftmostPattern(text)
		if position == -1 {
			break
		}
		text = applyAndRemovePattern(text, pattern, position)
	}

	return text
}

func findLeftmostPattern(text string) (string, int) {
	patterns := findAllPatterns(text)
	if len(patterns) == 0 {
		return "", -1
	}

	leftmost := patterns[0]
	for _, p := range patterns {
		if p.position < leftmost.position {
			leftmost = p
		}
	}

	return leftmost.text, leftmost.position
}

type PatternMatch struct {
	text     string
	position int
	command  string
	count    string
}

func findAllPatterns(text string) []PatternMatch {
	var patterns []PatternMatch

	simplePatternRegex := regexp.MustCompile(`\(\s*([hH][eE][xX]|[bB][iI][nN]|[uU][pP]|[lL][oO][wW]|[cC][aA][pP])\s*\)`)
	simpleMatches := simplePatternRegex.FindAllStringIndex(text, -1)
	for _, match := range simpleMatches {
		patternText := text[match[0]:match[1]]
		commandRaw := regexp.MustCompile(`([hH][eE][xX]|[bB][iI][nN]|[uU][pP]|[lL][oO][wW]|[cC][aA][pP])`).FindString(patternText)

		// Normalize command to lowercase
		command := strings.ToLower(commandRaw)

		patterns = append(patterns, PatternMatch{
			text:     patternText,
			position: match[0],
			command:  command,
		})
	}

	numberedRegex := regexp.MustCompile(`\(\s*([uU][pP]|[lL][oO][wW]|[cC][aA][pP])\s*,\s*(-?\d+)\s*\)`)
	matches := numberedRegex.FindAllStringIndex(text, -1)
	for _, match := range matches {
		patternText := text[match[0]:match[1]]
		cmdMatchesRaw := regexp.MustCompile(`([uU][pP]|[lL][oO][wW]|[cC][aA][pP])`).FindString(patternText)
		countMatches := regexp.MustCompile(`-?\d+`).FindString(patternText)

		// Normalize command to lowercase
		cmdMatches := strings.ToLower(cmdMatchesRaw)

		if strings.HasPrefix(countMatches, "-") {
			patterns = append(patterns, PatternMatch{
				text:     patternText,
				position: match[0],
				command:  "remove",
				count:    countMatches,
			})
		} else {
			patterns = append(patterns, PatternMatch{
				text:     patternText,
				position: match[0],
				command:  cmdMatches,
				count:    countMatches,
			})
		}
	}

	return patterns
}

func applyAndRemovePattern(text, pattern string, position int) string {
	var result string

	patternLower := strings.ToLower(pattern)

	switch {
	case strings.Contains(patternLower, "hex"):
		result = processHexAtPosition(text, position)
	case strings.Contains(patternLower, "bin"):
		result = processBinAtPosition(text, position)
	case strings.Contains(patternLower, "up") && !strings.Contains(patternLower, ","):
		result = processCaseAtPosition(text, position, "up", 1)
	case strings.Contains(patternLower, "low") && !strings.Contains(patternLower, ","):
		result = processCaseAtPosition(text, position, "low", 1)
	case strings.Contains(patternLower, "cap") && !strings.Contains(patternLower, ","):
		result = processCaseAtPosition(text, position, "cap", 1)
	case strings.Contains(patternLower, ","):
		result = processNumberedCasePattern(text, pattern, position)
	default:
		result = removePatternAt(text, pattern, position)
	}

	result = formatQuotes(result)

	return result
}

func processNoSpacePatterns(text string) string {
	noSpaceRegex := regexp.MustCompile(`(\w+)\(\s*([uU][pP]|[lL][oO][wW]|[cC][aA][pP])\s*\)`)

	type transformation struct {
		start int
		end   int
		word  string
	}
	var transformations []transformation

	matches := noSpaceRegex.FindAllStringSubmatchIndex(text, -1)
	if matches == nil {
		return text
	}

	for _, match := range matches {
		wordStart, wordEnd := match[2], match[3]
		caseTypeStart, caseTypeEnd := match[4], match[5]

		word := text[wordStart:wordEnd]
		caseTypeRaw := text[caseTypeStart:caseTypeEnd]

		// Normalize command to lowercase
		caseType := strings.ToLower(caseTypeRaw)

		var transformedWord string
		switch caseType {
		case "up":
			transformedWord = strings.ToUpper(word)
		case "low":
			transformedWord = strings.ToLower(word)
		case "cap":
			transformedWord = capitalize(word)
		default:
			transformedWord = word
		}

		transformations = append(transformations, transformation{
			start: wordStart,
			end:   match[1],
			word:  transformedWord,
		})
	}

	result := text
	for i := len(transformations) - 1; i >= 0; i-- {
		t := transformations[i]

		if i < len(transformations)-1 && !isPunctuation(result[t.end]) {
			result = result[:t.start] + t.word + " " + result[t.end:]
		} else {
			result = result[:t.start] + t.word + result[t.end:]
		}
	}

	return result
}

func processNestedPatterns(text string) string {
	// Phase 1: Iteratively flatten semantic nested patterns (e.g., (cap(hex)) -> (cap))
	// This regex matches (OUTER_CMD(INNER_CMD_PATTERN)) where INNER_CMD_PATTERN is a command in parentheses.
	nestedTransformApplyRegex := regexp.MustCompile(`\(([a-zA-Z]+)\s*(\(hex\)|\(bin\)|\(up\)|\(low\)|\(cap\))\)`)

	for {
		foundFlattening := false
		newText := nestedTransformApplyRegex.ReplaceAllStringFunc(text, func(fullMatch string) string {
			subMatches := nestedTransformApplyRegex.FindStringSubmatch(fullMatch)
			if len(subMatches) < 3 {
				return fullMatch // Should not happen if regex matched
			}

			outerCmdRaw := subMatches[1]  // e.g., "cap", "hex"
			innerPattern := subMatches[2] // e.g., "(hex)", "(up)"

			outerCmd := strings.ToLower(outerCmdRaw)
			innerCmd := strings.ToLower(strings.Trim(innerPattern, "()")) // Extract inner command (e.g., "hex" from "(hex)")

			var resultCmd string

			// Apply innerCmd to outerCmd if innerCmd is a case transformation
			if innerCmd == "up" || innerCmd == "low" || innerCmd == "cap" {
				resultCmd = applyCaseTransformation(outerCmd, innerCmd)
			} else if outerCmd == "hex" || outerCmd == "bin" {
				// If outer is hex/bin and inner is also hex/bin, or not a case transformation
				// E.g., (hex(bin)), (hex(up)) -> should result in (hex) as 'up' doesn't transform 'hex' numerically
				resultCmd = outerCmd
			} else {
				// Default for unhandled combinations, just pass through outer command
				resultCmd = outerCmd
			}

			// If the outer command was transformed, mark as found change
			if strings.ToLower(resultCmd) != strings.ToLower(outerCmd) {
				foundFlattening = true
			}

			return "(" + resultCmd + ")"
		})

		if !foundFlattening {
			break // No more nested command transformations applied in this pass
		}
		text = newText
	}

	// Phase 2: Handle patterns directly attached to words (no space) after flattening
	// This section is for patterns like WORD(COMMAND) e.g., BIN(low)
	commandWordRegex := regexp.MustCompile(`([hH][eE][xX]|[bB][iI][nN]|[uU][pP]|[lL][oO][wW]|[cC][aA][pP])\(\s*([uU][pP]|[lL][oO][wW]|[cC][aA][pP])\s*\)`)

	for {
		match := commandWordRegex.FindStringSubmatchIndex(text)
		if match == nil {
			break
		}
		word := text[match[2]:match[3]]        // This is the 'WORD' part of WORD(COMMAND)
		caseTypeRaw := text[match[4]:match[5]] // This is the 'COMMAND' part

		caseType := strings.ToLower(caseTypeRaw)
		wordLower := strings.ToLower(word)

		var transformedWord string
		switch caseType {
		case "up":
			transformedWord = strings.ToUpper(wordLower)
		case "low":
			transformedWord = strings.ToLower(wordLower)
		case "cap":
			transformedWord = capitalize(wordLower)
		default:
			transformedWord = wordLower
		}

		// Replace WORD(COMMAND) with TRANSFORMED_WORD
		replacement := transformedWord
		text = text[:match[0]] + replacement + text[match[1]:]
	}

	return text
}

func extractInnermostCommand(pattern string) string {
	commandPattern := regexp.MustCompile(`\(([A-Za-z]+)\)`)
	matches := commandPattern.FindAllStringSubmatch(pattern, -1)

	if len(matches) == 0 {
		return ""
	}

	// Reverting to original: take the last match, which is the innermost for simple (cmd) patterns
	innermost := matches[len(matches)-1][1]

	// Normalize command to lowercase for validation
	if isValidCommand(innermost) {
		return strings.ToLower(innermost)
	}

	return ""
}

func isValidCommand(cmd string) bool {
	cmd = strings.ToLower(cmd)
	validCommands := []string{"low", "up", "cap"}
	for _, valid := range validCommands {
		if cmd == valid {
			return true
		}
	}
	return false
}

func processAdjacentCasePatterns(text string) string {
	adjacentRegex := regexp.MustCompile(`([a-zA-Z])\(\s*([uU][pP]|[lL][oO][wW]|[cC][aA][pP])\s*\)`)

	for {
		matches := adjacentRegex.FindAllStringSubmatchIndex(text, -1)
		if len(matches) == 0 {
			break
		}

		leftmostMatch := matches[0]
		for _, match := range matches {
			if match[0] < leftmostMatch[0] {
				leftmostMatch = match
			}
		}

		match := leftmostMatch

		charStart, charEnd := match[2], match[3]
		caseTypeStart, caseTypeEnd := match[4], match[5]

		char := text[charStart:charEnd]
		caseTypeRaw := text[caseTypeStart:caseTypeEnd]

		// Normalize command to lowercase
		caseType := strings.ToLower(caseTypeRaw)

		var transformedChar string
		switch caseType {
		case "up":
			transformedChar = strings.ToUpper(char)
		case "low":
			transformedChar = strings.ToLower(char)
		case "cap":
			transformedChar = capitalize(char)
		default:
			transformedChar = char
		}

		patternEnd := match[1]
		text = text[:charStart] + transformedChar + text[patternEnd:]
	}

	return text
}

func processParenthesizedCharPatterns(text string) string {
	charPatternRegex := regexp.MustCompile(`\(([A-Za-z])\(\s*([uU][pP]|[lL][oO][wW]|[cC][aA][pP])\s*\)([A-Za-z])\(\s*([uU][pP]|[lL][oO][wW]|[cC][aA][pP])\s*\)([A-Za-z])\(\s*([uU][pP]|[lL][oO][wW]|[cC][aA][pP])\s*\)\)`)

	for {
		match := charPatternRegex.FindStringSubmatchIndex(text)
		if match == nil {
			break
		}

		char1 := text[match[2]:match[3]]
		caseRaw1 := text[match[4]:match[5]]
		char2 := text[match[6]:match[7]]
		caseRaw2 := text[match[8]:match[9]]
		char3 := text[match[10]:match[11]]
		caseRaw3 := text[match[12]:match[13]]

		// Normalize commands to lowercase
		case1 := strings.ToLower(caseRaw1)
		case2 := strings.ToLower(caseRaw2)
		case3 := strings.ToLower(caseRaw3)

		transformed1 := applyCaseTransformation(char1, case1)
		transformed2 := applyCaseTransformation(char2, case2)
		transformed3 := applyCaseTransformation(char3, case3)

		replacement := "(" + transformed1 + " " + transformed2 + " " + transformed3 + ")"

		text = text[:match[0]] + replacement + text[match[1]:]
	}

	return text
}

func applyCaseTransformation(word, caseType string) string {
	switch caseType {
	case "up":
		return strings.ToUpper(word)
	case "low":
		return strings.ToLower(word)
	case "cap":
		return capitalize(word)
	default:
		return word
	}
}

func processAdjacentCharPatterns(text string) string {
	charPatternRegex := regexp.MustCompile(`([a-zA-Z])\(\s*([uU][pP]|[lL][oO][wW]|[cC][aA][pP])\s*\)([a-zA-Z])\(\s*([uU][pP]|[lL][oO][wW]|[cC][aA][pP])\s*\)([a-zA-Z])\(\s*([uU][pP]|[lL][oO][wW]|[cC][aA][pP])\s*\)`)

	for {
		match := charPatternRegex.FindStringSubmatchIndex(text)
		if match == nil {
			break
		}

		char1 := text[match[2]:match[3]]
		caseRaw1 := text[match[4]:match[5]]
		char2 := text[match[6]:match[7]]
		caseRaw2 := text[match[8]:match[9]]
		char3 := text[match[10]:match[11]]
		caseRaw3 := text[match[12]:match[13]]

		// Normalize commands to lowercase
		case1 := strings.ToLower(caseRaw1)
		case2 := strings.ToLower(caseRaw2)
		case3 := strings.ToLower(caseRaw3)

		transformed1 := applyCaseTransformation(char1, case1)
		transformed2 := applyCaseTransformation(char2, case2)
		transformed3 := applyCaseTransformation(char3, case3)

		replacement := transformed1 + transformed2 + transformed3
		text = text[:match[0]] + replacement + text[match[1]:]
	}

	return text
}
