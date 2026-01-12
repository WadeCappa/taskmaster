package database

import (
	"fmt"

	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
)

type Status int

const (
	Tracking Status = iota
	Completed
	Backlog
)

func StatusFromWire(status taskspb.Status) (Status, error) {
	switch status {
	case taskspb.Status_BACKLOG:
		return Backlog, nil
	case taskspb.Status_COMPLETED:
		return Completed, nil
	case taskspb.Status_TRACKING:
		return Tracking, nil
	}
	return Tracking, fmt.Errorf("unrecognized status of %d", status.Number())
}

func StatusToWire(status Status) (taskspb.Status, error) {
	switch status {
	case Tracking:
		return taskspb.Status_TRACKING, nil
	case Completed:
		return taskspb.Status_COMPLETED, nil
	case Backlog:
		return taskspb.Status_BACKLOG, nil
	}
	return taskspb.Status_TRACKING, fmt.Errorf("unrecognized status of %d", status)
}
