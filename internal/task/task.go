// Package task contains the Task model and the business logic for managing tasks.
package task

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Available task priorities.
const (
	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"
)

// Available task statuses used for filtering.
const (
	StatusAll     = "all"
	StatusDone    = "done"
	StatusPending = "pending"
)

// Errors returned by the task package. They can be checked with errors.Is.
var (
	ErrTaskNotFound     = errors.New("task not found")
	ErrEmptyTitle       = errors.New("task title cannot be empty")
	ErrInvalidPriority  = errors.New("invalid priority")
	ErrInvalidStatus    = errors.New("invalid status")
	ErrAlreadyCompleted = errors.New("task is already completed")
	ErrNotCompleted     = errors.New("task is not completed")
	ErrEmptyQuery       = errors.New("search query cannot be empty")
	ErrInvalidTask      = errors.New("invalid task data")
)

// Task describes a single item in the task list.
type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	Priority  string    `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Status returns "done" for completed tasks and "pending" for active ones.
func (t Task) Status() string {
	if t.Completed {
		return StatusDone
	}
	return StatusPending
}

// Validate checks that a task loaded from storage contains correct data.
func (t Task) Validate() error {
	if t.ID <= 0 {
		return fmt.Errorf("%w: ID must be positive, got %d", ErrInvalidTask, t.ID)
	}
	if strings.TrimSpace(t.Title) == "" {
		return fmt.Errorf("%w: task %d has an empty title", ErrInvalidTask, t.ID)
	}
	if !IsValidPriority(t.Priority) {
		return fmt.Errorf("%w: task %d has unknown priority %q", ErrInvalidTask, t.ID, t.Priority)
	}
	return nil
}

// IsValidPriority reports whether the priority is one of low, medium or high.
func IsValidPriority(priority string) bool {
	switch priority {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	default:
		return false
	}
}

// NormalizePriority converts user input to a valid priority.
// An empty value means the default priority (medium).
// The check is case-insensitive, so "HIGH" and "high" are both accepted.
func NormalizePriority(priority string) (string, error) {
	priority = strings.ToLower(strings.TrimSpace(priority))
	if priority == "" {
		return PriorityMedium, nil
	}
	if !IsValidPriority(priority) {
		return "", fmt.Errorf("%w %q: use low, medium or high", ErrInvalidPriority, priority)
	}
	return priority, nil
}

// normalizeTitle removes extra spaces around the title and checks that it is not empty.
func normalizeTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", ErrEmptyTitle
	}
	return title, nil
}
