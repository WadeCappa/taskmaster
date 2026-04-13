package main

import (
	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
)

type SetTaskStatusCmd struct {
	TaskID uint64 `help:"Task ID." required:""`
	Status uint32 `help:"Status: 0=OPEN 1=IN_PROGRESS 2=CLOSED." required:""`
}

func (c *SetTaskStatusCmd) Run(creds *Credentials) error {
	conn, err := connect(creds)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = taskspb.NewTasksClient(conn).SetTaskStatus(authCtx(creds), &taskspb.SetTaskStatusRequest{
		TaskId: c.TaskID,
		Status: taskspb.Status(c.Status),
	})
	return err
}
