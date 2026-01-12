package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/WadeCappa/taskmaster/internal/auth"
	"github.com/WadeCappa/taskmaster/internal/database"
	"github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
	"google.golang.org/grpc"
)

type tasksServer struct {
	taskspb.TasksServer

	db   *database.Database
	auth *auth.Auth
}

func NewServer(
	db *database.Database,
	auth *auth.Auth,
) taskspb.TasksServer {
	return &tasksServer{
		db:   db,
		auth: auth,
	}
}

func (s *tasksServer) PutTask(
	ctx context.Context,
	request *taskspb.PutTaskRequest,
) (*taskspb.PutTaskResponse, error) {
	userId, err := s.auth.GetUserId(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user Id: %w", err)
	}
	task, err := database.FromWireType(request.GetTask())
	if err != nil {
		return nil, fmt.Errorf("converting task to wire type: %w", err)
	}
	newTaskId, err := s.db.Put(ctx, userId, task)
	if err != nil {
		return nil, fmt.Errorf("putting task id: %w", err)
	}
	return &taskspb.PutTaskResponse{
		TaskId: uint64(newTaskId),
	}, nil

}

func (s *tasksServer) GetTasks(
	request *taskspb.GetTasksRequest,
	stream grpc.ServerStreamingServer[taskspb.GetTasksResponse],
) error {
	userId, err := s.auth.GetUserId(stream.Context())
	if err != nil {
		return fmt.Errorf("getting user Id: %w", err)
	}

	tags := make([]database.Tag, len(request.GetTags()))
	for i, t := range request.GetTags() {
		tags[i] = database.Tag(t)
	}

	tasks, err := s.db.Get(
		stream.Context(),
		userId,
		database.Status(request.GetStatus()),
		tags...,
	)
	if err != nil {
		return fmt.Errorf("finding task: %w", err)
	}

	for _, taskAndId := range tasks {
		stream.Send(&taskspb.GetTasksResponse{
			TaskId: uint64(taskAndId.First),
			Task:   taskAndId.Second.ToWireType(),
		})
	}
	return nil
}

func (s *tasksServer) DescribeTask(
	request *taskspb.DescribeTaskRequest,
	stream grpc.ServerStreamingServer[taskspb.DescribeTaskResponse],
) error {
	userId, err := s.auth.GetUserId(stream.Context())
	if err != nil {
		return fmt.Errorf("getting user Id: %w", err)
	}

	itr, err := s.db.Describe(stream.Context(), userId, database.TaskId(request.TaskId))
	if err != nil {
		return fmt.Errorf("finding task: %w", err)
	}

	for _, task := range itr {
		stream.Send(&taskspb.DescribeTaskResponse{
			TaskId: uint64(task.First),
			Task:   task.Second.ToWireType(),
		})
	}
	return nil
}

func (s *tasksServer) MarkTask(
	ctx context.Context,
	request *taskspb.MarkTaskRequest,
) (*taskspb.MarkTaskResponse, error) {
	userId, err := s.auth.GetUserId(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user Id: %w", err)
	}

	if request.GetContent() == "" {
		return nil, errors.New("received mark request without any content")
	}

	if err := s.db.Mark(
		ctx,
		userId,
		database.TaskId(request.GetTaskId()),
		request.GetContent(),
	); err != nil {
		return nil, fmt.Errorf("setting addendum for task: %w", err)
	}
	return &taskspb.MarkTaskResponse{}, nil
}
