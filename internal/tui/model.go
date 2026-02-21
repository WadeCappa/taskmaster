package tui

import (
	"context"

	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type uiMode int

const (
	modeNormal        uiMode = iota // j/k navigate tasks; h/l switch status
	modeDetailFocused               // j/k scroll detail panel; Esc → normal
	modeAddendumInput               // composing multi-line addendum; ctrl+s submit, Esc cancel
	modeStatusSelect                // picking new status from menu; Enter confirm, Esc cancel
	modeTagEdit                     // existing tag editing
)

type Model struct {
	client taskspb.TasksClient
	ctx    context.Context

	activeStatus int
	tags         []string

	mode      uiMode
	tagInput  string
	savedTags []string

	addendumTextarea textarea.Model
	statusCursor     int
	detailOffset     int

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
	ta := textarea.New()
	ta.Placeholder = "Write addendum..."
	ta.CharLimit = 0
	return Model{
		client:           client,
		ctx:              ctx,
		addendumTextarea: ta,
	}
}

func (m Model) Init() tea.Cmd {
	return m.fetchTasksCmd()
}
