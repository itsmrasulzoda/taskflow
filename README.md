# TaskFlow

TaskFlow is a small command-line task manager written in Go.
It runs in a Linux terminal and keeps all tasks in a local JSON file.

The project uses **only the Go standard library** (`flag`, `encoding/json`, `os`, `io`, `time`, `testing`)
and is meant to show clean, readable Go code for a Junior Go Developer.

## Features

- Add tasks with a priority (`low`, `medium`, `high`)
- List all tasks or filter them by status (`done` / `pending`)
- Show details of a single task
- Mark tasks as done and return them to the pending state
- Edit task titles
- Delete tasks
- Case-insensitive search by title
- Statistics: total, completed, active tasks and completion percentage
- Tasks are saved to `data/tasks.json` after every change
- The data file is created automatically if it does not exist
- Clear error messages for wrong input, missing tasks and broken JSON
- Unit tests that never touch your real data

## Requirements

- Linux (any distribution, instructions for Arch Linux are below)
- Go 1.22 or newer
- Git (optional, to clone the repository)

## Installation

### Arch Linux

```bash
# Install Go and Git
sudo pacman -Syu go git

# Check the installation
go version
```

Get the project and build it:

```bash
git clone <repository-url> taskflow
cd taskflow
go build -o taskflow ./cmd/taskflow
./taskflow help
```

To use `taskflow` from any directory, install it into `~/go/bin`:

```bash
go install ./cmd/taskflow

# Add ~/go/bin to PATH (bash). For zsh use ~/.zshrc, for fish see below.
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

For the fish shell:

```fish
fish_add_path ~/go/bin
```

### Other Linux distributions

Install Go from your package manager (for example `sudo apt install golang-go`
on Debian/Ubuntu or `sudo dnf install golang` on Fedora) or from
<https://go.dev/dl/>, then follow the same build steps.

## Building

```bash
# Build the binary in the project root
go build -o taskflow ./cmd/taskflow

# Check the code with the standard tools
gofmt -l .
go vet ./...
```

## Where tasks are stored

By default tasks are stored in `data/tasks.json`. **The path is relative to the
current directory**, so run TaskFlow from the project root, or set the
`TASKFLOW_FILE` environment variable to use one file from anywhere:

```bash
export TASKFLOW_FILE="$HOME/.local/share/taskflow/tasks.json"
```

Add this line to `~/.bashrc` (or `~/.zshrc`) to make it permanent.
Missing directories and the file itself are created automatically.

How the file is handled:

| Situation                         | Behavior                                                 |
|-----------------------------------|----------------------------------------------------------|
| File does not exist               | Created with an empty list `[]`                          |
| File is empty or contains `null`  | Treated as an empty task list                            |
| File contains broken JSON         | Error is shown, the file is **not** changed              |
| Duplicate IDs or invalid fields   | Error is shown, the file is **not** changed              |
| File cannot be read or written    | Error with the file path and the system reason           |

Tasks are written to a temporary file first and then renamed, so a failed
write does not destroy existing data.

Example of `data/tasks.json`:

```json
[
  {
    "id": 1,
    "title": "Learn Golang",
    "completed": true,
    "priority": "high",
    "created_at": "2026-11-16T23:22:10.123456+05:00",
    "updated_at": "2026-11-16T23:25:41.654321+05:00"
  }
]
```

## Usage

```
taskflow <command> [arguments] [flags]
```

| Command                                        | Description                              |
|------------------------------------------------|------------------------------------------|
| `add "<title>" [--priority low\|medium\|high]` | Add a new task (default priority: medium) |
| `list [--status all\|done\|pending]`           | Show tasks (default: all)                |
| `show <id>`                                    | Show task details                        |
| `done <id>`                                    | Mark a task as completed                 |
| `undo <id>`                                    | Mark a completed task as pending again   |
| `edit <id> "<new title>"`                      | Change the title of a task               |
| `delete <id>`                                  | Delete a task                            |
| `search "<query>"`                             | Find tasks by title (case-insensitive)   |
| `stats`                                        | Show task statistics                     |
| `help`                                         | Show help (`-h` and `--help` also work)  |

The examples below assume `taskflow` is in your `PATH`.
If you only built the binary, use `./taskflow` instead.

### Add a task

```bash
taskflow add "Learn Golang"
taskflow add "Prepare project" --priority high
taskflow add --priority low "Read documentation"
taskflow add "Build REST API" --priority=medium
```

```
Task #1 added: Learn Golang (priority: medium)
```

Flags can be placed before or after the title, and priority is case-insensitive
(`HIGH` works too). Quotes are recommended; without them all words are joined:
`taskflow add Build REST API` creates the task "Build REST API".
To create a title that starts with `-`, put `--` before it:
`taskflow add -- "-1 bug left"`.

### List tasks

```bash
taskflow list
taskflow list --status done
taskflow list --status pending
```

### Show a task

```bash
taskflow show 1
```

### Complete a task and undo it

```bash
taskflow done 1
taskflow undo 1
```

```
Task #1 marked as done: Learn Golang
Task #1 marked as pending: Learn Golang
```

### Edit a task

```bash
taskflow edit 1 "Learn Go interfaces"
```

```
Task #1 updated: Learn Go interfaces
```

### Delete a task

```bash
taskflow delete 1
```

```
Task #1 deleted: Learn Go interfaces
```

IDs of deleted tasks are not shifted: other tasks keep their IDs.

### Search

```bash
taskflow search "golang"
```

### Statistics

```bash
taskflow stats
```

### Help

```bash
taskflow help
```

## Output examples

`taskflow list`:

```
==========================================================
                       TASKFLOW CLI
==========================================================

ID   TITLE                             STATUS    PRIORITY
1    Learn Golang                      DONE      HIGH
2    Build REST API                    PENDING   MEDIUM
3    Read documentation                PENDING   LOW

==========================================================
Total: 3 | Completed: 1 | Pending: 2
==========================================================
```

Titles longer than 32 characters are shortened with `...` in the table;
`taskflow show <id>` always prints the full title.

`taskflow show 1`:

```
==========================================================
                         TASK #1
==========================================================

Title:     Learn Golang
Status:    DONE
Priority:  HIGH
Created:   2026-11-16 23:22
Updated:   2026-11-16 23:25

==========================================================
```

`taskflow search "golang"`:

```
==========================================================
                      SEARCH RESULTS
==========================================================

ID   TITLE                             STATUS    PRIORITY
1    Learn Golang                      DONE      HIGH

==========================================================
Found: 1 task(s) matching "golang"
==========================================================
```

`taskflow stats`:

```
==========================================================
                      TASKFLOW STATS
==========================================================

Total tasks:      3
Completed tasks:  1
Active tasks:     2
Completion:       33.3%

Progress: [##########--------------------]

==========================================================
```

### Error examples

Errors are printed to stderr and the program exits with code `1`:

```
$ taskflow
(help text)
Error: no command provided

$ taskflow fly
Error: unknown command "fly", run 'taskflow help' to see available commands

$ taskflow show abc
Error: invalid task ID "abc": ID must be a positive number

$ taskflow done 99
Error: task not found: id 99

$ taskflow add "   "
Error: task title cannot be empty

$ taskflow add "Task" --priority urgent
Error: invalid priority "urgent": use low, medium or high

$ taskflow done 1
Error: task is already completed: id 1

$ taskflow undo 2
Error: task is not completed: id 2

$ taskflow list --status maybe
Error: invalid status "maybe": use all, done or pending

$ taskflow list   # with a broken data/tasks.json
Error: tasks file contains invalid JSON (data/tasks.json): unexpected end of JSON input; fix or remove the file
```

## Project structure

```
taskflow/
├── cmd/
│   └── taskflow/
│       └── main.go          # Entry point: creates the app and prints errors
├── internal/
│   ├── task/
│   │   ├── task.go          # Task model, priorities, statuses, errors, validation
│   │   ├── service.go       # Business logic: add, edit, done, undo, delete, search, filter, stats
│   │   └── service_test.go
│   ├── storage/
│   │   ├── json.go          # Loading and saving tasks in a JSON file
│   │   └── json_test.go
│   └── cli/
│       ├── cli.go           # App, command dispatch, argument and flag parsing
│       ├── commands.go      # One handler per command
│       ├── output.go        # Table, details, stats and help output
│       └── cli_test.go
├── data/
│   └── tasks.json           # Default data file
├── README.md
├── .gitignore
└── go.mod
```

How the parts work together:

1. `main.go` creates `cli.App` with the data file path and calls `Run(os.Args[1:])`.
2. `cli` parses the command, loads tasks with `storage`, and creates a `task.Service`.
3. `task.Service` changes the tasks in memory and returns errors such as `task.ErrTaskNotFound`.
4. `cli` saves the updated list with `storage` and prints the result.

## Running tests

```bash
# Run all tests
go test ./...

# Verbose output
go test -v ./...

# Coverage
go test -cover ./...
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
```

Tests create files only inside temporary directories (`t.TempDir()`),
so your real `data/tasks.json` is never changed.

Covered scenarios:

- adding tasks, priorities and default values;
- deleting, editing, completing and undoing tasks;
- searching and filtering by status;
- statistics;
- saving and loading JSON, creating a missing file, empty files;
- broken JSON, invalid task data, file system errors;
- CLI commands: wrong arguments, invalid and missing IDs, unknown commands.

## License

MIT
