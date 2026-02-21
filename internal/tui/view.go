package tui

import (
	"fmt"
	"strings"

	taskspb "github.com/WadeCappa/taskmaster/pkg/go/tasks/v1"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.tui.width == 0 || m.tui.height == 0 {
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
	if m.mode == modeTagEdit {
		tagSection = "Tags: " + tagInputStyle.Render(m.tagInput+"_")
	} else if len(m.tags) > 0 {
		tagSection = "Tags: " + strings.Join(m.tags, ", ")
	} else {
		tagSection = "Tags: " + dimStyle.Render("<none>")
	}

	bar := statusSection + "  | " + tagSection

	style := lipgloss.NewStyle().
		Width(m.tui.width).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color("241"))
	return style.Render(bar)
}

func (m Model) viewPanels() string {
	leftWidth := m.tui.width/2 - 2
	rightWidth := m.tui.width - leftWidth - 4
	panelHeight := m.listHeight()

	left := m.viewTaskList(leftWidth, panelHeight)

	var right string
	var rightBorder lipgloss.Style

	switch m.mode {
	case modeAddendumInput:
		prompt := detailLabel.Render("Addendum: ") + tagInputStyle.Render(m.addendumInput+"_")
		detail := m.viewDetail(rightWidth, panelHeight-1)
		right = detail + "\n" + prompt
		rightBorder = focusedBorderStyle
	case modeDetailFocused:
		right = m.viewDetail(rightWidth, panelHeight)
		rightBorder = focusedBorderStyle
	case modeStatusSelect:
		right = m.viewStatusSelect(rightWidth, panelHeight)
		rightBorder = focusedBorderStyle
	default:
		right = m.viewDetail(rightWidth, panelHeight)
		rightBorder = borderStyle
	}

	leftPanel := borderStyle.Width(leftWidth).Height(panelHeight).Render(left)
	rightPanel := rightBorder.Width(rightWidth).Height(panelHeight).Render(right)

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
	var allLines []string
	allLines = append(allLines, detailLabel.Render("Name: ")+d.name)
	allLines = append(allLines, detailLabel.Render("Priority: ")+taskspb.Priority_name[int32(d.priority)])
	allLines = append(allLines, detailLabel.Render("Status: ")+taskspb.Status_name[int32(d.status)])
	allLines = append(allLines, detailLabel.Render("Time: ")+formatDuration(d.minutes))

	if len(d.tags) > 0 {
		allLines = append(allLines, detailLabel.Render("Tags: ")+strings.Join(d.tags, ", "))
	} else {
		allLines = append(allLines, detailLabel.Render("Tags: ")+dimStyle.Render("none"))
	}

	allLines = append(allLines, "")
	if len(d.addendums) > 0 {
		allLines = append(allLines, detailLabel.Render(fmt.Sprintf("Addendums (%d):", len(d.addendums))))
		for _, a := range d.addendums {
			dateStr := a.time.Format("2006-01-02")
			prefix := "  " + dateStr + ": "
			indent := strings.Repeat(" ", len(prefix))
			contentWidth := width - len(prefix)
			if contentWidth < 10 {
				contentWidth = 10
			}
			wrapped := wordWrapLines(a.content, contentWidth)
			for i, line := range wrapped {
				if i == 0 {
					allLines = append(allLines, prefix+line)
				} else {
					allLines = append(allLines, indent+line)
				}
			}
		}
	} else {
		allLines = append(allLines, dimStyle.Render("No addendums"))
	}

	visible := allLines
	if m.detailOffset < len(allLines) {
		visible = allLines[m.detailOffset:]
	} else {
		visible = nil
	}
	if len(visible) > height {
		visible = visible[:height]
	}
	return strings.Join(visible, "\n")
}

func (m Model) viewStatusSelect(width, height int) string {
	lines := []string{detailLabel.Render("Select Status:"), ""}
	for i, s := range statuses {
		prefix := "  "
		name := taskspb.Status_name[int32(s)]
		if i == m.statusCursor {
			prefix = "> "
			name = selectedStyle.Render(name)
		}
		lines = append(lines, prefix+name)
	}
	return strings.Join(lines, "\n")
}

func (m Model) viewHelpBar() string {
	var help string
	switch m.mode {
	case modeNormal:
		help = "i: add addendum  x: set status  tab: focus detail  j/k: navigate  h/l: status  t: tags  q: quit"
	case modeDetailFocused:
		help = "j/k: scroll  tab/esc: back  q: quit"
	case modeAddendumInput:
		help = "enter: submit  esc: cancel  ctrl+c: quit"
	case modeStatusSelect:
		help = "j/k: move  enter: confirm  esc: cancel"
	case modeTagEdit:
		help = "enter: apply tags  esc: cancel  ctrl+c: quit"
	}
	return helpStyle.Padding(0, 1).Render(help)
}
