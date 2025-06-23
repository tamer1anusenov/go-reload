package processor

import (
	"regexp"
	"strings"
)

func ProcessText(text string) string {
	text = normalizeSpaces(text)

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimLeft(line, " ")
	}
	text = strings.Join(lines, "\n")

	text = processParenthesizedCharPatterns(text)

	text = processNestedPatterns(text)

	text = processAdjacentCharPatterns(text)

	text = processNoSpacePatterns(text)

	text = processAdjacentCasePatterns(text)

	text = processAllPatterns(text)

	text = formatPunctuation(text)

	text = fixArticles(text)

	text = normalizeSpaces(text)
	lines = strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimLeft(line, " ")
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

	simplePatternRegex := regexp.MustCompile(`\(\s*(hex|bin|up|low|cap)\s*\)`)
	simpleMatches := simplePatternRegex.FindAllStringIndex(text, -1)
	for _, match := range simpleMatches {
		patternText := text[match[0]:match[1]]
		command := regexp.MustCompile(`(hex|bin|up|low|cap)`).FindString(patternText)
		patterns = append(patterns, PatternMatch{
			text:     patternText,
			position: match[0],
			command:  command,
		})
	}

	numberedRegex := regexp.MustCompile(`\(\s*(up|low|cap)\s*,\s*(-?\d+)\s*\)`)
	matches := numberedRegex.FindAllStringIndex(text, -1)
	for _, match := range matches {
		patternText := text[match[0]:match[1]]
		cmdMatches := regexp.MustCompile(`(up|low|cap)`).FindString(patternText)
		countMatches := regexp.MustCompile(`-?\d+`).FindString(patternText)

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

	switch {
	case pattern == "(hex)" || strings.Contains(pattern, "hex"):
		result = processHexAtPosition(text, position)
	case pattern == "(bin)" || strings.Contains(pattern, "bin"):
		result = processBinAtPosition(text, position)
	case pattern == "(up)" || (strings.Contains(pattern, "up") && !strings.Contains(pattern, ",")):
		result = processCaseAtPosition(text, position, "up", 1)
	case pattern == "(low)" || (strings.Contains(pattern, "low") && !strings.Contains(pattern, ",")):
		result = processCaseAtPosition(text, position, "low", 1)
	case pattern == "(cap)" || (strings.Contains(pattern, "cap") && !strings.Contains(pattern, ",")):
		result = processCaseAtPosition(text, position, "cap", 1)
	case strings.Contains(pattern, ","):
		result = processNumberedCasePattern(text, pattern, position)
	default:
		result = removePatternAt(text, pattern, position)
	}

	result = formatQuotes(result)

	return result
}

func processNoSpacePatterns(text string) string {
	noSpaceRegex := regexp.MustCompile(`(\w+)\(\s*(up|low|cap)\s*\)`)

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

		caseType := text[caseTypeStart:caseTypeEnd]

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
	nestedPattern := regexp.MustCompile(`\([A-Za-z]+\([A-Za-z]+(?:\([A-Za-z]+(?:\([A-Za-z]+(?:\([A-Za-z]+\)[A-Za-z]*)*\)[A-Za-z]*)*\)[A-Za-z]*)*\)[A-Za-z]*\)`)

	for {
		match := nestedPattern.FindString(text)
		if match == "" {
			break
		}

		innermost := extractInnermostCommand(match)
		if innermost == "" {
			text = strings.Replace(text, match, "", 1)
			continue
		}

		replacement := "(" + innermost + ")"
		text = strings.Replace(text, match, replacement, 1)
	}

	return text
}

func extractInnermostCommand(pattern string) string {
	commandPattern := regexp.MustCompile(`\(([A-Za-z]+)\)`)
	matches := commandPattern.FindAllStringSubmatch(pattern, -1)

	if len(matches) == 0 {
		return ""
	}

	innermost := matches[len(matches)-1][1]

	if isValidCommand(innermost) {
		return innermost
	}

	return ""
}

func isValidCommand(cmd string) bool {
	validCommands := []string{"low", "up", "cap", "LOW", "UP", "CAP"}
	for _, valid := range validCommands {
		if cmd == valid {
			return true
		}
	}
	return false
}

func processAdjacentCasePatterns(text string) string {
	adjacentRegex := regexp.MustCompile(`([a-zA-Z])\(\s*(up|low|cap)\s*\)`)

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
		caseType := text[caseTypeStart:caseTypeEnd]

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

	charPatternRegex := regexp.MustCompile(`\(([A-Za-z])\(\s*(up|low|cap)\s*\)([A-Za-z])\(\s*(up|low|cap)\s*\)([A-Za-z])\(\s*(up|low|cap)\s*\)\)`)

	for {
		match := charPatternRegex.FindStringSubmatchIndex(text)
		if match == nil {
			break
		}

		char1 := text[match[2]:match[3]]
		case1 := text[match[4]:match[5]]
		char2 := text[match[6]:match[7]]
		case2 := text[match[8]:match[9]]
		char3 := text[match[10]:match[11]]
		case3 := text[match[12]:match[13]]

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
	charPatternRegex := regexp.MustCompile(`([a-zA-Z])\(\s*(up|low|cap)\s*\)([a-zA-Z])\(\s*(up|low|cap)\s*\)([a-zA-Z])\(\s*(up|low|cap)\s*\)`)

	for {
		match := charPatternRegex.FindStringSubmatchIndex(text)
		if match == nil {
			break
		}

		char1 := text[match[2]:match[3]]
		case1 := text[match[4]:match[5]]
		char2 := text[match[6]:match[7]]
		case2 := text[match[8]:match[9]]
		char3 := text[match[10]:match[11]]
		case3 := text[match[12]:match[13]]

		transformed1 := applyCaseTransformation(char1, case1)
		transformed2 := applyCaseTransformation(char2, case2)
		transformed3 := applyCaseTransformation(char3, case3)

		replacement := transformed1 + transformed2 + transformed3
		text = text[:match[0]] + replacement + text[match[1]:]
	}

	return text
}
