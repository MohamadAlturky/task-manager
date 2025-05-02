package ui

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	// Fixed heights for components
	defaultListHeight = 15
	defaultAppHeight  = 20
)

var (
	// Colors
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	// Base styles that will be modified based on window size
	BaseAppStyle = lipgloss.NewStyle().
			Padding(1, 2, 1, 2)

	BaseTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(highlight)

	BaseListStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(highlight).
			Padding(1, 1)

	BaseListItemStyle = lipgloss.NewStyle().
				PaddingLeft(2)

	BaseSelectedItemStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(special).
				Padding(0, 1)

	BaseDoneStyle = lipgloss.NewStyle().
			Strikethrough(true).
			Foreground(subtle)

	BaseTagStyle = lipgloss.NewStyle().
			Foreground(special).
			Padding(0, 1, 0, 1)

	BasePriorityHighStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF0000")).
				Bold(true)

	BasePriorityMediumStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFA500")).
				Bold(true)

	BasePriorityLowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00"))

	BaseStatusBarStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(highlight).
				Padding(0, 1)

	BaseErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	BaseInputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(highlight).
			Padding(0, 1)
)

// StyleConfig holds responsive styles based on window dimensions
type StyleConfig struct {
	AppStyle            lipgloss.Style
	TitleStyle          lipgloss.Style
	ListStyle           lipgloss.Style
	ListItemStyle       lipgloss.Style
	SelectedItemStyle   lipgloss.Style
	DoneStyle           lipgloss.Style
	TagStyle            lipgloss.Style
	PriorityHighStyle   lipgloss.Style
	PriorityMediumStyle lipgloss.Style
	PriorityLowStyle    lipgloss.Style
	StatusBarStyle      lipgloss.Style
	ErrorStyle          lipgloss.Style
	InputStyle          lipgloss.Style
}

// NewStyleConfig creates a new StyleConfig with responsive dimensions
func NewStyleConfig(width int) StyleConfig {
	// Calculate responsive widths
	contentWidth := width - 4 // Account for app padding
	if contentWidth < 40 {    // Minimum width to maintain readability
		contentWidth = 40
	}

	listWidth := contentWidth - 2  // Account for list borders
	itemWidth := listWidth - 4     // Account for item padding
	inputWidth := contentWidth - 4 // Account for input borders

	return StyleConfig{
		AppStyle: BaseAppStyle.Copy().
			Width(width),

		TitleStyle: BaseTitleStyle.Copy().
			Width(contentWidth).
			Align(lipgloss.Center),

		ListStyle: BaseListStyle.Copy().
			Width(listWidth).
			Height(defaultListHeight),

		ListItemStyle: BaseListItemStyle.Copy().
			Width(itemWidth),

		SelectedItemStyle: BaseSelectedItemStyle.Copy().
			Width(itemWidth - 2), // Account for selected item border

		DoneStyle:           BaseDoneStyle.Copy(),
		TagStyle:            BaseTagStyle.Copy(),
		PriorityHighStyle:   BasePriorityHighStyle.Copy(),
		PriorityMediumStyle: BasePriorityMediumStyle.Copy(),
		PriorityLowStyle:    BasePriorityLowStyle.Copy(),

		StatusBarStyle: BaseStatusBarStyle.Copy().
			Width(contentWidth),

		ErrorStyle: BaseErrorStyle.Copy().
			Width(contentWidth),

		InputStyle: BaseInputStyle.Copy().
			Width(inputWidth),
	}
}
