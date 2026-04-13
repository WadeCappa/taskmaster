package blockers

import (
	"context"
	"fmt"

	"github.com/WadeCappa/taskmaster/internal/auth"
	"github.com/WadeCappa/taskmaster/internal/model"
	"github.com/WadeCappa/taskmaster/internal/store"
	"github.com/jackc/pgx/v5"
)

type BlockerData struct {
	psqlUrl string
}

func NewBlockerData(psqlUrl string) *BlockerData {
	return &BlockerData{
		psqlUrl: psqlUrl,
	}
}

func (b *BlockerData) AddBlocker(
	ctx context.Context,
	userId auth.UserId,
	taskId model.TaskId,
	blockedBy model.TaskId,
) error {
	if err := store.Call(ctx, b.psqlUrl, func(c *pgx.Conn) error {
		if _, err := c.Exec(
			ctx,
			`insert into task_blocked_by (task_id, blocked_by) values ($1, $2)`,
			taskId, blockedBy,
		); err != nil {
			return fmt.Errorf("inserting blocker: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("adding blocker to postgres: %w", err)
	}
	return nil
}

func (b *BlockerData) RemoveBlocker(
	ctx context.Context,
	userId auth.UserId,
	taskId model.TaskId,
	blockedBy model.TaskId,
) error {
	if err := store.Call(ctx, b.psqlUrl, func(c *pgx.Conn) error {
		if _, err := c.Exec(
			ctx,
			`delete from task_blocked_by where task_id=$1 and blocked_by=$2`,
			taskId, blockedBy,
		); err != nil {
			return fmt.Errorf("deleting blocker: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("removing blocker from postgres: %w", err)
	}
	return nil
}

func (b *BlockerData) GetBlockers(
	ctx context.Context,
	userId auth.UserId,
	taskId model.TaskId,
) ([]uint64, error) {
	var blockers []uint64
	if err := store.Call(ctx, b.psqlUrl, func(c *pgx.Conn) error {
		rows, err := c.Query(
			ctx,
			`select blocked_by from task_blocked_by where task_id=$1`,
			taskId,
		)
		if err != nil {
			return fmt.Errorf("querying blockers: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var id uint64
			if err := rows.Scan(&id); err != nil {
				return fmt.Errorf("scanning blocker id: %w", err)
			}
			blockers = append(blockers, id)
		}
		return rows.Err()
	}); err != nil {
		return nil, fmt.Errorf("getting blockers from postgres: %w", err)
	}
	return blockers, nil
}
