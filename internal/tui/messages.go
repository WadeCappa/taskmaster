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
