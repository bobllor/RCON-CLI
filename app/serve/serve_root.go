package serve

import (
	"fmt"
	"os"

	ipcexec "github.com/bobllor/rcon/app/serve/exec"
	"github.com/bobllor/rcon/app/serve/internal"
	ipcstart "github.com/bobllor/rcon/app/serve/start"
	ipcstop "github.com/bobllor/rcon/app/serve/stop"
	"github.com/bobllor/rcon/app/utils/paths"
	"github.com/spf13/cobra"
)

// ServeCommand is a struct used to group subcommands for
// managing the IPC listener and connection.
//
// Certain flags exist to perform other file operations.
type ServeCommand struct {
	Cmd  *cobra.Command
	Path paths.AppPath
	data ServeData
}

type ServeData struct {
	// Clean is used to cleanup leftover files created by the
	// RCON process.
	Clean bool
	// Address is the socket address.
	Address string
	// PidFile is the path to the PID file of the service.
	PidFile string
}

func NewServeCommand(addr, pidFile string, paths paths.AppPath) *ServeCommand {
	cmd := &ServeCommand{
		Cmd: &cobra.Command{
			Use:   "serve [command]",
			Short: "Manage the IPC RCON service",
			Args:  cobra.NoArgs,
		},
		Path: paths,
		data: ServeData{
			Address: addr,
			PidFile: pidFile,
		},
	}

	startCmd := ipcstart.NewIpcStartCommand(addr, pidFile, paths)
	stopCmd := ipcstop.NewStopCommand(addr, pidFile, paths)
	execCmd := ipcexec.NewExecCommand(addr)

	cmd.Cmd.AddCommand(startCmd.Cmd)
	cmd.Cmd.AddCommand(stopCmd.Cmd)
	cmd.Cmd.AddCommand(execCmd.Cmd)

	cmd.Cmd.Run = cmd.Run
	cmd.initFlags()

	return cmd
}

func (sc *ServeCommand) Run(cmd *cobra.Command, args []string) {
	// alternative way to clean/remove the files in case the original code fails to
	if sc.data.Clean {
		errs := internal.RemoveFiles(sc.data.PidFile, sc.data.Address)
		for _, err := range errs {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func (sc *ServeCommand) initFlags() {
	sc.Cmd.Flags().BoolVar(&sc.data.Clean, "clean", false, "Remove leftover files during the service startup")
}
