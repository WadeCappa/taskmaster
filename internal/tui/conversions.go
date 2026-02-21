package tui

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
)

var statuses = []taskspb.Status{taskspb.Status_TRACKING, taskspb.Status_COMPLETED, taskspb.Status_BACKLOG}

func priorityLabel(p taskspb.Priority) string {
	switch p {
	case taskspb.Priority_DO_BEFORE_SLEEP:
		return "[!!]"
	case taskspb.Priority_DO_IMMEDIATELY:
		return "[! ]"
	case taskspb.Priority_SHOULD_DO:
		return "[~ ]"
	case taskspb.Priority_EVENTUALLY_DO:
		return "[  ]"
	default:
		return "[??]"
	}
}

func formatDuration(minutes uint64) string {
	if minutes == 0 {
		return "-"
	}
	d := time.Duration(minutes) * time.Minute
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh%dm", h, m)
}

// wordWrapLines wraps s to at most width runes per line, breaking at spaces.
func wordWrapLines(s string, width int) []string {
	if width <= 0 {
		return []string{s}
	}

	var result []string
	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{""}
	}

	line := ""
	lineLen := 0

	for _, word := range words {
		wordLen := utf8.RuneCountInString(word)
		if wordLen > width {
			// Hard-break long words
			if lineLen > 0 {
				result = append(result, line)
				line = ""
				lineLen = 0
			}
			runes := []rune(word)
			for len(runes) > 0 {
				take := width
				if take > len(runes) {
					take = len(runes)
				}
				result = append(result, string(runes[:take]))
				runes = runes[take:]
			}
			continue
		}
		if lineLen == 0 {
			line = word
			lineLen = wordLen
		} else if lineLen+1+wordLen <= width {
			line += " " + word
			lineLen += 1 + wordLen
		} else {
			result = append(result, line)
			line = word
			lineLen = wordLen
		}
	}
	if lineLen > 0 {
		result = append(result, line)
	}
	return result
}
