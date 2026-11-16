// Package cli parses command line arguments and runs TaskFlow commands.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"taskflow/internal/storage"
	"taskflow/internal/task"
)

const (
	// DefaultDataFile is used when the TASKFLOW_FILE variable is not set.
	DefaultDataFile = "data/tasks.json"

	// DataFileEnv is the environment variable that overrides the data file path.
	DataFileEnv = "TASKFLOW_FILE"
)

// App connects the command line with the task service and the storage.
type App struct {
	storage *storage.JSONStorage
	out     io.Writer
}

// NewApp creates an application that keeps tasks in dataFile
// and prints results to out.
func NewApp(dataFile string, out io.Writer) *App {
	return &App{
		storage: storage.NewJSONStorage(dataFile),
		out:     out,
	}
}

// DataFilePath returns the path of the tasks file.
// The TASKFLOW_FILE environment variable takes priority over the default path.
func DataFilePath() string {
	path := strings.TrimSpace(os.Getenv(DataFileEnv))
	if path != "" {
		return path
	}
	return DefaultDataFile
}

// Run executes the command described by args (without the program name).
func (a *App) Run(args []string) error {
	if len(args) == 0 {
		a.printUsage()
		return errors.New("no command provided")
	}

	command := args[0]
	commandArgs := args[1:]

	switch command {
	case "add":
		return a.runAdd(commandArgs)
	case "list":
		return a.runList(commandArgs)
	case "show":
		return a.runShow(commandArgs)
	case "done":
		return a.runDone(commandArgs)
	case "undo":
		return a.runUndo(commandArgs)
	case "delete":
		return a.runDelete(commandArgs)
	case "edit":
		return a.runEdit(commandArgs)
	case "search":
		return a.runSearch(commandArgs)
	case "stats":
		return a.runStats(commandArgs)
	case "help", "-h", "--help":
		a.printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q, run 'taskflow help' to see available commands", command)
	}
}

// loadService reads tasks from the storage and wraps them in a service.
func (a *App) loadService() (*task.Service, error) {
	tasks, err := a.storage.Load()
	if err != nil {
		return nil, err
	}
	return task.NewService(tasks), nil
}

// saveService writes the current state of the service to the storage.
func (a *App) saveService(service *task.Service) error {
	if err := a.storage.Save(service.Tasks()); err != nil {
		return fmt.Errorf("save tasks: %w", err)
	}
	return nil
}

// parseID converts a command line value into a task ID.
func parseID(value string) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid task ID %q: ID must be a positive number", value)
	}
	return id, nil
}

// parseSingleID checks that exactly one argument was passed and parses it as an ID.
func parseSingleID(args []string, usage string) (int, error) {
	if len(args) != 1 {
		return 0, usageError(usage)
	}
	return parseID(args[0])
}

// usageError builds an error that shows the correct command syntax.
func usageError(usage string) error {
	return fmt.Errorf("wrong arguments, usage: taskflow %s", usage)
}

// newFlagSet creates a flag set that returns errors instead of exiting
// and does not print anything by itself.
func newFlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	return flags
}

// separateFlags splits arguments into positional values and flags.
//
// The standard flag package stops parsing at the first positional argument,
// so `add "Title" --priority high` would not work without this step.
// Every flag in TaskFlow takes a value, so the argument after a flag
// (for example "high" after "--priority") belongs to that flag.
// Everything after "--" is treated as positional.
func separateFlags(args []string) (positional []string, flags []string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}

		if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			flags = append(flags, arg)
			hasValueInside := strings.Contains(arg, "=")
			if !hasValueInside && i+1 < len(args) {
				flags = append(flags, args[i+1])
				i++
			}
			continue
		}

		positional = append(positional, arg)
	}

	return positional, flags
}
