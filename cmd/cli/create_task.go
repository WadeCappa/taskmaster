package main

import (
	"fmt"

	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
)

type CreateTaskCmd struct {
	Title       string `help:"Task title." required:""`
	Description string `help:"Task description."`
	ParentID    uint64 `help:"Parent task ID (for sub-tasks)." name:"parent-task-id"`
	Status      uint32 `help:"Status: 0=OPEN 1=IN_PROGRESS 2=CLOSED." default:"0"`
}

func (c *CreateTaskCmd) Run(creds *Credentials) error {
	conn, err := connect(creds)
	if err != nil {
		return err
	}
	defer conn.Close()
	resp, err := taskspb.NewTasksClient(conn).CreateTask(authCtx(creds), &taskspb.CreateTaskRequest{
		Task: &taskspb.Task{
			Title:        c.Title,
			Description:  c.Description,
			ParentTaskId: c.ParentID,
			Status:       taskspb.Status(c.Status),
		},
	})
	if err != nil {
		return err
	}
	fmt.Println(resp.TaskId)
	return nil
}
