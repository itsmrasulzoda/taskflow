package cli

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"taskflow/internal/task"
)

const (
	lineWidth        = 58
	titleColumnWidth = 32
	progressBarWidth = 30
	timeLayout       = "2006-01-02 15:04"

	// tableRowFormat gives the columns ID, TITLE, STATUS and PRIORITY.
	// The title column is two characters wider than titleColumnWidth
	// to keep a gap before the STATUS column.
	tableRowFormat = "%-5v%-34v%-10v%v\n"
)

// printSeparator prints a line of "=" characters.
func (a *App) printSeparator() {
	fmt.Fprintln(a.out, strings.Repeat("=", lineWidth))
}

// printHeader prints a framed title in the middle of the line.
func (a *App) printHeader(title string) {
	a.printSeparator()
	fmt.Fprintln(a.out, centerText(title, lineWidth))
	a.printSeparator()
	fmt.Fprintln(a.out)
}

// printTaskTable prints tasks as a table, or a message if there are none.
func (a *App) printTaskTable(tasks []task.Task) {
	if len(tasks) == 0 {
		fmt.Fprintln(a.out, "No tasks found.")
		fmt.Fprintln(a.out)
		return
	}

	fmt.Fprintf(a.out, tableRowFormat, "ID", "TITLE", "STATUS", "PRIORITY")
	for _, t := range tasks {
		fmt.Fprintf(a.out, tableRowFormat,
			t.ID,
			truncate(t.Title, titleColumnWidth),
			statusLabel(t),
			strings.ToUpper(t.Priority),
		)
	}
	fmt.Fprintln(a.out)
}

// printSummary prints the total, completed and pending counters.
func (a *App) printSummary(stats task.Stats) {
	a.printSeparator()
	fmt.Fprintf(a.out, "Total: %d | Completed: %d | Pending: %d\n",
		stats.Total, stats.Completed, stats.Pending)
	a.printSeparator()
}

// printTaskDetails prints all fields of a single task.
func (a *App) printTaskDetails(t task.Task) {
	a.printHeader(fmt.Sprintf("TASK #%d", t.ID))
	fmt.Fprintf(a.out, "%-10s %s\n", "Title:", t.Title)
	fmt.Fprintf(a.out, "%-10s %s\n", "Status:", statusLabel(t))
	fmt.Fprintf(a.out, "%-10s %s\n", "Priority:", strings.ToUpper(t.Priority))
	fmt.Fprintf(a.out, "%-10s %s\n", "Created:", t.CreatedAt.Local().Format(timeLayout))
	fmt.Fprintf(a.out, "%-10s %s\n", "Updated:", t.UpdatedAt.Local().Format(timeLayout))
	fmt.Fprintln(a.out)
	a.printSeparator()
}

// printStats prints statistics with a simple progress bar.
func (a *App) printStats(stats task.Stats) {
	percent := stats.CompletionPercent()

	a.printHeader("TASKFLOW STATS")
	fmt.Fprintf(a.out, "%-17s %d\n", "Total tasks:", stats.Total)
	fmt.Fprintf(a.out, "%-17s %d\n", "Completed tasks:", stats.Completed)
	fmt.Fprintf(a.out, "%-17s %d\n", "Active tasks:", stats.Pending)
	fmt.Fprintf(a.out, "%-17s %.1f%%\n", "Completion:", percent)
	fmt.Fprintln(a.out)
	fmt.Fprintf(a.out, "Progress: %s\n", progressBar(percent, progressBarWidth))
	fmt.Fprintln(a.out)
	a.printSeparator()
}

// printUsage prints the help message.
func (a *App) printUsage() {
	fmt.Fprintf(a.out, `TaskFlow - a simple task manager for your terminal

Usage:
  taskflow <command> [arguments] [flags]

Commands:
  add "<title>" [--priority low|medium|high]   Add a new task (default priority: medium)
  list [--status all|done|pending]             Show tasks (default: all)
  show <id>                                    Show task details
  done <id>                                    Mark a task as completed
  undo <id>                                    Mark a completed task as pending again
  edit <id> "<new title>"                      Change the title of a task
  delete <id>                                  Delete a task
  search "<query>"                             Find tasks by title (case-insensitive)
  stats                                        Show task statistics
  help                                         Show this help

Examples:
  taskflow add "Learn Golang"
  taskflow add "Prepare project" --priority high
  taskflow list --status pending
  taskflow done 1
  taskflow edit 1 "Learn Go interfaces"
  taskflow search "golang"

Data file: %s
Set the %s environment variable to use another file.
`, a.storage.Path(), DataFileEnv)
}

// statusLabel returns "DONE" or "PENDING" for the table.
func statusLabel(t task.Task) string {
	return strings.ToUpper(t.Status())
}

// centerText adds spaces before text so that it is centered within width.
func centerText(text string, width int) string {
	padding := (width - utf8.RuneCountInString(text)) / 2
	if padding <= 0 {
		return text
	}
	return strings.Repeat(" ", padding) + text
}

// truncate shortens text longer than maxLength and adds "..." at the end.
// It works with runes, so non-Latin titles are cut correctly.
func truncate(text string, maxLength int) string {
	runes := []rune(text)
	if len(runes) <= maxLength {
		return text
	}
	return string(runes[:maxLength-3]) + "..."
}

// progressBar draws a bar like [#########---------] for the given percent.
func progressBar(percent float64, width int) string {
	filled := int(math.Round(percent / 100 * float64(width)))
	if filled > width {
		filled = width
	}
	return "[" + strings.Repeat("#", filled) + strings.Repeat("-", width-filled) + "]"
}
