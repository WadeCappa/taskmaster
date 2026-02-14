package tui

import (
	"fmt"
	"strings"

	"github.com/WadeCappa/taskmaster/internal/calls"
	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tasksLoadedMsg:
		m.tasksLoading = false
		if msg.err != nil {
			m.tasksErr = msg.err
			m.tasks = nil
			return m, nil
		}
		m.tasksErr = nil
		m.tasks = msg.tasks
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
	case taskDetailLoadedMsg:
		if msg.taskId != m.detailTaskId {
			return m, nil // stale response
		}
		m.detailLoading = false
		if msg.err != nil {
			m.detailErr = msg.err
			m.detail = nil
			return m, nil
		}
		m.detailErr = nil
		m.detail = msg.detail
		return m, nil
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editingTags {
		return m.handleTagEditKey(msg)
	}

	switch msg.String() {
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
		m.editingTags = true
		m.savedTags = m.tags
		m.tagInput = strings.Join(m.tags, ", ")
		return m, nil
	}
	return m, nil
}

func (m Model) handleTagEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.editingTags = false
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
		m.editingTags = false
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
		if len(msg.String()) == 1 {
			m.tagInput += msg.String()
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
	return max(m.height-7, 1)
}

func (m Model) fetchTasksCmd() tea.Cmd {
	return func() tea.Msg {
		responses, err := calls.Get(m.ctx, m.client, m.tags, taskspb.Status(m.activeStatus))
		if err != nil {
			return tasksLoadedMsg{err: fmt.Errorf("getting tasks from server: %w", err)}
		}

		tasks := make([]taskEntry, len(responses))
		for index, resp := range responses {
			tasks[index] = taskFromWire(resp)
		}
		return tasksLoadedMsg{tasks: tasks}
	}
}

func (m Model) fetchDetailCmd(taskId uint64) tea.Cmd {
	return func() tea.Msg {
		resp, err := m.client.DescribeTask(m.ctx, &taskspb.DescribeTaskRequest{
			TaskId: taskId,
		})
		if err != nil {
			return taskDetailLoadedMsg{taskId: taskId, err: fmt.Errorf("calling DescribeTask: %w", err)}
		}

		detail := detailFromWire(resp)
		return taskDetailLoadedMsg{taskId: taskId, detail: &detail}
	}
}
