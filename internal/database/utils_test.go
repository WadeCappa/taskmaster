package database_test

import (
	"testing"
	"unique"

	"github.com/WadeCappa/taskmaster/internal/database"
	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
	"github.com/stretchr/testify/require"
)

type taskOpt func(*taskspb.Task)

func makeInternalTask(t *testing.T, opts ...taskOpt) database.Task {
	wire := makeWireTask(opts...)
	res, err := database.FromWireType(wire)
	require.NoError(t, err)
	return res
}

func makeWireTask(opts ...taskOpt) *taskspb.Task {
	// Defaults
	t := &taskspb.Task{
		Name:              "test",
		MinutesToComplete: 4521,
		Priority:          taskspb.Priority_DO_BEFORE_SLEEP,
		Status:            taskspb.Status_BACKLOG,
		Tags:              []string{"some-tag", "some-other-tag"},
		Prerequisites:     []uint64{12, 453},
	}
	for _, o := range opts {
		o(t)
	}
	return t
}

func WithTags(tags ...database.Tag) taskOpt {
	return func(task *taskspb.Task) {
		task.Tags = nil
		for _, t := range tags {
			task.Tags = append(task.Tags, unique.Handle[string](t).Value())
		}
	}
}

func WithStatus(t *testing.T, s database.Status) taskOpt {
	return func(task *taskspb.Task) {
		r, err := database.StatusToWire(s)
		require.NoError(t, err)
		task.Status = r
	}
}
