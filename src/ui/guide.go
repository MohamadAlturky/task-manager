package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type GuideModel struct {
	width  int
	styles StyleConfig
	page   int
}

func NewGuideModel(width int) GuideModel {
	return GuideModel{
		width:  width,
		styles: NewStyleConfig(width),
		page:   0,
	}
}

func (m GuideModel) Update(msg tea.Msg) (GuideModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		default:
			m.page++
			if m.page >= 3 {
				return m, nil
			}
		}
	}
	return m, nil
}

func (m GuideModel) View() string {
	var s strings.Builder

	// Title
	title := m.styles.TitleStyle.Render("Welcome to tasker")
	s.WriteString(title)
	s.WriteString("\n\n")

	// Description
	desc := lipgloss.NewStyle().
		Width(m.width - 4).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#7D7D7D")).
		Render("A beautiful and efficient task management TUI application")
	s.WriteString(desc)
	s.WriteString("\n\n")

	// Page content
	switch m.page {
	case 0:
		// Basic Navigation
		sections := []struct {
			title   string
			content string
		}{
			{
				"Navigation",
				"↑/↓ or j/k  Navigate tasks\n" +
					"< or ,      Previous page\n" +
					"> or .      Next page",
			},
			{
				"Task Management",
				"n           New task\n" +
					"v           View task details\n" +
					"d           Delete task\n" +
					"Space       Toggle completion",
			},
		}
		s.WriteString(m.renderSections(sections))

	case 1:
		// Status & Filtering
		sections := []struct {
			title   string
			content string
		}{
			{
				"Status & Filtering",
				"s           Enter status filter\n" +
					"1-3         Set/Filter status\n" +
					"0           Clear status filter\n" +
					"/           Search tasks",
			},
			{
				"Task Details",
				"1-3         Change task status\n" +
					"Esc         Return to list",
			},
		}
		s.WriteString(m.renderSections(sections))

	case 2:
		// Advanced Features
		sections := []struct {
			title   string
			content string
		}{
			{
				"Status Indicators",
				"📝          Todo\n" +
					"🔄          In Progress\n" +
					"✅          Done",
			},
			{
				"Application",
				"q or Ctrl+C Quit application\n" +
					"?           Show this guide again",
			},
		}
		s.WriteString(m.renderSections(sections))
	}

	// Progress indicator
	progress := make([]string, 3)
	for i := range progress {
		if i == m.page {
			progress[i] = "●"
		} else {
			progress[i] = "○"
		}
	}
	progressBar := lipgloss.NewStyle().
		Width(m.width - 4).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#7D7D7D")).
		Render(strings.Join(progress, " "))
	s.WriteString("\n" + progressBar)

	// Footer
	footer := lipgloss.NewStyle().
		Width(m.width - 4).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#7D7D7D")).
		PaddingTop(1).
		Render("Press any key to continue...")
	s.WriteString("\n" + footer)

	return m.styles.AppStyle.Render(s.String())
}

func (m GuideModel) renderSections(sections []struct {
	title   string
	content string
}) string {
	// Calculate column width
	colWidth := (m.width - 8) / 2 // Account for padding and gap
	if colWidth < 30 {
		colWidth = 30
	}

	// Create two columns
	var leftCol, rightCol strings.Builder
	for i, section := range sections {
		sectionStyle := lipgloss.NewStyle().
			Width(colWidth).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D7D7D"))

		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFA500")).
			PaddingBottom(1)

		content := sectionStyle.Render(
			titleStyle.Render(section.title) + "\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#D9DCCF")).Render(section.content),
		)

		if i%2 == 0 {
			leftCol.WriteString(content + "\n\n")
		} else {
			rightCol.WriteString(content + "\n\n")
		}
	}

	// Combine columns
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftCol.String(),
		rightCol.String(),
	)
}
