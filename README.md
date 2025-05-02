# tasker - A Beautiful Task Management TUI

A beautiful and efficient task management Terminal User Interface (TUI) application built with Go and Bubble Tea.

## Features

- 📋 Task list view with status indicators
- ✏️ Task creation and editing
- 🏷️ Task categorization with tags/priorities
- 🔍 Task filtering and sorting
- ⌨️ Keyboard shortcuts for efficient navigation
- 🎨 Beautiful styling with bubbles and borders

## Installation

1. Ensure you have Go 1.16 or later installed
2. Clone this repository
3. Run `go install` in the project directory

## Usage

```bash
tasker
```

### Keyboard Shortcuts

- `j/k` or `↑/↓`: Navigate tasks
- `n`: New task
- `e`: Edit selected task
- `d`: Delete selected task
- `v`: View task details
- `Space`: Toggle task completion
- `/`: Enter search mode
- `s`: Enter status filter mode
- `q`: Quit application

### Task Details

View detailed information about a task and change its status:
- Press `v` to view task details
- In the details view, you can see:
  - Task title and description
  - Current status
  - Priority level
  - Due date (if set)
  - Tags (if any)
- While viewing details, you can:
  - Press `1` to set status to Todo
  - Press `2` to set status to In Progress
  - Press `3` to set status to Done
  - Press `Esc` to return to the task list

### Status Filtering

Tasks can be filtered by their status:
- Press `s` to enter status filter mode
- Press `1` to show only Todo tasks
- Press `2` to show only In Progress tasks
- Press `3` to show only Done tasks
- Press `0` to clear status filter
- Press `Esc` to cancel status filter selection

Each task is displayed with a status indicator:
- 📝 Todo
- 🔄 In Progress
- ✅ Done

## Development

To build from source:

```bash
go build -o tasker ./cmd/tasker
```

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions 