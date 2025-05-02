package models

import (
	"time"
)

type Priority int

const (
	Low Priority = iota
	Medium
	High
)

type Status int

const (
	Todo Status = iota
	InProgress
	Done
)

type Task struct {
	ID          int
	Title       string
	Description string
	Done        bool
	CreatedAt   time.Time
	DueDate     *time.Time
	Priority    Priority
	Status      Status
	Tags        []string
}

type TaskList struct {
	Tasks       []*Task
	Selected    int
	PageSize    int
	CurrentPage int
	TotalPages  int
}

func NewTask(title string, description string, priority Priority, tags []string) *Task {
	return &Task{
		Title:       title,
		Description: description,
		Done:        false,
		CreatedAt:   time.Now(),
		Priority:    priority,
		Status:      Todo,
		Tags:        tags,
	}
}

func NewTaskList() *TaskList {
	return &TaskList{
		Tasks:       make([]*Task, 0),
		Selected:    0,
		PageSize:    10,
		CurrentPage: 1,
		TotalPages:  1,
	}
}

func (tl *TaskList) AddTask(task *Task) {
	task.ID = len(tl.Tasks)
	tl.Tasks = append(tl.Tasks, task)
}

func (tl *TaskList) RemoveTask(index int) {
	if index < 0 || index >= len(tl.Tasks) {
		return
	}
	tl.Tasks = append(tl.Tasks[:index], tl.Tasks[index+1:]...)
	if tl.Selected >= len(tl.Tasks) {
		tl.Selected = len(tl.Tasks) - 1
	}
	if tl.Selected < 0 {
		tl.Selected = 0
	}
}

func (tl *TaskList) ToggleTask(index int) {
	if index < 0 || index >= len(tl.Tasks) {
		return
	}
	tl.Tasks[index].Done = !tl.Tasks[index].Done
}

func (t *Task) SetDueDate(date time.Time) {
	t.DueDate = &date
}

func (t *Task) ClearDueDate() {
	t.DueDate = nil
}

func (t *Task) AddTag(tag string) {
	for _, existingTag := range t.Tags {
		if existingTag == tag {
			return
		}
	}
	t.Tags = append(t.Tags, tag)
}

func (t *Task) RemoveTag(tag string) {
	for i, existingTag := range t.Tags {
		if existingTag == tag {
			t.Tags = append(t.Tags[:i], t.Tags[i+1:]...)
			return
		}
	}
}

func (tl *TaskList) UpdatePagination() {
	if len(tl.Tasks) == 0 {
		tl.TotalPages = 1
		tl.CurrentPage = 1
		return
	}
	tl.TotalPages = (len(tl.Tasks) + tl.PageSize - 1) / tl.PageSize
	if tl.CurrentPage > tl.TotalPages {
		tl.CurrentPage = tl.TotalPages
	}
}

func (tl *TaskList) GetCurrentPageTasks() []*Task {
	start := (tl.CurrentPage - 1) * tl.PageSize
	end := start + tl.PageSize
	if end > len(tl.Tasks) {
		end = len(tl.Tasks)
	}
	if start >= len(tl.Tasks) {
		return []*Task{}
	}
	return tl.Tasks[start:end]
}

func (tl *TaskList) NextPage() {
	if tl.CurrentPage < tl.TotalPages {
		tl.CurrentPage++
		tl.Selected = 0
	}
}

func (tl *TaskList) PrevPage() {
	if tl.CurrentPage > 1 {
		tl.CurrentPage--
		tl.Selected = 0
	}
}
