package addendums

import (
	"context"
	"fmt"

	"github.com/WadeCappa/taskmaster/internal/auth"
	"github.com/WadeCappa/taskmaster/internal/model"
	"github.com/WadeCappa/taskmaster/internal/tasks"
	"github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
)

var _ taskspb.AddendumsServer = &blockersServer{}

type blockersServer struct {
	taskspb.AddendumsServer

	db       *AddendumData
	taskData *tasks.TaskData
	auth     *auth.Auth
}

func NewServer(
	db *AddendumData,
	taskData *tasks.TaskData,
	auth *auth.Auth,
) taskspb.AddendumsServer {
	return &blockersServer{
		db:       db,
		taskData: taskData,
		auth:     auth,
	}
}

func (t *blockersServer) AddAddendum(
	ctx context.Context,
	request *taskspb.AddAddendumRequest,
) (*taskspb.AddAddendumResponse, error) {
	userId, err := t.auth.GetUserId(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user Id: %w", err)
	}
	exists, err := t.taskData.TasksExist(ctx, userId, model.TaskId(request.TaskId))
	if err != nil {
		return nil, fmt.Errorf("finding tasks before writing addendum: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("not found")
	}
	if err := t.db.AddAddendum(ctx, userId, model.TaskId(request.TaskId), request.Content); err != nil {
		return nil, fmt.Errorf("creating addendum: %w", err)
	}
	return &taskspb.AddAddendumResponse{}, nil
}
