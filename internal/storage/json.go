// Package storage saves and loads tasks from a JSON file.
package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"taskflow/internal/task"
)

// ErrInvalidJSON is returned when the tasks file cannot be parsed.
var ErrInvalidJSON = errors.New("tasks file contains invalid JSON")

// JSONStorage keeps tasks in a JSON file on disk.
type JSONStorage struct {
	path string
}

// NewJSONStorage creates a storage that works with the file at path.
func NewJSONStorage(path string) *JSONStorage {
	return &JSONStorage{path: path}
}

// Path returns the location of the tasks file.
func (s *JSONStorage) Path() string {
	return s.path
}

// Load reads all tasks from the file.
//
// If the file does not exist, it is created with an empty task list.
// An empty file (or a file with only whitespace) is treated as an empty list.
func (s *JSONStorage) Load() ([]task.Task, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		if err := s.Save([]task.Task{}); err != nil {
			return nil, fmt.Errorf("create tasks file: %w", err)
		}
		return []task.Task{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read tasks file %s: %w", s.path, err)
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return []task.Task{}, nil
	}

	var tasks []task.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("%w (%s): %v; fix or remove the file", ErrInvalidJSON, s.path, err)
	}

	// The file may contain the JSON value null.
	if tasks == nil {
		tasks = []task.Task{}
	}

	if err := validateTasks(tasks); err != nil {
		return nil, fmt.Errorf("check tasks file %s: %w", s.path, err)
	}

	return tasks, nil
}

// Save writes all tasks to the file, replacing its previous content.
//
// Data is first written to a temporary file and then renamed, so the
// existing file is not damaged if writing fails in the middle.
func (s *JSONStorage) Save(tasks []task.Task) error {
	if tasks == nil {
		tasks = []task.Task{}
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("encode tasks: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	tempPath := s.path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o644); err != nil {
		return fmt.Errorf("write tasks file %s: %w", tempPath, err)
	}

	if err := os.Rename(tempPath, s.path); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("replace tasks file %s: %w", s.path, err)
	}

	return nil
}

// validateTasks checks every task and makes sure that IDs are unique.
func validateTasks(tasks []task.Task) error {
	seenIDs := make(map[int]bool, len(tasks))

	for _, t := range tasks {
		if err := t.Validate(); err != nil {
			return err
		}
		if seenIDs[t.ID] {
			return fmt.Errorf("%w: duplicate task ID %d", task.ErrInvalidTask, t.ID)
		}
		seenIDs[t.ID] = true
	}

	return nil
}
