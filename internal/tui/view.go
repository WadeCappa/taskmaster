package tui

import (
	"fmt"
	"strings"

	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	topBar := m.viewTopBar()
	panels := m.viewPanels()
	helpBar := m.viewHelpBar()

	return lipgloss.JoinVertical(lipgloss.Left, topBar, panels, helpBar)
}

func (m Model) viewTopBar() string {
	var statusParts []string
	for i, status := range statuses {
		name := taskspb.Status_name[int32(i)]
		if status == taskspb.Status(m.activeStatus) {
			statusParts = append(statusParts, statusHighlight.Render("["+name+"]"))
		} else {
			statusParts = append(statusParts, dimStyle.Render(" "+name+" "))
		}
	}
	statusSection := "Status: " + strings.Join(statusParts, " ")

	var tagSection string
	if m.editingTags {
		tagSection = "Tags: " + tagInputStyle.Render(m.tagInput+"_")
	} else if len(m.tags) > 0 {
		tagSection = "Tags: " + strings.Join(m.tags, ", ")
	} else {
		tagSection = "Tags: " + dimStyle.Render("<none>")
	}

	bar := statusSection + "  | " + tagSection

	style := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color("241"))
	return style.Render(bar)
}

func (m Model) viewPanels() string {
	leftWidth := m.width/2 - 2
	rightWidth := m.width - leftWidth - 4
	panelHeight := m.listHeight()

	left := m.viewTaskList(leftWidth, panelHeight)
	right := m.viewDetail(rightWidth, panelHeight)

	leftPanel := borderStyle.Width(leftWidth).Height(panelHeight).Render(left)
	rightPanel := borderStyle.Width(rightWidth).Height(panelHeight).Render(right)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
}

func (m Model) viewTaskList(width, height int) string {
	if m.tasksLoading {
		return "Loading tasks..."
	}
	if m.tasksErr != nil {
		return errorStyle.Render("Error: " + m.tasksErr.Error())
	}
	if len(m.tasks) == 0 {
		return dimStyle.Render("No tasks found")
	}

	var lines []string
	end := min(m.taskListOffset+height, len(m.tasks))

	for i := m.taskListOffset; i < end; i++ {
		t := m.tasks[i]
		prefix := "  "
		if i == m.taskCursor {
			prefix = "> "
		}

		name := t.name
		prio := priorityLabel(t.priority)
		timeStr := formatDuration(t.minutes)

		// Truncate name if too long
		maxName := width - len(prefix) - len(prio) - len(timeStr) - 4
		if maxName < 5 {
			maxName = 5
		}
		if len(name) > maxName {
			name = name[:maxName-1] + "…"
		}

		gap := width - len(prefix) - len(name) - len(prio) - len(timeStr) - 2
		if gap < 1 {
			gap = 1
		}

		line := prefix + name + strings.Repeat(" ", gap) + timeStr + " " + prio
		if i == m.taskCursor {
			line = selectedStyle.Render(line)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (m Model) viewDetail(width, height int) string {
	if len(m.tasks) == 0 {
		return dimStyle.Render("Select a task")
	}
	if m.detailLoading {
		return "Loading..."
	}
	if m.detailErr != nil {
		return errorStyle.Render("Error: " + m.detailErr.Error())
	}
	if m.detail == nil {
		return ""
	}

	d := m.detail
	var lines []string
	lines = append(lines, detailLabel.Render("Name: ")+d.name)
	lines = append(lines, detailLabel.Render("Priority: ")+priorityName(d.priority))
	lines = append(lines, detailLabel.Render("Status: ")+taskspb.Status_name[int32(d.status)])
	lines = append(lines, detailLabel.Render("Time: ")+formatDuration(d.minutes))

	if len(d.tags) > 0 {
		lines = append(lines, detailLabel.Render("Tags: ")+strings.Join(d.tags, ", "))
	} else {
		lines = append(lines, detailLabel.Render("Tags: ")+dimStyle.Render("none"))
	}

	lines = append(lines, "")
	if len(d.addendums) > 0 {
		lines = append(lines, detailLabel.Render(fmt.Sprintf("Addendums (%d):", len(d.addendums))))
		for _, a := range d.addendums {
			dateStr := a.time.Format("2006-01-02")
			content := a.content
			maxContent := width - len(dateStr) - 6
			if maxContent > 0 && len(content) > maxContent {
				content = content[:maxContent-3] + "..."
			}
			lines = append(lines, "  "+dateStr+": "+content)
		}
	} else {
		lines = append(lines, dimStyle.Render("No addendums"))
	}

	return strings.Join(lines, "\n")
}

func (m Model) viewHelpBar() string {
	var help string
	if m.editingTags {
		help = "enter: apply tags  esc: cancel  ctrl+c: quit"
	} else {
		help = "j/k: navigate  J/L: status  t: edit tags  q: quit"
	}
	return helpStyle.Padding(0, 1).Render(help)
}
