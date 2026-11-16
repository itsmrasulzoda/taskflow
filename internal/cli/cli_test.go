package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"taskflow/internal/storage"
	"taskflow/internal/task"
)

// testApp holds an application that writes to a temporary file and a buffer.
type testApp struct {
	app    *App
	output *bytes.Buffer
	path   string
}

func newTestApp(t *testing.T) testApp {
	t.Helper()

	path := filepath.Join(t.TempDir(), "tasks.json")
	output := &bytes.Buffer{}
	return testApp{app: NewApp(path, output), output: output, path: path}
}

// run executes a command and fails the test if it returns an error.
func (ta testApp) run(t *testing.T, args ...string) string {
	t.Helper()

	ta.output.Reset()
	if err := ta.app.Run(args); err != nil {
		t.Fatalf("Run(%q) returned error: %v", args, err)
	}
	return ta.output.String()
}

// runWithError executes a command and fails the test if it succeeds.
func (ta testApp) runWithError(t *testing.T, args ...string) error {
	t.Helper()

	ta.output.Reset()
	err := ta.app.Run(args)
	if err == nil {
		t.Fatalf("Run(%q) succeeded, want an error", args)
	}
	return err
}

// savedTasks reads tasks directly from the test file.
func (ta testApp) savedTasks(t *testing.T) []task.Task {
	t.Helper()

	tasks, err := storage.NewJSONStorage(ta.path).Load()
	if err != nil {
		t.Fatalf("load saved tasks: %v", err)
	}
	return tasks
}

func TestRunWithoutArguments(t *testing.T) {
	ta := newTestApp(t)

	err := ta.runWithError(t)
	if !strings.Contains(err.Error(), "no command") {
		t.Errorf("error = %v, want message about missing command", err)
	}
	if !strings.Contains(ta.output.String(), "Usage:") {
		t.Error("usage must be printed when no command is given")
	}
}

func TestRunUnknownCommand(t *testing.T) {
	ta := newTestApp(t)

	err := ta.runWithError(t, "fly")
	if !strings.Contains(err.Error(), `unknown command "fly"`) {
		t.Errorf("error = %v", err)
	}
}

func TestRunHelp(t *testing.T) {
	ta := newTestApp(t)

	for _, command := range []string{"help", "-h", "--help"} {
		output := ta.run(t, command)
		if !strings.Contains(output, "Commands:") {
			t.Errorf("%s: help text not printed", command)
		}
	}

	if _, err := os.Stat(ta.path); err == nil {
		t.Error("help must not create the data file")
	}
}

func TestAddCommand(t *testing.T) {
	ta := newTestApp(t)

	output := ta.run(t, "add", "Prepare project", "--priority", "high")
	if !strings.Contains(output, "Task #1 added") {
		t.Errorf("unexpected output: %q", output)
	}

	ta.run(t, "add", "--priority=low", "Read", "documentation")
	ta.run(t, "add", "Learn Golang")

	tasks := ta.savedTasks(t)
	if len(tasks) != 3 {
		t.Fatalf("saved %d tasks, want 3", len(tasks))
	}

	want := []struct {
		title    string
		priority string
	}{
		{"Prepare project", task.PriorityHigh},
		{"Read documentation", task.PriorityLow},
		{"Learn Golang", task.PriorityMedium},
	}
	for i, w := range want {
		if tasks[i].Title != w.title || tasks[i].Priority != w.priority {
			t.Errorf("task %d = %q/%q, want %q/%q", i+1, tasks[i].Title, tasks[i].Priority, w.title, w.priority)
		}
	}
}

func TestAddCommandErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "no title", args: []string{"add"}, wantErr: "usage"},
		{name: "empty title", args: []string{"add", "   "}, wantErr: "title cannot be empty"},
		{name: "bad priority", args: []string{"add", "Task", "--priority", "urgent"}, wantErr: "invalid priority"},
		{name: "missing priority value", args: []string{"add", "Task", "--priority"}, wantErr: "flag needs an argument"},
		{name: "unknown flag", args: []string{"add", "Task", "--color", "red"}, wantErr: "flag provided but not defined"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := newTestApp(t)

			err := ta.runWithError(t, tt.args...)
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestListCommand(t *testing.T) {
	ta := newTestApp(t)
	ta.run(t, "add", "Learn Golang", "--priority", "high")
	ta.run(t, "add", "Build REST API")
	ta.run(t, "done", "1")

	output := ta.run(t, "list")
	for _, want := range []string{"TASKFLOW CLI", "Learn Golang", "DONE", "HIGH", "Build REST API", "PENDING", "Total: 2 | Completed: 1 | Pending: 1"} {
		if !strings.Contains(output, want) {
			t.Errorf("list output does not contain %q:\n%s", want, output)
		}
	}
}

func TestListCommandFilter(t *testing.T) {
	ta := newTestApp(t)
	ta.run(t, "add", "Learn Golang")
	ta.run(t, "add", "Build REST API")
	ta.run(t, "done", "1")

	doneOutput := ta.run(t, "list", "--status", "done")
	if !strings.Contains(doneOutput, "Learn Golang") || strings.Contains(doneOutput, "Build REST API") {
		t.Errorf("--status done shows wrong tasks:\n%s", doneOutput)
	}

	pendingOutput := ta.run(t, "list", "--status=pending")
	if strings.Contains(pendingOutput, "Learn Golang") || !strings.Contains(pendingOutput, "Build REST API") {
		t.Errorf("--status pending shows wrong tasks:\n%s", pendingOutput)
	}

	if err := ta.runWithError(t, "list", "--status", "finished"); !strings.Contains(err.Error(), "invalid status") {
		t.Errorf("error = %v", err)
	}
}

func TestListEmpty(t *testing.T) {
	ta := newTestApp(t)

	output := ta.run(t, "list")
	if !strings.Contains(output, "No tasks found.") {
		t.Errorf("unexpected output:\n%s", output)
	}
}

func TestShowCommand(t *testing.T) {
	ta := newTestApp(t)
	ta.run(t, "add", "Learn Golang", "--priority", "low")

	output := ta.run(t, "show", "1")
	for _, want := range []string{"TASK #1", "Learn Golang", "PENDING", "LOW", "Created:"} {
		if !strings.Contains(output, want) {
			t.Errorf("show output does not contain %q:\n%s", want, output)
		}
	}
}

func TestDoneAndUndoCommands(t *testing.T) {
	ta := newTestApp(t)
	ta.run(t, "add", "Learn Golang")

	ta.run(t, "done", "1")
	if !ta.savedTasks(t)[0].Completed {
		t.Error("task must be completed after done")
	}

	err := ta.runWithError(t, "done", "1")
	if !strings.Contains(err.Error(), "already completed") {
		t.Errorf("error = %v", err)
	}

	ta.run(t, "undo", "1")
	if ta.savedTasks(t)[0].Completed {
		t.Error("task must be pending after undo")
	}

	err = ta.runWithError(t, "undo", "1")
	if !strings.Contains(err.Error(), "not completed") {
		t.Errorf("error = %v", err)
	}
}

func TestEditCommand(t *testing.T) {
	ta := newTestApp(t)
	ta.run(t, "add", "Learn Golang")

	ta.run(t, "edit", "1", "Learn Go interfaces")
	if title := ta.savedTasks(t)[0].Title; title != "Learn Go interfaces" {
		t.Errorf("title = %q, want %q", title, "Learn Go interfaces")
	}

	ta.runWithError(t, "edit", "1")
	ta.runWithError(t, "edit", "1", " ")
}

func TestDeleteCommand(t *testing.T) {
	ta := newTestApp(t)
	ta.run(t, "add", "First")
	ta.run(t, "add", "Second")

	ta.run(t, "delete", "1")

	tasks := ta.savedTasks(t)
	if len(tasks) != 1 || tasks[0].ID != 2 {
		t.Errorf("remaining tasks = %+v, want only task 2", tasks)
	}
}

func TestSearchCommand(t *testing.T) {
	ta := newTestApp(t)
	ta.run(t, "add", "Learn Golang")
	ta.run(t, "add", "Build REST API")

	output := ta.run(t, "search", "GOLANG")
	if !strings.Contains(output, "Learn Golang") || strings.Contains(output, "Build REST API") {
		t.Errorf("unexpected search output:\n%s", output)
	}
	if !strings.Contains(output, `Found: 1 task(s) matching "GOLANG"`) {
		t.Errorf("search output has no counter:\n%s", output)
	}

	ta.runWithError(t, "search")
}

func TestStatsCommand(t *testing.T) {
	ta := newTestApp(t)
	ta.run(t, "add", "First")
	ta.run(t, "add", "Second")
	ta.run(t, "done", "2")

	output := ta.run(t, "stats")
	for _, want := range []string{"Total tasks:      2", "Completed tasks:  1", "Active tasks:     1", "50.0%"} {
		if !strings.Contains(output, want) {
			t.Errorf("stats output does not contain %q:\n%s", want, output)
		}
	}
}

func TestInvalidIDs(t *testing.T) {
	ta := newTestApp(t)
	ta.run(t, "add", "Task")

	commands := []string{"show", "done", "undo", "delete"}
	badIDs := []string{"abc", "0", "-1", "1.5"}

	for _, command := range commands {
		for _, id := range badIDs {
			err := ta.runWithError(t, command, id)
			if !strings.Contains(err.Error(), "invalid task ID") {
				t.Errorf("%s %s: error = %v", command, id, err)
			}
		}

		err := ta.runWithError(t, command)
		if !strings.Contains(err.Error(), "usage") {
			t.Errorf("%s without ID: error = %v", command, err)
		}

		err = ta.runWithError(t, command, "42")
		if !strings.Contains(err.Error(), "task not found") {
			t.Errorf("%s 42: error = %v", command, err)
		}
	}

	if len(ta.savedTasks(t)) != 1 {
		t.Error("failed commands must not change saved tasks")
	}
}

func TestCorruptedFile(t *testing.T) {
	ta := newTestApp(t)
	corrupted := `[{"id": 1, "title": "Task"`
	if err := os.WriteFile(ta.path, []byte(corrupted), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	err := ta.runWithError(t, "add", "New task")
	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("error = %v", err)
	}

	data, _ := os.ReadFile(ta.path)
	if string(data) != corrupted {
		t.Error("corrupted file must stay unchanged")
	}
}

func TestSeparateFlags(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantPositional []string
		wantFlags      []string
	}{
		{
			name:           "flag after title",
			args:           []string{"Prepare project", "--priority", "high"},
			wantPositional: []string{"Prepare project"},
			wantFlags:      []string{"--priority", "high"},
		},
		{
			name:           "flag before title",
			args:           []string{"--priority", "low", "Read", "docs"},
			wantPositional: []string{"Read", "docs"},
			wantFlags:      []string{"--priority", "low"},
		},
		{
			name:           "flag with equals sign",
			args:           []string{"--status=done"},
			wantPositional: nil,
			wantFlags:      []string{"--status=done"},
		},
		{
			name:           "double dash",
			args:           []string{"--", "-dash title"},
			wantPositional: []string{"-dash title"},
			wantFlags:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			positional, flags := separateFlags(tt.args)
			if !slices.Equal(positional, tt.wantPositional) {
				t.Errorf("positional = %q, want %q", positional, tt.wantPositional)
			}
			if !slices.Equal(flags, tt.wantFlags) {
				t.Errorf("flags = %q, want %q", flags, tt.wantFlags)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("short", 10); got != "short" {
		t.Errorf("truncate(short) = %q", got)
	}
	if got := truncate("Привет, мир!", 8); got != "Приве..." {
		t.Errorf("truncate(cyrillic) = %q", got)
	}
}
