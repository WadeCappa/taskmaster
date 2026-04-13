package main

import (
	"fmt"
	"io"

	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

type GetTasksCmd struct {
	Status uint32 `help:"Status: 0=OPEN 1=IN_PROGRESS 2=CLOSED." default:"0"`
}

func (c *GetTasksCmd) Run(creds *Credentials) error {
	conn, err := connect(creds)
	if err != nil {
		return err
	}
	defer conn.Close()
	stream, err := taskspb.NewTasksClient(conn).GetTasks(authCtx(creds), &taskspb.GetTasksRequest{
		Status: taskspb.Status(c.Status),
	})
	if err != nil {
		return err
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		b, err := protojson.Marshal(resp)
		if err != nil {
			return err
		}
		fmt.Println(string(b))
	}
}
