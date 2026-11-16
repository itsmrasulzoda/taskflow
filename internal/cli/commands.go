package cli

import (
	"fmt"
	"strings"

	"taskflow/internal/task"
)

// taskAction is an operation on a single task, for example Complete or Delete.
type taskAction func(service *task.Service, id int) (task.Task, error)

// runAdd handles: taskflow add "Title" [--priority low|medium|high]
func (a *App) runAdd(args []string) error {
	const usage = `add "<title>" [--priority low|medium|high]`

	positional, flagArgs := separateFlags(args)
	flags := newFlagSet("add")
	priority := flags.String("priority", task.PriorityMedium, "task priority")
	if err := flags.Parse(flagArgs); err != nil {
		return fmt.Errorf("%v, usage: taskflow %s", err, usage)
	}
	if len(positional) == 0 {
		return usageError(usage)
	}

	service, err := a.loadService()
	if err != nil {
		return err
	}

	title := strings.Join(positional, " ")
	newTask, err := service.Add(title, *priority)
	if err != nil {
		return err
	}

	if err := a.saveService(service); err != nil {
		return err
	}

	fmt.Fprintf(a.out, "Task #%d added: %s (priority: %s)\n", newTask.ID, newTask.Title, newTask.Priority)
	return nil
}

// runList handles: taskflow list [--status all|done|pending]
func (a *App) runList(args []string) error {
	const usage = "list [--status all|done|pending]"

	positional, flagArgs := separateFlags(args)
	flags := newFlagSet("list")
	status := flags.String("status", task.StatusAll, "filter by status")
	if err := flags.Parse(flagArgs); err != nil {
		return fmt.Errorf("%v, usage: taskflow %s", err, usage)
	}
	if len(positional) > 0 {
		return usageError(usage)
	}

	service, err := a.loadService()
	if err != nil {
		return err
	}

	tasks, err := service.Filter(*status)
	if err != nil {
		return err
	}

	a.printHeader("TASKFLOW CLI")
	a.printTaskTable(tasks)
	a.printSummary(service.Stats())
	return nil
}

// runShow handles: taskflow show <id>
func (a *App) runShow(args []string) error {
	id, err := parseSingleID(args, "show <id>")
	if err != nil {
		return err
	}

	service, err := a.loadService()
	if err != nil {
		return err
	}

	foundTask, err := service.Get(id)
	if err != nil {
		return err
	}

	a.printTaskDetails(foundTask)
	return nil
}

// runDone handles: taskflow done <id>
func (a *App) runDone(args []string) error {
	return a.runTaskAction(args, "done <id>", (*task.Service).Complete, "marked as done")
}

// runUndo handles: taskflow undo <id>
func (a *App) runUndo(args []string) error {
	return a.runTaskAction(args, "undo <id>", (*task.Service).Undo, "marked as pending")
}

// runDelete handles: taskflow delete <id>
func (a *App) runDelete(args []string) error {
	return a.runTaskAction(args, "delete <id>", (*task.Service).Delete, "deleted")
}

// runTaskAction contains the shared steps for done, undo and delete:
// parse the ID, load tasks, apply the action, save and print the result.
func (a *App) runTaskAction(args []string, usage string, action taskAction, message string) error {
	id, err := parseSingleID(args, usage)
	if err != nil {
		return err
	}

	service, err := a.loadService()
	if err != nil {
		return err
	}

	changedTask, err := action(service, id)
	if err != nil {
		return err
	}

	if err := a.saveService(service); err != nil {
		return err
	}

	fmt.Fprintf(a.out, "Task #%d %s: %s\n", changedTask.ID, message, changedTask.Title)
	return nil
}

// runEdit handles: taskflow edit <id> "New title"
func (a *App) runEdit(args []string) error {
	const usage = `edit <id> "<new title>"`

	if len(args) < 2 {
		return usageError(usage)
	}

	id, err := parseID(args[0])
	if err != nil {
		return err
	}

	service, err := a.loadService()
	if err != nil {
		return err
	}

	newTitle := strings.Join(args[1:], " ")
	editedTask, err := service.Edit(id, newTitle)
	if err != nil {
		return err
	}

	if err := a.saveService(service); err != nil {
		return err
	}

	fmt.Fprintf(a.out, "Task #%d updated: %s\n", editedTask.ID, editedTask.Title)
	return nil
}

// runSearch handles: taskflow search "query"
func (a *App) runSearch(args []string) error {
	if len(args) == 0 {
		return usageError(`search "<query>"`)
	}

	service, err := a.loadService()
	if err != nil {
		return err
	}

	query := strings.Join(args, " ")
	found, err := service.Search(query)
	if err != nil {
		return err
	}

	a.printHeader("SEARCH RESULTS")
	a.printTaskTable(found)
	a.printSeparator()
	fmt.Fprintf(a.out, "Found: %d task(s) matching %q\n", len(found), strings.TrimSpace(query))
	a.printSeparator()
	return nil
}

// runStats handles: taskflow stats
func (a *App) runStats(args []string) error {
	if len(args) > 0 {
		return usageError("stats")
	}

	service, err := a.loadService()
	if err != nil {
		return err
	}

	a.printStats(service.Stats())
	return nil
}
