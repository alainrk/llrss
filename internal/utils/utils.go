package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const (
	APISearchDateFormat = "2006-01-02"
)

func ParseAPISearchDate(dateStr string) (time.Time, error) {
	return time.Parse(APISearchDateFormat, dateStr)
}

func ParseRSSDate(dateStr string) (time.Time, error) {
	// Common RSS date formats
	formats := []string{
		time.RFC1123Z,               // "Mon, 02 Jan 2006 15:04:05 -0700"
		time.RFC1123,                // "Mon, 02 Jan 2006 15:04:05 MST"
		time.RFC822Z,                // "02 Jan 06 15:04 -0700"
		time.RFC822,                 // "02 Jan 06 15:04 MST"
		"2006-01-02T15:04:05Z07:00", // ISO 8601
		"2006-01-02T15:04:05",       // ISO 8601 without timezone
		"Mon, 2 Jan 2006 15:04:05 MST",
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 02 Jan 2006 15:04:05",
		"2 Jan 2006 15:04:05 MST",
		"2 Jan 2006 15:04:05 -0700",
		"02 Jan 2006 15:04:05 MST",
		"02 Jan 2006 15:04:05 -0700",
		"2006-01-02 15:04:05",
		"January 2, 2006 15:04:05",
	}

	dateStr = strings.TrimSpace(dateStr)

	// Try each format
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	// If none of the formats match, try parsing with some modifications
	dateStr = strings.ReplaceAll(dateStr, ",", "")
	dateStr = strings.ReplaceAll(dateStr, "GMT", "UTC")

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	// If all parsing attempts fail, return an error
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

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
