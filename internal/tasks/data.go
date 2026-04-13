package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/WadeCappa/taskmaster/internal/auth"
	"github.com/WadeCappa/taskmaster/internal/model"
	"github.com/WadeCappa/taskmaster/internal/store"
	"github.com/WadeCappa/taskmaster/internal/types"
	"github.com/jackc/pgx/v5"
)

type TaskData struct {
	psqlUrl string
}

func NewTaskData(psqlData string) *TaskData {
	return &TaskData{
		psqlUrl: psqlData,
	}
}

func (t *TaskData) AddTask(
	ctx context.Context,
	userId auth.UserId,
	parentId types.Option[model.TaskId],
	title string,
	description string,
	status model.Status,
) (newTaskId model.TaskId, err error) {
	if err := store.Call(ctx, t.psqlUrl, func(c *pgx.Conn) error {
		var dbParentId model.TaskId
		val, exists := parentId.Unwrap()
		if exists {
			dbParentId = val
		}
		if err := c.QueryRow(
			ctx,
			`insert into tasks (user_id, parent_task_id, title, description, status) 
			values ($1, $2, $3, $4, $5) 
			returning task_id`,
			userId, dbParentId, title, description, status,
		).Scan(&newTaskId); err != nil {
			return fmt.Errorf("query and scan for task insert: %w", err)
		}
		return nil
	}); err != nil {
		return newTaskId, fmt.Errorf("putting task into postgres: %w", err)
	}
	return newTaskId, nil
}

func (t *TaskData) GetTasks(
	ctx context.Context,
	userId auth.UserId,
	status model.Status,
) ([]task, error) {
	var tasks []task
	if err := store.Call(ctx, t.psqlUrl, func(c *pgx.Conn) error {
		rows, err := c.Query(
			ctx,
			`select task_id, parent_task_id, title, description, status, created_time 
			from tasks 
			where user_id = $1 and status = $2`,
			userId, status,
		)
		if err != nil {
			return fmt.Errorf("querying db for tasks: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var taskId model.TaskId
			var parentId model.TaskId
			var title string
			var description string
			var status model.Status
			var createdTime time.Time
			if err := rows.Scan(&taskId, &parentId, &title, &description, &status, &createdTime); err != nil {
				return fmt.Errorf("scanning task: %w", err)
			}
			taskParentId := types.None[model.TaskId]()
			if parentId != 0 {
				taskParentId = types.Some(parentId)
			}
			tasks = append(tasks, newTask(taskId, taskParentId, title, description, status, createdTime))
		}
		return rows.Err()
	}); err != nil {
		return nil, fmt.Errorf("looking for tasks: %w", err)
	}
	return tasks, nil
}

func (t *TaskData) LoadTask(
	ctx context.Context,
	userId auth.UserId,
	taskId model.TaskId,
) (task, error) {
	result, err := store.CallAndReturn(ctx, t.psqlUrl, func(c *pgx.Conn) (task, error) {
		var scannedTaskId model.TaskId
		var parentId model.TaskId
		var title string
		var description string
		var status model.Status
		var createdTime time.Time
		if err := c.QueryRow(
			ctx,
			`select task_id, parent_task_id, title, description, status, created_time
			from tasks
			where user_id = $1 and task_id = $2`,
			userId, taskId,
		).Scan(&scannedTaskId, &parentId, &title, &description, &status, &createdTime); err != nil {
			return task{}, fmt.Errorf("scanning task: %w", err)
		}
		taskParentId := types.None[model.TaskId]()
		if parentId != 0 {
			taskParentId = types.Some(parentId)
		}
		return newTask(scannedTaskId, taskParentId, title, description, status, createdTime), nil
	})
	if err != nil {
		return task{}, fmt.Errorf("reading task: %w", err)
	}
	return result, nil
}

func (t *TaskData) LoadAddendums(
	ctx context.Context,
	userId auth.UserId,
	taskId model.TaskId,
) ([]addendum, error) {
	var addendums []addendum
	if err := store.Call(ctx, t.psqlUrl, func(c *pgx.Conn) error {
		rows, err := c.Query(
			ctx,
			`select content, created_time from addendums where user_Id = $1 and task_id = $2`,
			userId, taskId,
		)
		if err != nil {
			return fmt.Errorf("querying for addendums on task: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var content string
			var createdTime time.Time
			if err := rows.Scan(&content, &createdTime); err != nil {
				return fmt.Errorf("scanning next row: %w", err)
			}
			addendums = append(addendums, newAddendum(content, createdTime))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("loading addendums from psql: %w", err)
	}
	return addendums, nil
}

func (t *TaskData) LoadBlockers(
	ctx context.Context,
	userId auth.UserId,
	taskId model.TaskId,
) ([]model.TaskId, error) {
	var blockers []model.TaskId
	if err := store.Call(ctx, t.psqlUrl, func(c *pgx.Conn) error {
		rows, err := c.Query(
			ctx,
			`select blocked_by from task_blocked_by where task_id = $1`,
			taskId,
		)
		if err != nil {
			return fmt.Errorf("querying for blockers: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var blocker model.TaskId
			if err := rows.Scan(&blocker); err != nil {
				return fmt.Errorf("scanning next row: %w", err)
			}
			blockers = append(blockers, blocker)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("loading all blockers: %w", err)
	}
	return blockers, nil
}

func (t *TaskData) SetStatus(
	ctx context.Context,
	userId auth.UserId,
	taskId model.TaskId,
	newStatus model.Status,
) error {
	if err := store.Call(ctx, t.psqlUrl, func(c *pgx.Conn) error {
		if _, err := c.Exec(
			ctx,
			`update tasks set status = $1 where task_id = $2 and user_id = $3`,
			newStatus, taskId, userId,
		); err != nil {
			return fmt.Errorf("setting status: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("accessing psql to set status: %w", err)
	}
	return nil
}

func (t *TaskData) TasksExist(
	ctx context.Context,
	userId auth.UserId,
	tasks ...model.TaskId,
) (bool, error) {
	exists, err := store.CallAndReturn(ctx, t.psqlUrl, func(c *pgx.Conn) (bool, error) {
		var totalRows int
		if err := c.QueryRow(
			ctx,
			`select count(*) from tasks where user_id = $1 and task_id = ANY($2)`,
			userId, tasks,
		).Scan(&totalRows); err != nil {
			return false, fmt.Errorf("checking task existence: %w", err)
		}
		return totalRows == len(tasks), nil
	})
	if err != nil {
		return false, fmt.Errorf("testing for existence: %w", err)
	}
	return exists, nil
}
