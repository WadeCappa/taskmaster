package tasks

import (
	"time"

	"github.com/WadeCappa/taskmaster/internal/model"
	"github.com/WadeCappa/taskmaster/internal/types"
	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type task struct {
	taskId      model.TaskId
	parentId    types.Option[model.TaskId]
	title       string
	description string
	status      model.Status
	createdTime time.Time
}

func newTask(
	taskId model.TaskId,
	parentId types.Option[model.TaskId],
	title string,
	description string,
	status model.Status,
	createdTime time.Time,
) task {
	return task{
		taskId:      taskId,
		parentId:    parentId,
		title:       title,
		description: description,
		status:      status,
		createdTime: createdTime,
	}
}

func (t task) ToWire() *taskspb.Task {
	parentId, _ := t.parentId.Unwrap()
	return &taskspb.Task{
		TaskId:       uint64(t.taskId),
		Title:        t.title,
		Description:  t.description,
		ParentTaskId: uint64(parentId),
		Status:       taskspb.Status(t.status),
		CreatedTime:  timestamppb.New(t.createdTime),
	}
}
