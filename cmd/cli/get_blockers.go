package main

import (
	"fmt"

	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

type GetBlockersCmd struct {
	TaskID uint64 `help:"Task ID." required:""`
}

func (c *GetBlockersCmd) Run(creds *Credentials) error {
	conn, err := connect(creds)
	if err != nil {
		return err
	}
	defer conn.Close()
	resp, err := taskspb.NewBlockersClient(conn).GetBlockers(authCtx(creds), &taskspb.GetBlockersRequest{
		TaskId: c.TaskID,
	})
	if err != nil {
		return err
	}
	b, err := protojson.Marshal(resp)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}
