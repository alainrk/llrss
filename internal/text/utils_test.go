package text

import (
	"strings"
	"testing"
)

func TestCleanDescription(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "plain text without formatting",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "multiple spaces",
			input:    "Hello    world   with    many    spaces",
			expected: "Hello world with many spaces",
		},
		{
			name:     "multiple newlines",
			input:    "Hello\n\n\nworld\n\n\nwith\n\n\nmany\n\n\nnewlines",
			expected: "Hello\nworld\nwith\nmany\nnewlines",
		},
		{
			name:     "spaces and newlines",
			input:    "Hello   world\n\n    with   \n\n   mixed    formatting",
			expected: "Hello world\nwith\nmixed formatting",
		},
		{
			name:     "simple HTML",
			input:    "<p>Hello</p><p>world</p>",
			expected: "Hello\nworld",
		},
		{
			name: "complex HTML with links",
			input: `<p>Article URL: <a href="https://example.com">https://example.com</a></p>
                    <p>Comments URL: <a href="https://example.com">https://example.com</a></p>
                    <p>Points: 3</p>`,
			expected: "Article URL: https://example.com\nComments URL: https://example.com\nPoints: 3",
		},
		{
			name: "HTML with multiple spaces and newlines",
			input: `<p>First    paragraph    with    spaces</p>


                    <p>Second    paragraph    with    spaces</p>`,
			expected: "First paragraph with spaces\nSecond paragraph with spaces",
		},
		{
			name:     "text with leading/trailing whitespace",
			input:    "    \n\n   Hello world    \n\n    ",
			expected: "Hello world",
		},
		{
			name:     "HTML with nested tags",
			input:    `<div><p>Hello <strong>world</strong> with <em>formatting</em></p></div>`,
			expected: "Hello world with formatting",
		},
		{
			name: "Real world RSS example",
			input: `<p>Article URL: <a href="https://wooping.io/blog/2024/11/why-were-not-releasing-on-wp-org-yet/">https://wooping.io/blog/2024/11/why-were-not-releasing-on-wp-org-yet/</a></p>
                    <p>Comments URL: <a href="https://news.ycombinator.com/item?id=42132987">https://news.ycombinator.com/item?id=42132987</a></p>
                    <p>Points: 3</p>
                    <p># Comments: 1</p>`,
			expected: "Article URL: https://wooping.io/blog/2024/11/why-were-not-releasing-on-wp-org-yet/\nComments URL: https://news.ycombinator.com/item?id=42132987\nPoints: 3\n# Comments: 1",
		},
		{
			name: "Real world text example",
			input: `Where should I put the search box on my documentation site?
                    What text should I put inside the search box? What should
                    happen when I type stuff into the search box? What should the
                    search results page look like?`,
			expected: "Where should I put the search box on my documentation site?\nWhat text should I put inside the search box? What should\nhappen when I type stuff into the search box? What should the\nsearch results page look like?",
		},
		{
			name:     "HTML with br tags",
			input:    "First line<br>Second line<br>Third line",
			expected: "First line\nSecond line\nThird line",
		},
		{
			name:     "Mixed case HTML",
			input:    "<P>Upper case</P><p>Lower case</p><BR>Line break",
			expected: "Upper case\nLower case\nLine break",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanDescription(tt.input)
			if got != tt.expected {
				t.Errorf("\nname: %s\ngot:\n%s\nwant:\n%s\n",
					tt.name,
					strings.ReplaceAll(got, " ", "·"),
					strings.ReplaceAll(tt.expected, " ", "·"))
			}
		})
	}
}
