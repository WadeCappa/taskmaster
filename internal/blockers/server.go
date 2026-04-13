package blockers

import (
	"context"
	"fmt"

	"github.com/WadeCappa/taskmaster/internal/auth"
	"github.com/WadeCappa/taskmaster/internal/model"
	"github.com/WadeCappa/taskmaster/internal/tasks"
	"github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
)

var _ taskspb.BlockersServer = &blockersServer{}

type blockersServer struct {
	taskspb.BlockersServer

	db       *BlockerData
	taskData *tasks.TaskData
	auth     *auth.Auth
}

func NewServer(
	db *BlockerData,
	taskData *tasks.TaskData,
	auth *auth.Auth,
) taskspb.BlockersServer {
	return &blockersServer{
		db:       db,
		taskData: taskData,
		auth:     auth,
	}
}

func (t *blockersServer) AddBlocker(
	ctx context.Context,
	request *taskspb.AddBlockerRequest,
) (*taskspb.AddBlockerResponse, error) {
	userId, err := t.auth.GetUserId(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user id: %w", err)
	}
	exists, err := t.taskData.TasksExist(
		ctx,
		userId,
		model.TaskId(request.GetTaskId()),
		model.TaskId(request.GetBlockedBy()),
	)
	if err != nil {
		return nil, fmt.Errorf("finding tasks before writing link: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("not found")
	}
	if err := t.db.AddBlocker(ctx, userId, model.TaskId(request.TaskId), model.TaskId(request.BlockedBy)); err != nil {
		return nil, fmt.Errorf("adding blocker: %w", err)
	}
	return &taskspb.AddBlockerResponse{}, nil
}

func (t *blockersServer) GetBlockers(
	ctx context.Context,
	request *taskspb.GetBlockersRequest,
) (*taskspb.GetBlockersResponse, error) {
	userId, err := t.auth.GetUserId(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user id: %w", err)
	}
	exists, err := t.taskData.TasksExist(ctx, userId, model.TaskId(request.GetTaskId()))
	if err != nil {
		return nil, fmt.Errorf("finding task for blockers: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("not found")
	}
	blockers, err := t.db.GetBlockers(ctx, userId, model.TaskId(request.TaskId))
	if err != nil {
		return nil, fmt.Errorf("getting blockers: %w", err)
	}
	return &taskspb.GetBlockersResponse{
		BlockedByTaskIds: blockers,
	}, nil
}

func (t *blockersServer) RemoveBlocker(
	ctx context.Context,
	request *taskspb.RemoveBlockerRequest,
) (*taskspb.RemoveBlockerResponse, error) {
	userId, err := t.auth.GetUserId(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user id: %w", err)
	}
	exists, err := t.taskData.TasksExist(
		ctx,
		userId,
		model.TaskId(request.GetTaskId()),
		model.TaskId(request.GetBlockedBy()),
	)
	if err != nil {
		return nil, fmt.Errorf("finding tasks: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("not found")
	}
	if err := t.db.RemoveBlocker(ctx, userId, model.TaskId(request.TaskId), model.TaskId(request.BlockedBy)); err != nil {
		return nil, fmt.Errorf("removing blocker: %w", err)
	}
	return &taskspb.RemoveBlockerResponse{}, nil
}
