package parser

import (
	"regexp"
	"strings"
)

const scissorsLine = "# ------------------------ >8 ------------------------"

var trailerLine = regexp.MustCompile(`^(?:[A-Za-z][A-Za-z0-9-]*|BREAKING CHANGE): .+$`)

// Message is the part of a commit message shared by every header format.
type Message struct {
	Header   string
	Body     string
	Trailers string
}

// SplitMessage applies the project's comment and scissors convention to both
// draft files and stored commits, regardless of Git's cleanup mode.
func SplitMessage(message string) Message {
	message = strings.ReplaceAll(message, "\r\n", "\n")
	message = removeCommentLines(message)
	lines := strings.SplitN(message, "\n", 2)
	result := Message{Header: strings.TrimSuffix(lines[0], "\r")}
	if len(lines) > 1 {
		result.Body, result.Trailers = splitBodyAndTrailers(strings.TrimSpace(lines[1]))
	}
	return result
}

func splitBodyAndTrailers(text string) (string, string) {
	if text == "" {
		return "", ""
	}
	lines := strings.Split(text, "\n")
	start := len(lines)
	for start > 0 && trailerLine.MatchString(lines[start-1]) {
		start--
	}
	if start == len(lines) || start > 0 && strings.TrimSpace(lines[start-1]) != "" {
		return text, ""
	}
	return strings.TrimSpace(strings.Join(lines[:start], "\n")), strings.Join(lines[start:], "\n")
}

// removeCommentLines removes Git's editor hint lines. Git ignores lines whose
// first character is '#'. A scissors line marks the start of content Git
// discards, so all following lines must be ignored as well.
func removeCommentLines(message string) string {
	lines := strings.Split(message, "\n")
	filteredLines := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSuffix(line, "\r") == scissorsLine {
			break
		}
		if !strings.HasPrefix(line, "#") {
			filteredLines = append(filteredLines, line)
		}
	}
	return strings.Join(filteredLines, "\n")
}
