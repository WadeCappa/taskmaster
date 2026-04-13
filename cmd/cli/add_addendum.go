package main

import (
	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
)

type AddAddendumCmd struct {
	TaskID  uint64 `help:"Task ID." required:""`
	Content string `help:"Addendum content." required:""`
}

func (c *AddAddendumCmd) Run(creds *Credentials) error {
	conn, err := connect(creds)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = taskspb.NewAddendumsClient(conn).AddAddendum(authCtx(creds), &taskspb.AddAddendumRequest{
		TaskId:  c.TaskID,
		Content: c.Content,
	})
	return err
}
