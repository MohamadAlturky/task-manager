package ui

import (
	"fmt"
	"strings"
	"time"

	"devtui/src/db"
	"devtui/src/models"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type inputMode int

const (
	normalMode inputMode = iota
	taskInputMode
	searchMode
	statusMode
	detailsMode
	calendarMode
	editMode
)

type editField int

const (
	titleField editField = iota
	descriptionField
	priorityField
	statusField
)

type Model struct {
	taskList      *models.TaskList
	input         textinput.Model
	mode          inputMode
	err           error
	width         int
	filterQuery   string
	statusFilter  models.Status
	db            *db.Database
	styles        StyleConfig
	selectedTask  *models.Task
	showGuide     bool
	guideModel    GuideModel
	selectedDate  *time.Time
	calendarMonth int
	calendarYear  int
	calendarDay   int
	editInputs    [4]textinput.Model
	editingField  editField
}

func NewModel() (Model, error) {
	ti := textinput.New()
	ti.Placeholder = "New task..."
	ti.CharLimit = 156
	ti.Width = 50

	database, err := db.NewDatabase()
	if err != nil {
		return Model{}, fmt.Errorf("failed to initialize database: %v", err)
	}

	tasks, err := database.LoadTasks()
	if err != nil {
		return Model{}, fmt.Errorf("failed to load tasks: %v", err)
	}

	taskList := models.NewTaskList()
	taskList.Tasks = tasks
	taskList.UpdatePagination()

	m := Model{
		taskList:  taskList,
		input:     ti,
		mode:      normalMode,
		db:        database,
		width:     80,
		showGuide: true,
	}
	m.styles = NewStyleConfig(m.width)
	m.guideModel = NewGuideModel(m.width)
	now := time.Now()
	m.calendarMonth = int(now.Month())
	m.calendarYear = now.Year()
	m.calendarDay = now.Day()
	m.editingField = titleField
	for i := range m.editInputs {
		m.editInputs[i] = textinput.New()
		m.editInputs[i].CharLimit = 156
		m.editInputs[i].Width = 50
	}
	return m, nil
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.styles = NewStyleConfig(m.width)
		m.input.Width = m.styles.InputStyle.GetWidth()
		m.guideModel = NewGuideModel(m.width)

	case tea.KeyMsg:
		if m.showGuide {
			var guideCmd tea.Cmd
			m.guideModel, guideCmd = m.guideModel.Update(msg)
			if m.guideModel.page >= 3 {
				m.showGuide = false
				m.mode = calendarMode
				return m, nil
			}
			return m, guideCmd
		}
		if m.mode == calendarMode {
			var handled bool
			m.selectedDate, handled = handleCalendarInput(m.selectedDate, msg.String(), &m)
			if handled {
				m.mode = normalMode
				return m, nil
			}
			if msg.String() == "q" || msg.String() == "esc" {
				m.selectedDate = nil
				m.mode = normalMode
				return m, nil
			}
			return m, nil
		}
		switch m.mode {
		case normalMode:
			switch msg.String() {
			case "q", "ctrl+c":
				m.db.Close()
				return m, tea.Quit
			case "j", "down":
				currentTasks := m.taskList.GetCurrentPageTasks()
				if m.taskList.Selected < len(currentTasks)-1 {
					m.taskList.Selected++
				}
			case "k", "up":
				if m.taskList.Selected > 0 {
					m.taskList.Selected--
				}
			case "n":
				m.input.Placeholder = "New task..."
				m.input.Focus()
				m.input.Reset()
				m.mode = taskInputMode
				return m, nil
			case " ":
				currentTasks := m.taskList.GetCurrentPageTasks()
				if len(currentTasks) > 0 {
					globalIndex := (m.taskList.CurrentPage-1)*m.taskList.PageSize + m.taskList.Selected
					m.taskList.ToggleTask(globalIndex)
					if err := m.db.SaveTask(m.taskList.Tasks[globalIndex]); err != nil {
						m.err = fmt.Errorf("failed to save task: %v", err)
					}
				}
			case "d":
				currentTasks := m.taskList.GetCurrentPageTasks()
				if len(currentTasks) > 0 {
					globalIndex := (m.taskList.CurrentPage-1)*m.taskList.PageSize + m.taskList.Selected
					taskID := m.taskList.Tasks[globalIndex].ID
					if err := m.db.DeleteTask(taskID); err != nil {
						m.err = fmt.Errorf("failed to delete task: %v", err)
					} else {
						m.taskList.RemoveTask(globalIndex)
						m.taskList.UpdatePagination()
					}
				}
			case "v":
				currentTasks := m.taskList.GetCurrentPageTasks()
				if len(currentTasks) > 0 {
					globalIndex := (m.taskList.CurrentPage-1)*m.taskList.PageSize + m.taskList.Selected
					m.selectedTask = m.taskList.Tasks[globalIndex]
					m.mode = detailsMode
				}
			case "/":
				m.input.Placeholder = "Search..."
				m.input.Focus()
				m.mode = searchMode
				return m, nil
			case "s":
				m.mode = statusMode
				return m, nil
			case ">", ".":
				m.taskList.NextPage()
			case "<", ",":
				m.taskList.PrevPage()
			}
		case taskInputMode:
			switch msg.Type {
			case tea.KeyEnter:
				if m.input.Value() != "" {
					task := models.NewTask(m.input.Value(), "", models.Medium, []string{})
					if err := m.db.SaveTask(task); err != nil {
						m.err = fmt.Errorf("failed to save task: %v", err)
					} else {
						m.taskList.AddTask(task)
						m.taskList.UpdatePagination()
					}
					m.input.Reset()
					m.mode = normalMode
				}
			case tea.KeyEsc:
				m.input.Reset()
				m.mode = normalMode
			}
		case searchMode:
			switch msg.Type {
			case tea.KeyEnter, tea.KeyEsc:
				m.filterQuery = m.input.Value()
				m.input.Reset()
				m.mode = normalMode
			}
		case statusMode:
			switch msg.String() {
			case "1":
				m.statusFilter = models.Todo
				m.mode = normalMode
			case "2":
				m.statusFilter = models.InProgress
				m.mode = normalMode
			case "3":
				m.statusFilter = models.Done
				m.mode = normalMode
			case "0":
				m.statusFilter = -1
				m.mode = normalMode
			case "esc":
				m.mode = normalMode
			}
		case detailsMode:
			switch msg.String() {
			case "1":
				m.selectedTask.Status = models.Todo
				if err := m.db.SaveTask(m.selectedTask); err != nil {
					m.err = fmt.Errorf("failed to update task status: %v", err)
				}
				m.mode = normalMode
			case "2":
				m.selectedTask.Status = models.InProgress
				if err := m.db.SaveTask(m.selectedTask); err != nil {
					m.err = fmt.Errorf("failed to update task status: %v", err)
				}
				m.mode = normalMode
			case "3":
				m.selectedTask.Status = models.Done
				if err := m.db.SaveTask(m.selectedTask); err != nil {
					m.err = fmt.Errorf("failed to update task status: %v", err)
				}
				m.mode = normalMode
			case "e":
				if m.selectedTask != nil {
					m.editInputs[0].SetValue(m.selectedTask.Title)
					m.editInputs[1].SetValue(m.selectedTask.Description)
					m.editInputs[2].SetValue(priorityToString(m.selectedTask.Priority))
					m.editInputs[3].SetValue(statusToString(m.selectedTask.Status))
					m.editingField = titleField
					m.editInputs[0].Focus()
					m.mode = editMode
				}
			case "esc":
				m.mode = normalMode
			}
		case editMode:
			if keyMsg, ok := any(msg).(tea.KeyMsg); ok {
				switch keyMsg.String() {
				case "tab":
					m.editInputs[m.editingField].Blur()
					m.editingField = (m.editingField + 1) % 4
					m.editInputs[m.editingField].Focus()
				case "shift+tab":
					m.editInputs[m.editingField].Blur()
					m.editingField = (m.editingField + 3) % 4
					m.editInputs[m.editingField].Focus()
				case "enter":
					title := m.editInputs[0].Value()
					desc := m.editInputs[1].Value()
					priority := stringToPriority(m.editInputs[2].Value())
					status := stringToStatus(m.editInputs[3].Value())
					m.selectedTask.Title = title
					m.selectedTask.Description = desc
					m.selectedTask.Priority = priority
					m.selectedTask.Status = status
					if err := m.db.SaveTask(m.selectedTask); err != nil {
						m.err = fmt.Errorf("failed to save task: %v", err)
					}
					m.mode = detailsMode
				case "esc":
					m.mode = detailsMode
				}
			} else {
				m.editInputs[m.editingField], cmd = m.editInputs[m.editingField].Update(msg)
			}
		}
	}

	if m.mode == taskInputMode || m.mode == searchMode {
		m.input, cmd = m.input.Update(msg)
	}

	return m, cmd
}

func (m Model) View() string {
	if m.showGuide {
		return m.guideModel.View()
	}
	if m.mode == calendarMode {
		return m.calendarView()
	}
	if m.mode == detailsMode && m.selectedTask != nil {
		return m.detailsView()
	}
	if m.mode == editMode && m.selectedTask != nil {
		return m.editView()
	}
	var s strings.Builder

	s.WriteString(m.styles.TitleStyle.Render("DevTUI - Task Manager"))
	s.WriteString("\n\n")

	if len(m.taskList.Tasks) == 0 {
		s.WriteString(m.styles.ListStyle.Render("No tasks yet. Press 'n' to add a new task."))
	} else {
		var taskView strings.Builder
		currentTasks := m.taskList.GetCurrentPageTasks()
		// Responsive column widths
		totalWidth := m.width - 6
		colStatus := 4
		colPriority := 10
		colDate := 12
		colTitle := totalWidth - (colStatus + colPriority + colDate + 6)
		if colTitle < 10 {
			colTitle = 10
		}

		for i, task := range currentTasks {
			if m.filterQuery != "" && !strings.Contains(strings.ToLower(task.Title), strings.ToLower(m.filterQuery)) {
				continue
			}
			if m.statusFilter != -1 && task.Status != m.statusFilter {
				continue
			}
			if m.selectedDate != nil && task.DueDate != nil && !sameDay(*m.selectedDate, *task.DueDate) {
				continue
			}

			statusStr := " " // emoji
			switch task.Status {
			case models.Todo:
				statusStr = "📝"
			case models.InProgress:
				statusStr = "🔄"
			case models.Done:
				statusStr = "✅"
			}

			priorityStr := "Low"
			switch task.Priority {
			case models.Medium:
				priorityStr = "Medium"
			case models.High:
				priorityStr = "High"
			}

			dueDateStr := "-"
			if task.DueDate != nil {
				dueDateStr = task.DueDate.Format("2006-01-02")
			}

			row := fmt.Sprintf("%-*s %-*s %-*s %-*s", colStatus, statusStr, colTitle, task.Title, colPriority, priorityStr, colDate, dueDateStr)
			itemStyle := m.styles.ListItemStyle
			if i == m.taskList.Selected {
				itemStyle = itemStyle.Underline(true).Foreground(lipgloss.Color("#7D56F4"))
			}
			taskView.WriteString(itemStyle.Render(row) + "\n")
		}
		s.WriteString(m.styles.ListStyle.Render(taskView.String()))
	}

	statusBar := fmt.Sprintf(" %d tasks | Page %d/%d ", len(m.taskList.Tasks), m.taskList.CurrentPage, m.taskList.TotalPages)
	if m.filterQuery != "" {
		statusBar += fmt.Sprintf("| Filter: %s ", m.filterQuery)
	}
	if m.statusFilter != -1 {
		statusStr := "All"
		switch m.statusFilter {
		case models.Todo:
			statusStr = "Todo"
		case models.InProgress:
			statusStr = "In Progress"
		case models.Done:
			statusStr = "Done"
		}
		statusBar += fmt.Sprintf("| Status: %s ", statusStr)
	}
	s.WriteString("\n" + m.styles.StatusBarStyle.Render(statusBar))

	if m.mode == taskInputMode || m.mode == searchMode {
		s.WriteString("\n\n" + m.styles.InputStyle.Render(m.input.View()))
	} else if m.mode == statusMode {
		s.WriteString("\n\n" + m.styles.InputStyle.Render("Select status: [1] Todo [2] In Progress [3] Done [0] Clear [Esc] Cancel"))
	} else if m.mode == detailsMode && m.selectedTask != nil {
		return m.detailsView()
	} else if m.mode == editMode && m.selectedTask != nil {
		s.WriteString("\n\n" + m.styles.InputStyle.Render("Editing task..."))
	}

	if m.err != nil {
		s.WriteString("\n" + m.styles.ErrorStyle.Render(m.err.Error()))
	}

	return m.styles.AppStyle.Render(s.String())
}

func (m Model) selectionView() string {
	var s strings.Builder
	s.WriteString(m.styles.TitleStyle.Render("Task View Selection"))
	s.WriteString("\n\n")
	s.WriteString("1. Show all tasks\n")
	s.WriteString("2. Show tasks by date\n")
	s.WriteString("\nPress 1 or 2 to select, or 'q'/'esc' to skip.")
	return m.styles.AppStyle.Render(s.String())
}

func (m Model) calendarView() string {
	return renderModernCalendar(m.selectedDate, m.styles, &m)
}

func handleCalendarInput(current *time.Time, key string, m *Model) (*time.Time, bool) {
	selected := time.Date(m.calendarYear, time.Month(m.calendarMonth), m.calendarDay, 0, 0, 0, 0, time.Local)
	switch key {
	case "left":
		m.calendarDay--
		if m.calendarDay < 1 {
			m.calendarMonth--
			if m.calendarMonth < 1 {
				m.calendarMonth = 12
				m.calendarYear--
			}
			m.calendarDay = daysInMonth(m.calendarYear, m.calendarMonth)
		}
	case "right":
		m.calendarDay++
		if m.calendarDay > daysInMonth(m.calendarYear, m.calendarMonth) {
			m.calendarDay = 1
			m.calendarMonth++
			if m.calendarMonth > 12 {
				m.calendarMonth = 1
				m.calendarYear++
			}
		}
	case "up":
		m.calendarDay -= 7
		if m.calendarDay < 1 {
			m.calendarDay = 1
		}
	case "down":
		m.calendarDay += 7
		if m.calendarDay > daysInMonth(m.calendarYear, m.calendarMonth) {
			m.calendarDay = daysInMonth(m.calendarYear, m.calendarMonth)
		}
	case "enter":
		return &selected, true
	}
	return nil, false
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func renderModernCalendar(selected *time.Time, styles StyleConfig, m *Model) string {
	now := time.Now()
	year := m.calendarYear
	month := m.calendarMonth
	selectedDay := m.calendarDay
	firstOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	firstWeekday := int(firstOfMonth.Weekday())
	if firstWeekday == 0 {
		firstWeekday = 7
	}
	days := daysInMonth(year, month)

	cellWidth := 4
	if m.width > 80 {
		cellWidth = 6
	}
	cellStyle := lipgloss.NewStyle().Width(cellWidth).Align(lipgloss.Center)
	selectedStyle := lipgloss.NewStyle().
		Width(cellWidth).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#fff")).
		Background(lipgloss.Color("#7D56F4")).
		Bold(true).
		Margin(0).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, false, false)
	todayStyle := lipgloss.NewStyle().Width(cellWidth).Align(lipgloss.Center).Foreground(lipgloss.Color("#fff")).Background(lipgloss.Color("#43BF6D")).Bold(true).Padding(0, 0)
	defaultStyle := lipgloss.NewStyle().Width(cellWidth).Align(lipgloss.Center).Foreground(lipgloss.Color("#D9DCCF")).Padding(0, 0)

	var s strings.Builder
	s.WriteString(styles.TitleStyle.Render("📅 Select a Date to View Tasks"))
	s.WriteString("\n\n")
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render(fmt.Sprintf("%s %d", firstOfMonth.Month(), year)))
	s.WriteString("\n")
	weekdays := []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
	for _, wd := range weekdays {
		s.WriteString(cellStyle.Render(wd))
	}
	s.WriteString("\n")

	dayCounter := 1
	for week := 0; week < 6 && dayCounter <= days; week++ {
		for wd := 1; wd <= 7; wd++ {
			if week == 0 && wd < firstWeekday {
				s.WriteString(cellStyle.Render(" "))
			} else if dayCounter > days {
				s.WriteString(cellStyle.Render(" "))
			} else {
				isToday := dayCounter == now.Day() && month == int(now.Month()) && year == now.Year()
				isSelected := dayCounter == selectedDay
				var dayStr string
				if isSelected {
					dayStr = selectedStyle.Render(fmt.Sprintf("%2d", dayCounter))
				} else if isToday {
					dayStr = todayStyle.Render(fmt.Sprintf("%2d", dayCounter))
				} else {
					dayStr = defaultStyle.Render(fmt.Sprintf("%2d", dayCounter))
				}
				s.WriteString(dayStr)
				dayCounter++
			}
		}
		s.WriteString("\n")
	}

	s.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Render("Use arrow keys to pick a date, Enter to confirm, or Esc/Q to show all tasks."))
	return styles.AppStyle.Render(s.String())
}

func daysInMonth(year, month int) int {
	return time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.Local).Day()
}

func (m Model) detailsView() string {
	var s strings.Builder
	// Card style
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 4).
		Margin(2, 0).
		Width(60)

	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#fff"))
	iconStyle := lipgloss.NewStyle().Bold(true)

	// Status icon
	statusIcon := "📝"
	switch m.selectedTask.Status {
	case models.InProgress:
		statusIcon = "🔄"
	case models.Done:
		statusIcon = "✅"
	}
	// Priority icon
	priorityIcon := "🟢"
	switch m.selectedTask.Priority {
	case models.Medium:
		priorityIcon = "🟡"
	case models.High:
		priorityIcon = "🔴"
	}

	dueDateStr := "-"
	if m.selectedTask.DueDate != nil {
		dueDateStr = m.selectedTask.DueDate.Format("2006-01-02")
	}

	details := strings.Builder{}
	details.WriteString(labelStyle.Render("Title") + ": " + valueStyle.Render(m.selectedTask.Title) + "\n\n")
	details.WriteString(labelStyle.Render("Description") + ":\n" + valueStyle.Render(m.selectedTask.Description) + "\n\n")
	details.WriteString(labelStyle.Render("Status") + ": " + iconStyle.Render(statusIcon) + " " + valueStyle.Render(statusToString(m.selectedTask.Status)) + "\n")
	details.WriteString(labelStyle.Render("Priority") + ": " + iconStyle.Render(priorityIcon) + " " + valueStyle.Render(priorityToString(m.selectedTask.Priority)) + "\n")
	details.WriteString(labelStyle.Render("Due Date") + ": " + valueStyle.Render(dueDateStr) + "\n")
	if len(m.selectedTask.Tags) > 0 {
		details.WriteString(labelStyle.Render("Tags") + ": " + valueStyle.Render(strings.Join(m.selectedTask.Tags, ", ")) + "\n")
	}

	s.WriteString(cardStyle.Render(details.String()))
	s.WriteString("\n" + m.styles.InputStyle.Render("[e] Edit  [1] Todo  [2] In Progress  [3] Done  [Esc] Back"))
	return m.styles.AppStyle.Render(s.String())
}

func (m Model) editView() string {
	var s strings.Builder
	s.WriteString(m.styles.TitleStyle.Render("Edit Task"))
	s.WriteString("\n\n")
	s.WriteString("Title:\n" + m.styles.InputStyle.Render(m.editInputs[0].View()) + "\n")
	s.WriteString("Description:\n" + m.styles.InputStyle.Render(m.editInputs[1].View()) + "\n")
	s.WriteString("Priority (Low/Medium/High):\n" + m.styles.InputStyle.Render(m.editInputs[2].View()) + "\n")
	s.WriteString("Status (Todo/In Progress/Done):\n" + m.styles.InputStyle.Render(m.editInputs[3].View()) + "\n")
	s.WriteString("\n")
	s.WriteString(m.styles.InputStyle.Render("[Tab] Next  [Shift+Tab] Prev  [Enter] Save  [Esc] Cancel"))
	return m.styles.AppStyle.Render(s.String())
}

func priorityToString(p models.Priority) string {
	switch p {
	case models.Low:
		return "Low"
	case models.Medium:
		return "Medium"
	case models.High:
		return "High"
	}
	return "Low"
}

func stringToPriority(s string) models.Priority {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "low":
		return models.Low
	case "medium":
		return models.Medium
	case "high":
		return models.High
	}
	return models.Low
}

func statusToString(s models.Status) string {
	switch s {
	case models.Todo:
		return "Todo"
	case models.InProgress:
		return "In Progress"
	case models.Done:
		return "Done"
	}
	return "Todo"
}

func stringToStatus(s string) models.Status {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "todo":
		return models.Todo
	case "in progress":
		return models.InProgress
	case "done":
		return models.Done
	}
	return models.Todo
}
