// Command taskflow is a simple CLI task manager that stores tasks in a JSON file.
package main

import (
	"fmt"
	"os"

	"taskflow/internal/cli"
)

func main() {
	app := cli.NewApp(cli.DataFilePath(), os.Stdout)

	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
