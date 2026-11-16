package task

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

// Stats contains summary information about tasks.
type Stats struct {
	Total     int
	Completed int
	Pending   int
}

// CompletionPercent returns the share of completed tasks in percent (0-100).
func (s Stats) CompletionPercent() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.Completed) * 100 / float64(s.Total)
}

// Service performs operations on a list of tasks kept in memory.
// Saving and loading the list is the responsibility of the storage package.
type Service struct {
	tasks []Task
}

// NewService creates a service from already loaded tasks.
// The slice is copied, so the service never changes the caller's data.
func NewService(tasks []Task) *Service {
	return &Service{tasks: slices.Clone(tasks)}
}

// Tasks returns a copy of all tasks.
func (s *Service) Tasks() []Task {
	return slices.Clone(s.tasks)
}

// Add creates a new task and returns it.
// An empty priority means the default priority (medium).
func (s *Service) Add(title, priority string) (Task, error) {
	title, err := normalizeTitle(title)
	if err != nil {
		return Task{}, err
	}

	priority, err = NormalizePriority(priority)
	if err != nil {
		return Task{}, err
	}

	now := time.Now()
	newTask := Task{
		ID:        s.nextID(),
		Title:     title,
		Completed: false,
		Priority:  priority,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.tasks = append(s.tasks, newTask)

	return newTask, nil
}

// Get returns the task with the given ID.
func (s *Service) Get(id int) (Task, error) {
	index, err := s.findIndex(id)
	if err != nil {
		return Task{}, err
	}
	return s.tasks[index], nil
}

// Complete marks the task as done.
func (s *Service) Complete(id int) (Task, error) {
	index, err := s.findIndex(id)
	if err != nil {
		return Task{}, err
	}
	if s.tasks[index].Completed {
		return Task{}, fmt.Errorf("%w: id %d", ErrAlreadyCompleted, id)
	}

	s.tasks[index].Completed = true
	s.tasks[index].UpdatedAt = time.Now()
	return s.tasks[index], nil
}

// Undo marks a completed task as pending again.
func (s *Service) Undo(id int) (Task, error) {
	index, err := s.findIndex(id)
	if err != nil {
		return Task{}, err
	}
	if !s.tasks[index].Completed {
		return Task{}, fmt.Errorf("%w: id %d", ErrNotCompleted, id)
	}

	s.tasks[index].Completed = false
	s.tasks[index].UpdatedAt = time.Now()
	return s.tasks[index], nil
}

// Edit changes the title of the task.
func (s *Service) Edit(id int, title string) (Task, error) {
	title, err := normalizeTitle(title)
	if err != nil {
		return Task{}, err
	}

	index, err := s.findIndex(id)
	if err != nil {
		return Task{}, err
	}

	s.tasks[index].Title = title
	s.tasks[index].UpdatedAt = time.Now()
	return s.tasks[index], nil
}

// Delete removes the task and returns the removed task.
func (s *Service) Delete(id int) (Task, error) {
	index, err := s.findIndex(id)
	if err != nil {
		return Task{}, err
	}

	deleted := s.tasks[index]
	s.tasks = slices.Delete(s.tasks, index, index+1)
	return deleted, nil
}

// Search returns tasks whose title contains the query (case-insensitive).
func (s *Service) Search(query string) ([]Task, error) {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil, ErrEmptyQuery
	}

	var found []Task
	for _, t := range s.tasks {
		if strings.Contains(strings.ToLower(t.Title), query) {
			found = append(found, t)
		}
	}
	return found, nil
}

// Filter returns tasks with the given status: "all", "done" or "pending".
// An empty status is treated as "all".
func (s *Service) Filter(status string) ([]Task, error) {
	status = strings.ToLower(strings.TrimSpace(status))

	switch status {
	case "", StatusAll:
		return s.Tasks(), nil
	case StatusDone, StatusPending:
		var filtered []Task
		for _, t := range s.tasks {
			if t.Status() == status {
				filtered = append(filtered, t)
			}
		}
		return filtered, nil
	default:
		return nil, fmt.Errorf("%w %q: use all, done or pending", ErrInvalidStatus, status)
	}
}

// Stats counts total, completed and pending tasks.
func (s *Service) Stats() Stats {
	stats := Stats{Total: len(s.tasks)}
	for _, t := range s.tasks {
		if t.Completed {
			stats.Completed++
		}
	}
	stats.Pending = stats.Total - stats.Completed
	return stats
}

// findIndex returns the position of the task with the given ID in the slice.
func (s *Service) findIndex(id int) (int, error) {
	for i, t := range s.tasks {
		if t.ID == id {
			return i, nil
		}
	}
	return -1, fmt.Errorf("%w: id %d", ErrTaskNotFound, id)
}

// nextID returns an ID that is greater than any existing one.
func (s *Service) nextID() int {
	maxID := 0
	for _, t := range s.tasks {
		if t.ID > maxID {
			maxID = t.ID
		}
	}
	return maxID + 1
}
