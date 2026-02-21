package tui

import (
	"context"

	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	client taskspb.TasksClient
	ctx    context.Context

	activeStatus int
	tags         []string

	editingTags bool
	tagInput    string
	savedTags   []string

	tasks          []taskEntry
	taskCursor     int
	taskListOffset int

	detail        *taskDetail
	detailLoading bool
	detailTaskId  uint64

	tasksLoading bool
	tasksErr     error
	detailErr    error

	tui tuiState
}

type tuiState struct {
	width  int
	height int
}

func NewModel(client taskspb.TasksClient, ctx context.Context) Model {
	return Model{
		client: client,
		ctx:    ctx,
	}
}

func (m Model) Init() tea.Cmd {
	return m.fetchTasksCmd()
}
