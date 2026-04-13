package addendums

import (
	"context"
	"fmt"

	"github.com/WadeCappa/taskmaster/internal/auth"
	"github.com/WadeCappa/taskmaster/internal/model"
	"github.com/WadeCappa/taskmaster/internal/store"
	"github.com/jackc/pgx/v5"
)

type AddendumData struct {
	psqlUrl string
}

func NewAddendumData(psqlUrl string) *AddendumData {
	return &AddendumData{
		psqlUrl: psqlUrl,
	}
}

func (a *AddendumData) AddAddendum(
	ctx context.Context,
	userId auth.UserId,
	taskId model.TaskId,
	content string,
) error {
	if err := store.Call(ctx, a.psqlUrl, func(c *pgx.Conn) error {
		if _, err := c.Exec(
			ctx,
			"insert into addendums (user_id, task_id, content) values ($1, $2, $3)",
			userId, taskId, content,
		); err != nil {
			return fmt.Errorf("query and scan for addendum insert: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("putting addendum into postgres: %w", err)
	}
	return nil
}
