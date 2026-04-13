package main

import (
	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
)

type AddBlockerCmd struct {
	TaskID    uint64 `help:"Task ID." required:""`
	BlockedBy uint64 `help:"ID of the task that blocks it." required:""`
}

func (c *AddBlockerCmd) Run(creds *Credentials) error {
	conn, err := connect(creds)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = taskspb.NewBlockersClient(conn).AddBlocker(authCtx(creds), &taskspb.AddBlockerRequest{
		TaskId:    c.TaskID,
		BlockedBy: c.BlockedBy,
	})
	return err
}
