package tui

import (
	"fmt"
	"strings"

	"github.com/WadeCappa/taskmaster/internal/calls"
	"github.com/WadeCappa/taskmaster/internal/types"
	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.KeyMsg:
		return m.handleKey(message)
	case tea.WindowSizeMsg:
		m.tui.width = message.Width
		m.tui.height = message.Height
		return m, nil
	case maybeTasksLoadedEvent:
		m.tasksLoading = false
		tasks, err := message.result.Unwrap()
		if err != nil {
			m.tasksErr = err
			m.tasks = nil
			return m, nil
		}
		m.tasksErr = nil
		m.tasks = tasks
		m.taskCursor = 0
		m.taskListOffset = 0
		m.detail = nil
		m.detailErr = nil
		if len(m.tasks) > 0 {
			m.detailLoading = true
			m.detailTaskId = m.tasks[0].id
			return m, m.fetchDetailCmd(m.tasks[0].id)
		}
		return m, nil
	case maybeTaskDetailLoadedEvent:
		event, err := message.result.Unwrap()
		if err != nil {
			m.detailErr = err
			m.detail = nil
			return m, nil
		}
		// We can race here and not see the results of a details call until a user has
		// already selected the next task. In this case, we should just throw away the
		// result of this call.
		if event.taskId != m.detailTaskId {
			return m, nil
		}
		m.detailLoading = false
		m.detailErr = nil
		m.detail = event.detail
		return m, nil
	case maybeAddendumCreatedEvent:
		taskId, err := message.result.Unwrap()
		if err != nil {
			m.detailErr = err
			return m, nil
		}
		m.mode = modeNormal
		m.detailLoading = true
		m.detailTaskId = taskId
		return m, m.fetchDetailCmd(taskId)
	case maybeStatusSetEvent:
		taskId, err := message.result.Unwrap()
		if err != nil {
			m.detailErr = err
			return m, nil
		}
		m.mode = modeNormal
		m.detailLoading = true
		m.detailTaskId = taskId
		return m, tea.Batch(m.refetch(), m.fetchDetailCmd(taskId))
	}
	return m, nil
}

func (m Model) handleKey(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case modeNormal:
		return m.handleNormalKey(message)
	case modeDetailFocused:
		return m.handleDetailFocusedKey(message)
	case modeAddendumInput:
		return m.handleAddendumInputKey(message)
	case modeStatusSelect:
		return m.handleStatusSelectKey(message)
	case modeTagEdit:
		return m.handleTagEditKey(message)
	}
	return m, nil
}

func (m Model) handleNormalKey(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch message.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "j", "down":
		if len(m.tasks) > 0 {
			m.taskCursor = (m.taskCursor + 1) % len(m.tasks)
			m.adjustOffset()
			m.detailLoading = true
			m.detailTaskId = m.tasks[m.taskCursor].id
			return m, m.fetchDetailCmd(m.tasks[m.taskCursor].id)
		}
	case "k", "up":
		if len(m.tasks) > 0 {
			m.taskCursor--
			if m.taskCursor < 0 {
				m.taskCursor = len(m.tasks) - 1
			}
			m.adjustOffset()
			m.detailLoading = true
			m.detailTaskId = m.tasks[m.taskCursor].id
			return m, m.fetchDetailCmd(m.tasks[m.taskCursor].id)
		}
	case "h", "left":
		m.activeStatus--
		if m.activeStatus < 0 {
			m.activeStatus = len(taskspb.Status_value) - 1
		}
		return m, m.refetch()
	case "l", "right":
		m.activeStatus = (m.activeStatus + 1) % len(taskspb.Status_value)
		return m, m.refetch()
	case "t":
		m.mode = modeTagEdit
		m.savedTags = m.tags
		m.tagInput = strings.Join(m.tags, ", ")
		return m, nil
	case "i":
		m.mode = modeAddendumInput
		m.addendumTextarea.Reset()
		return m, m.addendumTextarea.Focus()
	case "x":
		if m.detail != nil {
			m.statusCursor = int(m.detail.status)
		}
		m.mode = modeStatusSelect
		return m, nil
	case "tab":
		m.mode = modeDetailFocused
		m.detailOffset = 0
		return m, nil
	}
	return m, nil
}

func (m Model) handleDetailFocusedKey(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch message.String() {
	case "j", "down":
		m.detailOffset++
		return m, nil
	case "k", "up":
		if m.detailOffset > 0 {
			m.detailOffset--
		}
		return m, nil
	case "esc", "tab":
		m.mode = modeNormal
		return m, nil
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleAddendumInputKey(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch message.String() {
	case "esc":
		m.mode = modeNormal
		m.addendumTextarea.Blur()
		m.addendumTextarea.Reset()
		return m, nil
	case "ctrl+s":
		content := m.addendumTextarea.Value()
		if content != "" && len(m.tasks) > 0 {
			taskId := m.tasks[m.taskCursor].id
			m.addendumTextarea.Blur()
			m.addendumTextarea.Reset()
			return m, m.createAddendumCmd(taskId, content)
		}
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	default:
		var cmd tea.Cmd
		m.addendumTextarea, cmd = m.addendumTextarea.Update(message)
		return m, cmd
	}
}

func (m Model) handleStatusSelectKey(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch message.String() {
	case "j", "down":
		m.statusCursor = (m.statusCursor + 1) % len(statuses)
		return m, nil
	case "k", "up":
		m.statusCursor = (m.statusCursor - 1 + len(statuses)) % len(statuses)
		return m, nil
	case "enter":
		if len(m.tasks) > 0 {
			taskId := m.tasks[m.taskCursor].id
			status := statuses[m.statusCursor]
			return m, m.setStatusCmd(taskId, status)
		}
		m.mode = modeNormal
		return m, nil
	case "esc":
		m.mode = modeNormal
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleTagEditKey(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch message.String() {
	case "enter":
		m.mode = modeNormal
		input := strings.TrimSpace(m.tagInput)
		if input == "" {
			m.tags = nil
		} else {
			parts := strings.Split(input, ",")
			m.tags = nil
			for _, p := range parts {
				t := strings.TrimSpace(p)
				if t != "" {
					m.tags = append(m.tags, t)
				}
			}
		}
		return m, m.refetch()
	case "esc":
		m.mode = modeNormal
		m.tags = m.savedTags
		return m, nil
	case "backspace":
		if len(m.tagInput) > 0 {
			m.tagInput = m.tagInput[:len(m.tagInput)-1]
		}
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	default:
		if len(message.String()) == 1 {
			m.tagInput += message.String()
		}
		return m, nil
	}
}

// TODO: Figure out what this code does
func (m *Model) adjustOffset() {
	listHeight := m.listHeight()
	if m.taskCursor < m.taskListOffset {
		m.taskListOffset = m.taskCursor
	}
	if m.taskCursor >= m.taskListOffset+listHeight {
		m.taskListOffset = m.taskCursor - listHeight + 1
	}
}

func (m *Model) refetch() tea.Cmd {
	m.tasksLoading = true
	return m.fetchTasksCmd()
}

func (m Model) listHeight() int {
	// height minus top bar (3) minus help bar (2) minus panel borders (2)
	return max(m.tui.height-7, 1)
}

func (m Model) fetchTasksCmd() tea.Cmd {
	return func() tea.Msg {
		responses, err := calls.Get(m.ctx, m.client, m.tags, taskspb.Status(m.activeStatus))
		if err != nil {
			return maybeTasksLoadedEvent{
				result: types.Failure[[]taskEntry](fmt.Errorf("getting tasks from server: %w", err)),
			}
		}

		tasks := make([]taskEntry, len(responses))
		for index, resp := range responses {
			tasks[index] = taskFromWire(resp)
		}
		return maybeTasksLoadedEvent{types.Success(tasks)}
	}
}

func (m Model) fetchDetailCmd(taskId uint64) tea.Cmd {
	return func() tea.Msg {
		resp, err := m.client.DescribeTask(m.ctx, &taskspb.DescribeTaskRequest{
			TaskId: taskId,
		})
		if err != nil {
			return maybeTaskDetailLoadedEvent{
				result: types.Failure[taskDetailLoadedEvent](fmt.Errorf("getting task details from server: %w", err)),
			}
		}
		detail := detailFromWire(resp)
		return maybeTaskDetailLoadedEvent{
			types.Success(taskDetailLoadedEvent{
				taskId: taskId,
				detail: &detail,
			}),
		}
	}
}

func (m Model) createAddendumCmd(taskId uint64, content string) tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.MarkTask(m.ctx, &taskspb.MarkTaskRequest{
			TaskId:  taskId,
			Content: content,
		})
		if err != nil {
			return maybeAddendumCreatedEvent{types.Failure[uint64](fmt.Errorf("creating addendum: %w", err))}
		}
		return maybeAddendumCreatedEvent{types.Success(taskId)}
	}
}

func (m Model) setStatusCmd(taskId uint64, status taskspb.Status) tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.SetStatus(m.ctx, &taskspb.SetStatusRequest{
			TaskId: taskId,
			Status: status,
		})
		if err != nil {
			return maybeStatusSetEvent{types.Failure[uint64](fmt.Errorf("setting status: %w", err))}
		}
		return maybeStatusSetEvent{types.Success(taskId)}
	}
}
