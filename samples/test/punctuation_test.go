package transformations_test

import (
	"as/transformations"
	"testing"
)

func TestPunctuation(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello , world !", "Hello, world!"},
		{"Wait ... really ?", "Wait... really?"},
		{"' hello '", "'hello'"},
		{"' multi word '", "'multi word'"},
		{"He said , ' hello ' , then left", "He said, 'hello', then left"},
	}

	for _, tc := range tests {
		got := transformations.FixPunctuation(tc.input)
		if got != tc.want {
			t.Errorf("FixPunctuation(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
