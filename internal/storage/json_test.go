package storage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"taskflow/internal/task"
)

// newTestStorage returns a storage that works inside a temporary directory,
// so tests never touch the real data/tasks.json file.
func newTestStorage(t *testing.T) *JSONStorage {
	t.Helper()
	return NewJSONStorage(filepath.Join(t.TempDir(), "tasks.json"))
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}

func TestLoadCreatesMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "dir", "tasks.json")
	storage := NewJSONStorage(path)

	tasks, err := storage.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("got %d tasks, want 0", len(tasks))
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file was not created: %v", err)
	}
	if string(data) != "[]\n" {
		t.Errorf("file content = %q, want %q", data, "[]\n")
	}
}

func TestSaveAndLoad(t *testing.T) {
	storage := newTestStorage(t)
	createdAt := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)

	want := []task.Task{
		{ID: 1, Title: "Learn Golang", Completed: true, Priority: task.PriorityHigh, CreatedAt: createdAt, UpdatedAt: createdAt},
		{ID: 2, Title: "Изучить JSON", Completed: false, Priority: task.PriorityLow, CreatedAt: createdAt, UpdatedAt: createdAt},
	}

	if err := storage.Save(want); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	got, err := storage.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if len(got) != len(want) {
		t.Fatalf("got %d tasks, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].ID != want[i].ID ||
			got[i].Title != want[i].Title ||
			got[i].Completed != want[i].Completed ||
			got[i].Priority != want[i].Priority ||
			!got[i].CreatedAt.Equal(want[i].CreatedAt) ||
			!got[i].UpdatedAt.Equal(want[i].UpdatedAt) {
			t.Errorf("task %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestSaveKeepsExistingTasks(t *testing.T) {
	storage := newTestStorage(t)

	for _, title := range []string{"First", "Second", "Third"} {
		tasks, err := storage.Load()
		if err != nil {
			t.Fatalf("Load returned error: %v", err)
		}

		service := task.NewService(tasks)
		if _, err := service.Add(title, ""); err != nil {
			t.Fatalf("Add returned error: %v", err)
		}

		if err := storage.Save(service.Tasks()); err != nil {
			t.Fatalf("Save returned error: %v", err)
		}
	}

	tasks, err := storage.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("got %d tasks, want 3", len(tasks))
	}
	if tasks[2].ID != 3 || tasks[2].Title != "Third" {
		t.Errorf("last task = %+v, want ID 3 with title Third", tasks[2])
	}
}

func TestSaveNilWritesEmptyArray(t *testing.T) {
	storage := newTestStorage(t)

	if err := storage.Save(nil); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	data, err := os.ReadFile(storage.Path())
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(data) != "[]\n" {
		t.Errorf("file content = %q, want %q", data, "[]\n")
	}
}

func TestLoadEmptyFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{name: "empty", content: ""},
		{name: "whitespace", content: "  \n\t\n"},
		{name: "empty array", content: "[]"},
		{name: "null", content: "null"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := newTestStorage(t)
			writeFile(t, storage.Path(), tt.content)

			tasks, err := storage.Load()
			if err != nil {
				t.Fatalf("Load returned error: %v", err)
			}
			if tasks == nil || len(tasks) != 0 {
				t.Errorf("got %v, want an empty non-nil slice", tasks)
			}
		})
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{name: "broken syntax", content: `[{"id": 1, "title": "Task"`},
		{name: "plain text", content: "hello"},
		{name: "object instead of array", content: `{"id": 1}`},
		{name: "wrong field type", content: `[{"id": "one", "title": "Task", "priority": "low"}]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := newTestStorage(t)
			writeFile(t, storage.Path(), tt.content)

			_, err := storage.Load()
			if !errors.Is(err, ErrInvalidJSON) {
				t.Fatalf("error = %v, want %v", err, ErrInvalidJSON)
			}

			data, _ := os.ReadFile(storage.Path())
			if string(data) != tt.content {
				t.Error("Load must not change a corrupted file")
			}
		})
	}
}

func TestLoadInvalidTaskData(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{name: "zero ID", content: `[{"id": 0, "title": "Task", "priority": "low"}]`},
		{name: "empty title", content: `[{"id": 1, "title": "", "priority": "low"}]`},
		{name: "unknown priority", content: `[{"id": 1, "title": "Task", "priority": "urgent"}]`},
		{
			name: "duplicate ID",
			content: `[{"id": 1, "title": "First", "priority": "low"},
			           {"id": 1, "title": "Second", "priority": "low"}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := newTestStorage(t)
			writeFile(t, storage.Path(), tt.content)

			if _, err := storage.Load(); !errors.Is(err, task.ErrInvalidTask) {
				t.Errorf("error = %v, want %v", err, task.ErrInvalidTask)
			}
		})
	}
}

func TestLoadReadError(t *testing.T) {
	// A directory cannot be read as a file.
	storage := NewJSONStorage(t.TempDir())

	_, err := storage.Load()
	if err == nil {
		t.Fatal("expected an error when the path is a directory")
	}
	if errors.Is(err, ErrInvalidJSON) {
		t.Errorf("read error must not be reported as invalid JSON: %v", err)
	}
}

func TestSaveWriteError(t *testing.T) {
	// The parent "directory" is a regular file, so the tasks file cannot be created.
	parentFile := filepath.Join(t.TempDir(), "not-a-directory")
	writeFile(t, parentFile, "")
	storage := NewJSONStorage(filepath.Join(parentFile, "tasks.json"))

	if err := storage.Save([]task.Task{}); err == nil {
		t.Fatal("expected an error when the directory cannot be created")
	}
}
