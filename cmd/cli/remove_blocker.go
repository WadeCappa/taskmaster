package main

import (
	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
)

type RemoveBlockerCmd struct {
	TaskID    uint64 `help:"Task ID." required:""`
	BlockedBy uint64 `help:"ID of the blocking task to remove." required:""`
}

func (c *RemoveBlockerCmd) Run(creds *Credentials) error {
	conn, err := connect(creds)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = taskspb.NewBlockersClient(conn).RemoveBlocker(authCtx(creds), &taskspb.RemoveBlockerRequest{
		TaskId:    c.TaskID,
		BlockedBy: c.BlockedBy,
	})
	return err
}
