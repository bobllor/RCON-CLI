package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bobllor/rcon-cli/app/add"
	"github.com/bobllor/rcon-cli/app/edit"
	rconexec "github.com/bobllor/rcon-cli/app/exec"
	"github.com/bobllor/rcon-cli/app/list"
	"github.com/bobllor/rcon-cli/app/remove"
	"github.com/bobllor/rcon-cli/app/root"
	"github.com/bobllor/rcon-cli/app/serve"
	"github.com/bobllor/rcon-cli/app/utils/files"
	"github.com/bobllor/rcon-cli/app/utils/paths"
)

// Execute is the main entry point of the root and its children. It will create a new
// App struct and use it for the rest of the command.
func Execute() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	configPath := filepath.Join(home, paths.ConfigPathRel)
	runtimePath := filepath.Join(home, paths.RuntimePathRel)
	logPath := filepath.Join(home, paths.LogPathRel)

	appPaths := paths.AppPath{
		Home:    home,
		Config:  configPath,
		Runtime: runtimePath,
		Log:     logPath,
	}

	errs := appPaths.MkdirAll()
	for _, err := range errs {
		fmt.Fprintln(os.Stderr, err)
	}

	rootCmd := root.NewRootCommand(appPaths)

	execCmd := rconexec.NewExecCommand(appPaths)
	addCmd := add.NewAddCommand(appPaths)
	listCmd := list.NewListCommand(appPaths)
	rmCmd := remove.NewRemoveCommand(appPaths)
	editCmd := edit.NewEditCommand(appPaths)

	socketAddr := filepath.Join(appPaths.Runtime, files.SocketFile)
	pidPath := filepath.Join(appPaths.Runtime, files.PidFile)

	serveCmd := serve.NewServeCommand(socketAddr, pidPath, appPaths)

	rootCmd.Cmd.AddCommand(execCmd.Cmd)
	rootCmd.Cmd.AddCommand(addCmd.Cmd)
	rootCmd.Cmd.AddCommand(listCmd.Cmd)
	rootCmd.Cmd.AddCommand(rmCmd.Cmd)
	rootCmd.Cmd.AddCommand(editCmd.Cmd)
	rootCmd.Cmd.AddCommand(serveCmd.Cmd)

	// errors are handled with PrintFatal in the commands
	rootCmd.Cmd.Execute()
}
