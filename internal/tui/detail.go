package tui

import (
	"time"

	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
)

type taskDetail struct {
	name      string
	priority  taskspb.Priority
	status    taskspb.Status
	minutes   uint64
	tags      []string
	addendums []addendumEntry
}

type addendumEntry struct {
	time    time.Time
	content string
}

func detailFromWire(describeTaskResponse *taskspb.DescribeTaskResponse) taskDetail {
	t := describeTaskResponse.GetTask()
	detail := taskDetail{
		name:     t.GetName(),
		priority: t.GetPriority(),
		status:   t.GetStatus(),
		minutes:  t.GetMinutesToComplete(),
		tags:     t.GetTags(),
	}
	addendums := make([]addendumEntry, len(describeTaskResponse.GetAddendum()))
	for index, a := range describeTaskResponse.GetAddendum() {
		addendums[index] = addendumEntry{
			time:    a.GetTimeCreated().AsTime(),
			content: a.GetContent(),
		}
	}
	detail.addendums = addendums
	return detail
}
