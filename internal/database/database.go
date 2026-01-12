package database

import (
	"context"
	"fmt"
	"time"

	"github.com/WadeCappa/taskmaster/internal/auth"
	"github.com/WadeCappa/taskmaster/internal/store"
	"github.com/WadeCappa/taskmaster/internal/types"
	"github.com/jackc/pgx/v5"
)

const (
	insertTaskQuery           = "insert into tasks (task_id, user_id, fields, priority, status) values (nextval('task_ids'), $1, $2, $3, $4) returning task_id"
	getTasksWithTags          = "select t.task_id, t.fields, t.priority, t.status from tasks t join tags_to_tasks ttt on t.task_id = ttt.task_id join tags tg on tg.tag_id = ttt.tag_id where t.user_id = $1 and tg.name = any ($2) and t.status = $3 group by t.task_id having count(distinct tg.tag_id) = cardinality($2) order by priority;"
	getTasksWithIgnoringTasks = "select t.task_id, t.fields, t.priority, t.status from tasks t join tags_to_tasks ttt on t.task_id = ttt.task_id join tags tg on tg.tag_id = ttt.tag_id where t.user_id = $1 and t.status = $2 group by t.task_id order by priority"
	describeTask              = "select t.fields, t.priority, t.status from tasks t where t.task_id = $1 and t.user_id = $2"
	getTagsForTasksQuery      = "select distinct tg.tag_id, ttt.task_id, tg.name from tags tg join tags_to_tasks ttt on tg.tag_id = ttt.tag_id where ttt.task_id = any($1)"
	getAddendumsForTasksQuery = "select a.task_id, a.content, a.write_time from addendums a where a.task_id = any($1) order by a.write_time"
	getTagsFromString         = "select tg.tag_id, tg.name from tags tg where tg.name = any ($1)"
	insertTag                 = "insert into tags (user_id, tag_id, write_time, name) values ($1, nextval('tag_ids'), now(), $2) returning tag_id, name"
	insertTagsToTasks         = "insert into tags_to_tasks (task_id, tag_id) values ($1, $2)"
	insertAddundum            = "insert into addendums (addendum_id, user_id, task_id, content, write_time) values (nextval('addendum_ids'), $1, $2, $3, now())"
)

type Database struct {
	psqlUrl string
}

type TaskAttributes struct {
	Name           string        `json:"name"`
	TimeToComplete time.Duration `json:"timeToComplete"`
	CreatedTime    time.Time     `json:"createdTime"`
}

func NewDatabase(psqlUrl string) *Database {
	return &Database{
		psqlUrl: psqlUrl,
	}
}

func (e *Database) Describe(
	ctx context.Context,
	userId auth.UserId,
	taskId TaskId,
) ([]types.Pair[TaskId, Task], error) {
	var res []types.Pair[TaskId, Task]
	if err := store.Call(ctx, e.psqlUrl, func(c *pgx.Conn) error {
		var fields TaskAttributes
		var priority int
		var status int
		err := c.QueryRow(ctx, describeTask, taskId, userId).Scan(&fields, &priority, &status)
		if err != nil {
			return fmt.Errorf("getting task row for describe: %w", err)
		}

		describedTask := TaskFromDb(fields, Priority(priority), Status(status))
		if err := getTagsForTasks(ctx, c, map[TaskId]*Task{taskId: &describedTask}); err != nil {
			return fmt.Errorf("getting tags for tasks: %w", err)
		}
		if err := getAddendumsForTasks(ctx, c, map[TaskId]*Task{taskId: &describedTask}); err != nil {
			return fmt.Errorf("getting addendums for tasks: %w", err)
		}

		res = append(res, types.Of(TaskId(taskId), describedTask))
		return nil
	}); err != nil {
		return nil, fmt.Errorf("making connection to postgres for query: %w", err)
	}
	return res, nil
}

func (e *Database) Get(
	ctx context.Context,
	userId auth.UserId,
	status Status,
	tags ...Tag,
) ([]types.Pair[TaskId, Task], error) {
	var res []types.Pair[TaskId, Task]
	if err := store.Call(ctx, e.psqlUrl, func(c *pgx.Conn) error {
		var rows pgx.Rows
		var err error
		if len(tags) > 0 {
			rows, err = c.Query(ctx, getTasksWithTags, userId, tags, status)
		} else {
			rows, err = c.Query(ctx, getTasksWithIgnoringTasks, userId, status)
		}

		if err != nil {
			return fmt.Errorf("calling postgres for query: %w", err)
		}
		defer rows.Close()

		lookup := map[TaskId]*Task{}
		for rows.Next() {
			var taskId uint64
			var fields TaskAttributes
			var priority int
			var status int
			if err := rows.Scan(&taskId, &fields, &priority, &status); err != nil {
				return fmt.Errorf("scanning next tag: %w", err)
			}
			task := TaskFromDb(fields, Priority(priority), Status(status))
			res = append(res, types.Of(TaskId(taskId), task))
			lookup[TaskId(taskId)] = &task
		}

		if err := getTagsForTasks(ctx, c, lookup); err != nil {
			return fmt.Errorf("getting tags for tasks: %w", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("making connection to postgres for query: %w", err)
	}
	return res, nil
}

func (e *Database) Mark(
	ctx context.Context,
	userId auth.UserId,
	taskId TaskId,
	content string,
) error {
	if err := store.Call(ctx, e.psqlUrl, func(c *pgx.Conn) error {
		if _, err := c.Exec(ctx, insertAddundum, userId, taskId, content); err != nil {
			return fmt.Errorf("putting new addendum into db: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("writing addendum: %w", err)
	}
	return nil
}

func (e *Database) Put(
	ctx context.Context,
	userId auth.UserId,
	task Task,
) (TaskId, error) {
	newTaskId, err := store.CallAndReturn(ctx, e.psqlUrl, func(c *pgx.Conn) (*uint64, error) {
		rows, err := c.Query(ctx, getTagsFromString, task.tags)
		if err != nil {
			return nil, fmt.Errorf("getting tags from names on task: %w", err)
		}
		defer rows.Close()
		seen := map[Tag]uint64{}
		for rows.Next() {
			var tagId uint64
			var name string
			if err := rows.Scan(&tagId, &name); err != nil {
				return nil, fmt.Errorf("scanning next tag: %w", err)
			}
			seen[Tag(name)] = tagId
		}

		batch := &pgx.Batch{}
		for _, t := range task.tags {
			if _, exists := seen[t]; exists {
				continue
			}
			batch.Queue(insertTag, userId, t)
		}
		if batch.Len() > 0 {
			batchResult := c.SendBatch(ctx, batch)
			defer batchResult.Close()
			for i := 0; i < batch.Len(); i++ {
				var tagId uint64
				var name string
				err := batchResult.QueryRow().Scan(&tagId, &name)
				if err != nil {
					return nil, fmt.Errorf("executing batch tag insert: %w", err)
				}
				seen[Tag(name)] = tagId
			}
		}

		var newTaskId uint64
		if err := c.QueryRow(ctx, insertTaskQuery, userId, TaskAttributes{
			TimeToComplete: task.timeToComplete,
			Name:           task.name,
			CreatedTime:    time.Now(),
		},
			task.priority, task.status).Scan(&newTaskId); err != nil {
			return nil, fmt.Errorf("putting task into db: %w", err)
		}

		if len(task.tags) == 0 {
			return &newTaskId, nil
		}

		batch = &pgx.Batch{}
		for _, t := range task.tags {
			batch.Queue(insertTagsToTasks, newTaskId, seen[t])
		}

		batchResult := c.SendBatch(ctx, batch)
		defer batchResult.Close()
		for i := 0; i < batch.Len(); i++ {
			if _, err := batchResult.Exec(); err != nil {
				return nil, fmt.Errorf("executing batch tag insert: %w", err)
			}
		}

		return &newTaskId, nil
	})
	if err != nil {
		return 0, fmt.Errorf("calling db for insert task request: %w", err)
	}
	return TaskId(*newTaskId), nil
}

func getTagsForTasks(
	ctx context.Context,
	conn *pgx.Conn,
	tasks map[TaskId]*Task,
) error {
	taskIds := make([]uint64, len(tasks))
	for taskId := range tasks {
		taskIds = append(taskIds, uint64(taskId))
	}

	rows, err := conn.Query(ctx, getTagsForTasksQuery, taskIds)
	if err != nil {
		return fmt.Errorf("getting tags for tasks: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var taskId uint64
		var tagId uint64
		var name string
		if err := rows.Scan(&tagId, &taskId, &name); err != nil {
			return fmt.Errorf("scanning next tag: %w", err)
		}
		t := tasks[TaskId(taskId)]
		t.tags = append(t.tags, Tag(name))
	}
	return nil
}

func getAddendumsForTasks(
	ctx context.Context,
	conn *pgx.Conn,
	tasks map[TaskId]*Task,
) error {
	taskIds := make([]uint64, 0, len(tasks))
	for taskId := range tasks {
		taskIds = append(taskIds, uint64(taskId))
	}

	rows, err := conn.Query(ctx, getAddendumsForTasksQuery, taskIds)
	if err != nil {
		return fmt.Errorf("getting addendums for tasks: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var taskId uint64
		var content string
		var writeTime time.Time
		if err := rows.Scan(&taskId, &content, &writeTime); err != nil {
			return fmt.Errorf("scanning next tag: %w", err)
		}
		t := tasks[TaskId(taskId)]
		t.addendums = append(t.addendums, NewAddendum(writeTime, content))
	}
	return nil
}
