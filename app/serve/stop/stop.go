package ipcstop

import (
	"errors"
	"fmt"
	"os"

	"github.com/bobllor/rcon/app/serve/internal"
	"github.com/bobllor/rcon/app/utils"
	"github.com/bobllor/rcon/app/utils/paths"
	"github.com/spf13/cobra"
)

type StopCommand struct {
	Cmd  *cobra.Command
	Path paths.AppPath
	data StopData
}

type StopData struct {
	Address string
	PidFile string
}

func NewStopCommand(addr, pidFile string, appPaths paths.AppPath) *StopCommand {
	cmd := StopCommand{
		Cmd: &cobra.Command{
			Use:   "stop",
			Short: "Stop the RCON service",
		},
		Path: appPaths,
		data: StopData{
			Address: addr,
			PidFile: pidFile,
		},
	}

	cmd.Cmd.Run = cmd.Run

	return &cmd
}

func (sc *StopCommand) Run(cmd *cobra.Command, args []string) {
	pid, err := internal.ReadPID(sc.data.PidFile)
	if errors.Is(err, os.ErrNotExist) {
		utils.PrintFatalString("RCON service is not running")
	}
	if err != nil {
		utils.PrintFatal(err)
	}

	isRunning := internal.CheckProcessRunning(pid)
	if !isRunning {
		utils.PrintFatalString("RCON service is not running")
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		utils.PrintFatal(err)
	}

	err = p.Kill()
	if err != nil {
		utils.PrintFatal(err)
	}

	fmt.Printf("Stopped RCON service (%d)\n", pid)
	// errors do not matter. this cleanup can only occur at the end
	internal.RemoveFiles(sc.data.PidFile, sc.data.Address)
}
