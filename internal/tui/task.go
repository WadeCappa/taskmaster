package tui

import taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"

type taskEntry struct {
	id       uint64
	name     string
	priority taskspb.Priority
	minutes  uint64
	tags     []string
	addCount uint64
}

func taskFromWire(getTaskResponse *taskspb.GetTasksResponse) taskEntry {
	task := getTaskResponse.GetTask()
	id := getTaskResponse.GetTaskId()
	return taskEntry{
		id:       id,
		name:     task.GetName(),
		priority: task.GetPriority(),
		minutes:  task.GetMinutesToComplete(),
		tags:     task.GetTags(),
		addCount: task.GetNumberOfAddendums(),
	}
}
