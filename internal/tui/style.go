package tui

import "github.com/charmbracelet/lipgloss"

var (
	selectedStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	dimStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	borderStyle     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("241"))
	helpStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	tagInputStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	statusHighlight = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	detailLabel     = lipgloss.NewStyle().Bold(true)
)
