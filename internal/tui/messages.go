package tui

type tasksLoadedMsg struct {
	tasks []taskEntry
	err   error
}

type taskDetailLoadedMsg struct {
	taskId uint64
	detail *taskDetail
	err    error
}

