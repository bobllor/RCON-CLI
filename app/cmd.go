package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bobllor/rcon/app/add"
	"github.com/bobllor/rcon/app/edit"
	"github.com/bobllor/rcon/app/list"
	"github.com/bobllor/rcon/app/remove"
	"github.com/bobllor/rcon/app/root"
	"github.com/bobllor/rcon/app/run"
	"github.com/bobllor/rcon/app/types"
	"github.com/bobllor/rcon/app/utils"
)

// Execute is the main entry point of the root and its children. It will create a new
// App struct and use it for the rest of the info.
func Execute() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	configPath := filepath.Join(home, ".config", "mcrcon")

	paths := types.AppPath{
		Home:   home,
		Config: configPath,
	}

	mkErr := utils.MkdirAll(configPath)
	// errors will not exit
	if mkErr != nil {
		fmt.Fprintf(os.Stderr, "failed to make files: %v\n", mkErr)
	}

	rootCmd := root.NewRootCommand(paths)

	addCmd := add.NewAddCommand(paths)
	listCmd := list.NewListCommand(paths)
	rmCmd := remove.NewRemoveCommand(paths)
	editCmd := edit.NewEditCommand(paths)
	runCmd := run.NewRunCommand(paths)

	rootCmd.Cmd.AddCommand(addCmd.Cmd)
	rootCmd.Cmd.AddCommand(listCmd.Cmd)
	rootCmd.Cmd.AddCommand(rmCmd.Cmd)
	rootCmd.Cmd.AddCommand(editCmd.Cmd)
	rootCmd.Cmd.AddCommand(runCmd.Cmd)

	// errors are handled with PrintFatal in the commands
	rootCmd.Cmd.Execute()
}
