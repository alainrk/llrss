package text

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestURLToID(t *testing.T) {
	tests := []struct {
		checkFn  func(*testing.T, string)
		wantDiff map[string]struct{}
		name     string
		url      string
		wantSame bool
	}{
		{
			name: "simple url",
			url:  "https://example.com",
			checkFn: func(t *testing.T, got string) {
				// Should only contain URL-safe characters
				assert.NotContains(t, got, "+")
				assert.NotContains(t, got, "/")
				assert.NotContains(t, got, "=")
			},
			wantSame: true,
		},
		{
			name: "url with special chars",
			url:  "https://example.com/path?q=1&b=2#fragment",
			checkFn: func(t *testing.T, got string) {
				assert.NotContains(t, got, "+")
				assert.NotContains(t, got, "/")
				assert.NotContains(t, got, "=")
			},
			wantSame: true,
		},
		{
			name: "unicode url",
			url:  "https://例子.com/路徑",
			checkFn: func(t *testing.T, got string) {
				assert.NotContains(t, got, "+")
				assert.NotContains(t, got, "/")
				assert.NotContains(t, got, "=")
			},
			wantSame: true,
		},
		{
			name: "empty string",
			url:  "",
			checkFn: func(t *testing.T, got string) {
				assert.NotContains(t, got, "+")
				assert.NotContains(t, got, "/")
				assert.NotContains(t, got, "=")
			},
			wantSame: true,
		},
		{
			name: "uniqueness check",
			url:  "https://example.com",
			checkFn: func(t *testing.T, got string) {
				// This will be used with multiple different URLs
				// to ensure they generate different IDs
			},
			wantDiff: map[string]struct{}{
				"https://example.com":      {},
				"https://example.com/":     {},
				"https://example.com/path": {},
				"https://example2.com":     {},
				"http://example.com":       {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := URLToID(tt.url)

			// Run custom checks
			if tt.checkFn != nil {
				tt.checkFn(t, got)
			}

			// Check idempotency
			if tt.wantSame {
				assert.Equal(t, strings.TrimSpace(got), got, "same URL should generate same ID")
			}

			// Check uniqueness
			if tt.wantDiff != nil {
				ids := make(map[string]string)
				for url := range tt.wantDiff {
					id := URLToID(url)
					for existingURL, existingID := range ids {
						assert.NotEqual(t, id, existingID,
							"URLs should generate unique IDs - collision between %q and %q",
							url, existingURL)
					}
					ids[url] = id
				}
			}

			// Additional sanity checks
			assert.Equal(t, strings.TrimSpace(got), got, "ID should not have leading/trailing spaces")
			assert.Len(t, strings.Split(got, " "), 1, "ID should not contain spaces")
		})
	}
}
