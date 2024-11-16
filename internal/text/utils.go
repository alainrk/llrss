package text

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

func URLToID(url string) string {
	hasher := sha256.New()
	hasher.Write([]byte(url))
	hash := hasher.Sum(nil)

	// encoded := base64.URLEncoding.EncodeToString(hash)
	// return strings.TrimRight(encoded, "=")
	return strings.ToLower(hex.EncodeToString(hash))
}

func CleanDescription(input string) string {
	if input == "" {
		return ""
	}

	// Check if the input contains HTML
	if strings.Contains(strings.ToLower(input), "<") && strings.Contains(strings.ToLower(input), ">") {
		input = stripHTML(input)
	}

	// Normalize newlines
	input = strings.ReplaceAll(input, "\r\n", "\n")

	// Replace multiple spaces with single space
	input = regexp.MustCompile(`[ \t]+`).ReplaceAllString(input, " ")

	// Replace multiple newlines with single newline
	input = regexp.MustCompile(`\n+`).ReplaceAllString(input, "\n")

	// Clean up spaces around newlines
	input = regexp.MustCompile(`[ \t]*\n[ \t]*`).ReplaceAllString(input, "\n")

	// Trim spaces and newlines from start and end
	input = strings.TrimSpace(input)

	return input
}

func stripHTML(input string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(input))
	var buffer bytes.Buffer
	lastWasNewline := false

	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			break
		}

		switch tokenType {
		case html.TextToken:
			text := strings.TrimSpace(string(tokenizer.Text()))
			if text != "" {
				if !lastWasNewline && buffer.Len() > 0 {
					buffer.WriteString(" ")
				}
				buffer.WriteString(text)
				lastWasNewline = false
			}
		case html.StartTagToken, html.EndTagToken:
			tag, _ := tokenizer.TagName()
			tagName := strings.ToLower(string(tag))

			if tagName == "br" || tagName == "p" || tagName == "div" {
				if !lastWasNewline {
					buffer.WriteString("\n")
					lastWasNewline = true
				}
			}
		}
	}

	return buffer.String()
}
