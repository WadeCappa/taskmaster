package main

import (
	"fmt"

	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

type DescribeTaskCmd struct {
	TaskID uint64 `help:"Task ID." required:""`
}

func (c *DescribeTaskCmd) Run(creds *Credentials) error {
	conn, err := connect(creds)
	if err != nil {
		return err
	}
	defer conn.Close()
	resp, err := taskspb.NewTasksClient(conn).DescribeTask(authCtx(creds), &taskspb.DescribeTaskRequest{
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
