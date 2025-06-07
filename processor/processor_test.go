package processor

import (
	"fmt"
	"testing"
)

func TestConsecutivePatterns(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "This is so exciting (up, 2)(low, 2)",
			expected: "This is so exciting", // First (up, 2) makes "so exciting" -> "SO EXCITING", then (low, 2) makes it "so exciting"
		},
		{
			input:    "Hello WORLD (low)(up)",
			expected: "Hello WORLD", // First (low) makes "WORLD" -> "world", then (up) makes it "WORLD" again
		},
		{
			input:    "test TEST (cap)(up)",
			expected: "test TEST", // First (cap) makes "TEST" -> "Test", then (up) makes it "TEST" again
		},
		{
			input:    "test (up) test (low)",
			expected: "TEST test", // First (up) makes "test" -> "TEST", then (low) makes second "test" -> "test"
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case %d", i+1), func(t *testing.T) {
			result := ProcessText(tc.input)
			if result != tc.expected {
				t.Errorf("Expected: %q, Got: %q", tc.expected, result)
			}
		})
	}
}

func TestFormatting(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "(   ggg   gg   )",
			expected: "( ggg gg )", // Excess spaces inside parentheses should be trimmed
		},
		{
			input:    "Hello   world",
			expected: "Hello world", // Multiple spaces between words should be reduced to one
		},
		{
			input:    "(text) (   more   text   )",
			expected: "(text) ( more text )", // Test multiple parentheses
		},
		{
			input:    "DFGHJKL DFGHJKL ( low, 1 )",
			expected: "dfghjkl dfghjkl", // Pattern with spaces should work
		},
		{
			input:    "test ( up ) test",
			expected: "TEST test", // Simple pattern with spaces should work
		},
		{
			input:    "a 'higher' (up)",
			expected: "an 'HIGHER'", // Quoted words should keep their quotes when transformed
		},
		{
			input:    "test \"quoted\" (cap)",
			expected: "test \"Quoted\"", // Double quoted words should keep their quotes
		},
		{
			input:    "a 'apple' is good",
			expected: "an 'apple' is good", // Article handling with single quotes
		},
		{
			input:    "a \"elephant\" is big",
			expected: "an \"elephant\" is big", // Article handling with double quotes
		},
		{
			input:    "an 'banana' is yellow",
			expected: "a 'banana' is yellow", // Article handling with consonant in quotes
		},
		{
			input:    "I was sitting over    !? . there ,and then           BAMM !  !  !",
			expected: "I was sitting over!?. there, and then BAMM!!!", // Multiple spaces between words should be reduced
		},
		{
			input:    "I am exactly how they describe me: '    awesome ' asdasdas\nAs Elton John said: ' I am the most well-known homosexual in the world '\nThere it was. A amazing rock!",
			expected: "I am exactly how they describe me: 'awesome' asdasdas\nAs Elton John said: 'I am the most well-known homosexual in the world'\nThere it was. An amazing rock!", // Preserve newlines and spaces after quotes
		},
		{
			input:    "This has a quote: 'text' with space after",
			expected: "This has a quote: 'text' with space after", // Space after quote should be preserved
		},
		{
			input:    "I am exactly how they describe me: '    awesome 'asdasdas",
			expected: "I am exactly how they describe me: 'awesome' asdasdas", // Add space after quote when followed by a word
		},
		{
			input:    "As Elton John said: ' I am the most well-known homosexual in the world '\n\nThere it was. A amazing rock!",
			expected: "As Elton John said: 'I am the most well-known homosexual in the world'\n\nThere it was. An amazing rock!", // Preserve newlines after quotes
		},
		{
			input:    "a a a a a\nA A A A",
			expected: "an a an a an a\nAn A An A", // Preserve newlines between lines and use "An" not "AN"
		},
		{
			input:    "a(hex).(bin)",
			expected: "10.2", // Handle punctuation in hex and bin conversions
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Formatting Case %d", i+1), func(t *testing.T) {
			result := ProcessText(tc.input)
			if result != tc.expected {
				t.Errorf("Expected: %q, Got: %q", tc.expected, result)
			}
		})
	}
}
