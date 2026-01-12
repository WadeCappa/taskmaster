package database_test

import (
	"testing"
	"unique"

	"github.com/WadeCappa/taskmaster/internal/database"
	"github.com/stretchr/testify/require"
)

func TestTaskFromWireType(t *testing.T) {
	wireType := makeWireTask()
	res, err := database.FromWireType(wireType)
	require.NoError(t, err)

	require.Equal(t, wireType, res.ToWireType())
}

func TestHasTags(t *testing.T) {
	tags := []database.Tag{
		database.Tag(unique.Make("some-tag")),
		database.Tag(unique.Make("some-other-tag")),
	}
	wireType := makeWireTask(WithTags(tags...))

	res, err := database.FromWireType(wireType)
	require.NoError(t, err)

	require.True(t, res.HasAllTags(tags[0]))
	require.True(t, res.HasAllTags(tags[1]))
	require.True(t, res.HasAllTags(tags...))
	require.True(t, res.HasAllTags())
	require.False(t, res.HasAllTags(database.Tag(unique.Make("unrecognized-tag"))))
	require.False(t, res.HasAllTags(tags[1], database.Tag(unique.Make("unrecognized-tag"))))
}
