package task

import (
	"errors"
	"testing"
	"time"
)

// newServiceWithTasks creates a service with three tasks:
// 1 "Learn Golang" (done, high), 2 "Build REST API" (pending, medium),
// 3 "Read Go documentation" (pending, low).
func newServiceWithTasks(t *testing.T) *Service {
	t.Helper()

	service := NewService(nil)
	mustAdd(t, service, "Learn Golang", PriorityHigh)
	mustAdd(t, service, "Build REST API", PriorityMedium)
	mustAdd(t, service, "Read Go documentation", PriorityLow)

	if _, err := service.Complete(1); err != nil {
		t.Fatalf("Complete(1) returned error: %v", err)
	}
	return service
}

func mustAdd(t *testing.T, service *Service, title, priority string) Task {
	t.Helper()

	added, err := service.Add(title, priority)
	if err != nil {
		t.Fatalf("Add(%q, %q) returned error: %v", title, priority, err)
	}
	return added
}

func TestAddTask(t *testing.T) {
	service := NewService(nil)
	before := time.Now()

	added, err := service.Add("  Learn Golang  ", "")
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if added.ID != 1 {
		t.Errorf("ID = %d, want 1", added.ID)
	}
	if added.Title != "Learn Golang" {
		t.Errorf("Title = %q, want %q", added.Title, "Learn Golang")
	}
	if added.Priority != PriorityMedium {
		t.Errorf("Priority = %q, want default %q", added.Priority, PriorityMedium)
	}
	if added.Completed {
		t.Error("new task must not be completed")
	}
	if added.CreatedAt.Before(before) || !added.UpdatedAt.Equal(added.CreatedAt) {
		t.Errorf("unexpected timestamps: created %v, updated %v", added.CreatedAt, added.UpdatedAt)
	}
	if len(service.Tasks()) != 1 {
		t.Errorf("service has %d tasks, want 1", len(service.Tasks()))
	}
}

func TestAddTaskPriority(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "low", want: PriorityLow},
		{input: "medium", want: PriorityMedium},
		{input: "high", want: PriorityHigh},
		{input: "HIGH", want: PriorityHigh},
		{input: " Low ", want: PriorityLow},
		{input: "", want: PriorityMedium},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			added, err := NewService(nil).Add("Task", tt.input)
			if err != nil {
				t.Fatalf("Add returned error: %v", err)
			}
			if added.Priority != tt.want {
				t.Errorf("Priority = %q, want %q", added.Priority, tt.want)
			}
		})
	}
}

func TestAddTaskInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		priority string
		wantErr  error
	}{
		{name: "empty title", title: "", priority: "low", wantErr: ErrEmptyTitle},
		{name: "spaces only", title: "   ", priority: "low", wantErr: ErrEmptyTitle},
		{name: "unknown priority", title: "Task", priority: "urgent", wantErr: ErrInvalidPriority},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(nil)

			_, err := service.Add(tt.title, tt.priority)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if len(service.Tasks()) != 0 {
				t.Error("invalid task must not be added")
			}
		})
	}
}

func TestAddTaskKeepsExistingTasks(t *testing.T) {
	service := newServiceWithTasks(t)

	added := mustAdd(t, service, "New task", PriorityLow)

	if added.ID != 4 {
		t.Errorf("ID = %d, want 4", added.ID)
	}
	if len(service.Tasks()) != 4 {
		t.Errorf("service has %d tasks, want 4", len(service.Tasks()))
	}
}

func TestAddTaskDoesNotReuseIDInTheMiddle(t *testing.T) {
	service := newServiceWithTasks(t)
	if _, err := service.Delete(2); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	added := mustAdd(t, service, "New task", PriorityLow)

	if added.ID != 4 {
		t.Errorf("ID = %d, want 4", added.ID)
	}
}

func TestGetTask(t *testing.T) {
	service := newServiceWithTasks(t)

	found, err := service.Get(2)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if found.Title != "Build REST API" {
		t.Errorf("Title = %q, want %q", found.Title, "Build REST API")
	}

	if _, err := service.Get(100); !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("Get(100) error = %v, want %v", err, ErrTaskNotFound)
	}
}

func TestCompleteTask(t *testing.T) {
	service := newServiceWithTasks(t)

	completed, err := service.Complete(2)
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
	if !completed.Completed {
		t.Error("task must be completed")
	}
	if completed.UpdatedAt.Before(completed.CreatedAt) {
		t.Error("UpdatedAt must not be before CreatedAt")
	}

	stored, _ := service.Get(2)
	if !stored.Completed {
		t.Error("change must be stored in the service")
	}
}

func TestCompleteTaskErrors(t *testing.T) {
	service := newServiceWithTasks(t)

	if _, err := service.Complete(1); !errors.Is(err, ErrAlreadyCompleted) {
		t.Errorf("Complete(1) error = %v, want %v", err, ErrAlreadyCompleted)
	}
	if _, err := service.Complete(100); !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("Complete(100) error = %v, want %v", err, ErrTaskNotFound)
	}
}

func TestUndoTask(t *testing.T) {
	service := newServiceWithTasks(t)

	undone, err := service.Undo(1)
	if err != nil {
		t.Fatalf("Undo returned error: %v", err)
	}
	if undone.Completed {
		t.Error("task must be pending after undo")
	}

	if _, err := service.Undo(1); !errors.Is(err, ErrNotCompleted) {
		t.Errorf("second Undo error = %v, want %v", err, ErrNotCompleted)
	}
	if _, err := service.Undo(100); !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("Undo(100) error = %v, want %v", err, ErrTaskNotFound)
	}
}

func TestDeleteTask(t *testing.T) {
	service := newServiceWithTasks(t)

	deleted, err := service.Delete(2)
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if deleted.ID != 2 {
		t.Errorf("deleted ID = %d, want 2", deleted.ID)
	}

	tasks := service.Tasks()
	if len(tasks) != 2 {
		t.Fatalf("service has %d tasks, want 2", len(tasks))
	}
	if tasks[0].ID != 1 || tasks[1].ID != 3 {
		t.Errorf("remaining IDs = %d, %d, want 1, 3", tasks[0].ID, tasks[1].ID)
	}

	if _, err := service.Delete(2); !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("second Delete error = %v, want %v", err, ErrTaskNotFound)
	}
}

func TestEditTask(t *testing.T) {
	service := newServiceWithTasks(t)

	edited, err := service.Edit(1, "  Learn Go interfaces ")
	if err != nil {
		t.Fatalf("Edit returned error: %v", err)
	}
	if edited.Title != "Learn Go interfaces" {
		t.Errorf("Title = %q, want %q", edited.Title, "Learn Go interfaces")
	}
	if !edited.Completed || edited.Priority != PriorityHigh {
		t.Error("Edit must change only the title")
	}
}

func TestEditTaskErrors(t *testing.T) {
	service := newServiceWithTasks(t)

	if _, err := service.Edit(1, "   "); !errors.Is(err, ErrEmptyTitle) {
		t.Errorf("Edit with empty title error = %v, want %v", err, ErrEmptyTitle)
	}
	if _, err := service.Edit(100, "Title"); !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("Edit(100) error = %v, want %v", err, ErrTaskNotFound)
	}

	unchanged, _ := service.Get(1)
	if unchanged.Title != "Learn Golang" {
		t.Errorf("title changed after failed edit: %q", unchanged.Title)
	}
}

func TestSearchTasks(t *testing.T) {
	service := newServiceWithTasks(t)

	tests := []struct {
		query   string
		wantIDs []int
	}{
		{query: "golang", wantIDs: []int{1}},
		{query: "GO", wantIDs: []int{1, 3}},
		{query: "  rest ", wantIDs: []int{2}},
		{query: "python", wantIDs: nil},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			found, err := service.Search(tt.query)
			if err != nil {
				t.Fatalf("Search returned error: %v", err)
			}
			assertIDs(t, found, tt.wantIDs)
		})
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	service := newServiceWithTasks(t)

	if _, err := service.Search("  "); !errors.Is(err, ErrEmptyQuery) {
		t.Errorf("error = %v, want %v", err, ErrEmptyQuery)
	}
}

func TestFilterTasks(t *testing.T) {
	service := newServiceWithTasks(t)

	tests := []struct {
		status  string
		wantIDs []int
	}{
		{status: "", wantIDs: []int{1, 2, 3}},
		{status: "all", wantIDs: []int{1, 2, 3}},
		{status: "done", wantIDs: []int{1}},
		{status: "pending", wantIDs: []int{2, 3}},
		{status: "DONE", wantIDs: []int{1}},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			filtered, err := service.Filter(tt.status)
			if err != nil {
				t.Fatalf("Filter returned error: %v", err)
			}
			assertIDs(t, filtered, tt.wantIDs)
		})
	}
}

func TestFilterInvalidStatus(t *testing.T) {
	service := newServiceWithTasks(t)

	if _, err := service.Filter("finished"); !errors.Is(err, ErrInvalidStatus) {
		t.Errorf("error = %v, want %v", err, ErrInvalidStatus)
	}
}

func TestStats(t *testing.T) {
	stats := newServiceWithTasks(t).Stats()

	want := Stats{Total: 3, Completed: 1, Pending: 2}
	if stats != want {
		t.Errorf("Stats = %+v, want %+v", stats, want)
	}

	percent := stats.CompletionPercent()
	if percent < 33.3 || percent > 33.4 {
		t.Errorf("CompletionPercent = %.2f, want about 33.33", percent)
	}
}

func TestStatsEmpty(t *testing.T) {
	stats := NewService(nil).Stats()

	if stats != (Stats{}) {
		t.Errorf("Stats = %+v, want all zeros", stats)
	}
	if stats.CompletionPercent() != 0 {
		t.Errorf("CompletionPercent = %v, want 0", stats.CompletionPercent())
	}
}

func TestNewServiceCopiesTasks(t *testing.T) {
	original := []Task{
		{ID: 1, Title: "First", Priority: PriorityLow},
		{ID: 2, Title: "Second", Priority: PriorityLow},
	}
	service := NewService(original)

	if _, err := service.Delete(1); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	if original[0].ID != 1 || original[1].ID != 2 {
		t.Error("service must not change the slice passed to NewService")
	}
}

func TestTaskValidate(t *testing.T) {
	tests := []struct {
		name    string
		task    Task
		wantErr bool
	}{
		{name: "valid", task: Task{ID: 1, Title: "Task", Priority: PriorityLow}, wantErr: false},
		{name: "zero ID", task: Task{ID: 0, Title: "Task", Priority: PriorityLow}, wantErr: true},
		{name: "negative ID", task: Task{ID: -5, Title: "Task", Priority: PriorityLow}, wantErr: true},
		{name: "empty title", task: Task{ID: 1, Title: " ", Priority: PriorityLow}, wantErr: true},
		{name: "bad priority", task: Task{ID: 1, Title: "Task", Priority: "urgent"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if tt.wantErr && !errors.Is(err, ErrInvalidTask) {
				t.Errorf("error = %v, want %v", err, ErrInvalidTask)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// assertIDs checks that tasks have exactly the expected IDs in the same order.
func assertIDs(t *testing.T, tasks []Task, wantIDs []int) {
	t.Helper()

	if len(tasks) != len(wantIDs) {
		t.Fatalf("got %d tasks, want %d", len(tasks), len(wantIDs))
	}
	for i, task := range tasks {
		if task.ID != wantIDs[i] {
			t.Errorf("task[%d].ID = %d, want %d", i, task.ID, wantIDs[i])
		}
	}
}
