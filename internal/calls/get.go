package calls

import (
	"context"
	"fmt"
	"io"

	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
)

func Get(
	ctx context.Context,
	client taskspb.TasksClient,
	tags []string,
	statusId taskspb.Status,
) ([]*taskspb.GetTasksResponse, error) {
	task, err := client.GetTasks(ctx, &taskspb.GetTasksRequest{
		Status: statusId,
		Tags:   tags,
	})
	if err != nil {
		return nil, fmt.Errorf("calling client: %w", err)
	}
	var tasks []*taskspb.GetTasksResponse
	for {
		res, err := task.Recv()
		if err == io.EOF {
			return tasks, nil
		}

		if err != nil {
			return nil, fmt.Errorf("could not receive next entry: %w", err)
		}
		tasks = append(tasks, res)
	}
	return tasks, nil
}
