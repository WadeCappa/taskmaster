package tui

import "github.com/WadeCappa/taskmaster/internal/types"

type maybeTasksLoadedEvent struct {
	result types.Result[[]taskEntry]
}

type maybeTaskDetailLoadedEvent struct {
	result types.Result[taskDetailLoadedEvent]
}

type taskDetailLoadedEvent struct {
	taskId uint64
	detail *taskDetail
}

type maybeAddendumCreatedEvent struct {
	result types.Result[uint64] // taskId that was updated
}

type maybeStatusSetEvent struct {
	result types.Result[uint64] // taskId that was updated
}
