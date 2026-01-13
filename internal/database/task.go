package database

import (
	"errors"
	"time"

	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
)

type Task struct {
	name              string
	timeToComplete    time.Duration
	priority          Priority
	status            Status
	tags              []Tag
	prerequisites     []TaskId
	numberOfAddendums uint64
}

func FromWireType(wire *taskspb.Task) (Task, error) {
	var errs []error

	tags := make([]Tag, len(wire.GetTags()))
	for i, t := range wire.GetTags() {
		tags[i] = Tag(t)
	}
	if len(tags) == 0 {
		errs = append(errs, errors.New("task must have at least one tag"))
	}

	if wire.GetName() == "" {
		errs = append(errs, errors.New("task must have name"))
	}

	if wire.GetMinutesToComplete() == 0 {
		errs = append(errs, errors.New("task must have time to complete"))
	}

	status, err := StatusFromWire(wire.GetStatus())
	if err != nil {
		errs = append(errs, errors.New("task must have time to complete"))
	}

	if len(errs) > 0 {
		return Task{}, errors.Join(errs...)
	}

	prereqs := make([]TaskId, len(wire.GetPrerequisites()))
	for i, p := range wire.GetPrerequisites() {
		prereqs[i] = TaskId(p)
	}
	return Task{
		name:           wire.GetName(),
		timeToComplete: time.Duration(wire.MinutesToComplete * uint64(time.Minute)),
		priority:       Priority(wire.GetPriority()),
		status:         status,
		tags:           tags,
		prerequisites:  prereqs,
	}, nil
}

func TaskFromDb(
	attributes TaskAttributes,
	priority Priority,
	status Status,
) Task {
	return Task{
		name:           attributes.Name,
		timeToComplete: attributes.TimeToComplete,
		priority:       priority,
		status:         status,
	}
}

func (t *Task) ToWireType() *taskspb.Task {
	tags := make([]string, len(t.tags))
	for i, t := range t.tags {
		tags[i] = string(t)
	}
	prereqs := make([]uint64, len(t.prerequisites))
	for i, p := range t.prerequisites {
		prereqs[i] = uint64(p)
	}
	return &taskspb.Task{
		Name:              t.name,
		MinutesToComplete: uint64(t.timeToComplete.Minutes()),
		Priority:          taskspb.Priority(t.priority),
		Status:            taskspb.Status(t.status),
		Tags:              tags,
		Prerequisites:     prereqs,
		NumberOfAddendums: t.numberOfAddendums,
	}
}

func (t *Task) HasStatus(status Status) bool {
	return t.status == status
}

func (t *Task) HasAllTags(tags ...Tag) bool {
	if len(tags) == 0 {
		return true
	}

	expected := map[Tag]struct{}{}
	for _, t := range tags {
		expected[t] = struct{}{}
	}

	actual := map[Tag]struct{}{}
	for _, t := range t.tags {
		actual[t] = struct{}{}
	}

	for t := range expected {
		if _, exists := actual[t]; !exists {
			return false
		}
	}

	return true
}
