package database_test

import (
	"iter"
	"testing"
	"unique"

	"github.com/WadeCappa/taskmaster/internal/database"
	"github.com/WadeCappa/taskmaster/internal/types"
	"github.com/stretchr/testify/require"
)

const (
	TEST_USER_ID = 101
)

func TestPutAndGetTask(t *testing.T) {
	db := database.NewEphemeralDatabase()

	tag := database.Tag(unique.Make("some-tag"))
	task := makeInternalTask(
		t,
		WithStatus(t, database.Completed),
		WithTags(tag),
	)

	newId, err := db.Put(TEST_USER_ID, types.None[database.TaskId](), task)
	require.NoError(t, err)

	itr, err := db.Get(TEST_USER_ID, database.Completed)
	require.NoError(t, err)

	expected := map[database.TaskId]database.Task{
		newId: task,
	}
	verifyTaskFound(t, expected, itr)

	itr, err = db.Get(TEST_USER_ID, database.Tracking)
	require.NoError(t, err)
	verifyTaskNotFound(t, itr)

	itr, err = db.Get(TEST_USER_ID, database.Completed, database.Tag(unique.Make("unrecognized-tag")))
	require.NoError(t, err)
	verifyTaskNotFound(t, itr)

	itr, err = db.Get(TEST_USER_ID, database.Completed, tag)
	require.NoError(t, err)
	verifyTaskFound(t, expected, itr)

	itr, err = db.Get(TEST_USER_ID, database.Completed, database.Tag(unique.Make("unrecognized-tag")), tag)
	require.NoError(t, err)
	verifyTaskNotFound(t, itr)
}

func verifyTaskFound(
	t *testing.T,
	expected map[database.TaskId]database.Task,
	result iter.Seq2[database.TaskId, database.Task],
) {
	seen := map[database.TaskId]struct{}{}
	for taskId, task := range result {
		require.NotContains(t, seen, taskId)
		seen[taskId] = struct{}{}
		require.Contains(t, expected, taskId)
		require.Equal(t, task.ToWireType(), task.ToWireType())
	}
	require.Equal(t, len(expected), len(seen))
}

func verifyTaskNotFound(
	t *testing.T,
	result iter.Seq2[database.TaskId, database.Task],
) {
	for range result {
		require.Fail(t, "should not have had any elements")
	}
}
