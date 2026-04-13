package tasks

import (
	"context"
	"fmt"

	"github.com/WadeCappa/taskmaster/internal/auth"
	"github.com/WadeCappa/taskmaster/internal/model"
	"github.com/WadeCappa/taskmaster/internal/types"
	"github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
	"google.golang.org/grpc"
)

var _ taskspb.TasksServer = &tasksServer{}

type tasksServer struct {
	taskspb.TasksServer

	data *TaskData
	auth *auth.Auth
}

func NewServer(
	data *TaskData,
	auth *auth.Auth,
) taskspb.TasksServer {
	return &tasksServer{
		data: data,
		auth: auth,
	}
}

func (t *tasksServer) CreateTask(
	ctx context.Context,
	request *taskspb.CreateTaskRequest,
) (*taskspb.CreateTaskResponse, error) {
	userId, err := t.auth.GetUserId(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user Id: %w", err)
	}
	parentTask := types.None[model.TaskId]()
	if request.GetTask().GetParentTaskId() != 0 {
		parentTask = types.Some(model.TaskId(request.GetTask().GetParentTaskId()))
	}
	newTaskId, err := t.data.AddTask(
		ctx, userId, parentTask, request.GetTask().GetTitle(),
		request.GetTask().GetDescription(), model.Status(request.GetTask().GetStatus()),
	)
	if err != nil {
		return nil, fmt.Errorf("putting task id: %w", err)
	}
	return &taskspb.CreateTaskResponse{
		TaskId: uint64(newTaskId),
	}, nil
}

func (t *tasksServer) GetTasks(
	request *taskspb.GetTasksRequest,
	stream grpc.ServerStreamingServer[taskspb.GetTasksResponse],
) error {
	userId, err := t.auth.GetUserId(stream.Context())
	if err != nil {
		return fmt.Errorf("getting user Id: %w", err)
	}
	tasks, err := t.data.GetTasks(stream.Context(), userId, model.Status(request.GetStatus()))
	if err != nil {
		return fmt.Errorf("getting tasks: %w", err)
	}

	for _, t := range tasks {
		if err := stream.Send(&taskspb.GetTasksResponse{
			Task: t.ToWire(),
		}); err != nil {
			return fmt.Errorf("sending task: %w", err)
		}
	}
	return nil
}

func (t *tasksServer) DescribeTask(
	ctx context.Context,
	request *taskspb.DescribeTaskRequest,
) (*taskspb.DescribeTaskResponse, error) {
	userId, err := t.auth.GetUserId(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user Id: %w", err)
	}
	task, err := t.data.LoadTask(ctx, userId, model.TaskId(request.GetTaskId()))
	if err != nil {
		return nil, fmt.Errorf("finding task: %w", err)
	}
	addendums, err := t.data.LoadAddendums(ctx, userId, task.taskId)
	if err != nil {
		return nil, fmt.Errorf("loading addendums: %w", err)
	}
	blockers, err := t.data.LoadBlockers(ctx, userId, task.taskId)
	if err != nil {
		return nil, fmt.Errorf("loading blockers: %w", err)
	}
	wireBlockers := make([]uint64, len(blockers))
	for idx, b := range blockers {
		wireBlockers[idx] = uint64(b)
	}
	return &taskspb.DescribeTaskResponse{
		Task:      task.ToWire(),
		Addendums: ToWire(addendums),
		BlockedBy: wireBlockers,
	}, nil
}

func (t *tasksServer) SetTaskStatus(
	ctx context.Context,
	request *taskspb.SetTaskStatusRequest,
) (*taskspb.SetTaskStatusResponse, error) {
	userId, err := t.auth.GetUserId(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user Id: %w", err)
	}
	if err := t.data.SetStatus(ctx, userId, model.TaskId(request.GetTaskId()), model.Status(request.GetStatus())); err != nil {
		return nil, fmt.Errorf("setting status: %w", err)
	}
	return &taskspb.SetTaskStatusResponse{}, nil
}
