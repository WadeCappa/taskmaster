package tui

import (
	"fmt"
	"time"

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
