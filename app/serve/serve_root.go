package serve

import (
	ipcstart "github.com/bobllor/rcon/app/serve/start"
	ipcstop "github.com/bobllor/rcon/app/serve/stop"
	"github.com/bobllor/rcon/app/types"
	"github.com/spf13/cobra"
)

// ServeCommand is a struct used to group subcommands for
// managing the IPC listener and connection.
//
// It's only purpose is to group commands together.
type ServeCommand struct {
	Cmd *cobra.Command
}

func NewServeCommand(addr string, paths types.AppPath) *ServeCommand {
	cmd := &ServeCommand{
		Cmd: &cobra.Command{
			Use:   "serve [command]",
			Short: "Manage the IPC RCON service",
			Args:  cobra.NoArgs,
		},
	}

	startCmd := ipcstart.NewIpcStartCommand(addr, paths)
	stopCmd := ipcstop.NewStopCommand(addr, paths)

	cmd.Cmd.AddCommand(startCmd.Cmd)
	cmd.Cmd.AddCommand(stopCmd.Cmd)

	return cmd
}
