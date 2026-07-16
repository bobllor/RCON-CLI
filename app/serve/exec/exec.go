package ipcexec

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/bobllor/rcon-cli/app/utils"
	"github.com/spf13/cobra"
)

type ExecCommand struct {
	Cmd  *cobra.Command
	data ExecData
}

type ExecData struct {
	Address string
}

func NewExecCommand(addr string) *ExecCommand {
	cmd := &ExecCommand{
		Cmd: &cobra.Command{
			Use:   "exec <command>...",
			Short: "Execute a command on a running RCON service",
		},
		data: ExecData{
			Address: addr,
		},
	}

	cmd.Cmd.Run = cmd.Run

	return cmd
}

func (ec *ExecCommand) Run(cmd *cobra.Command, args []string) {
	con, err := net.Dial("unix", ec.data.Address)
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ECONNREFUSED) {
		utils.PrintFatalString("RCON service is not running")
	} else if err != nil {
		utils.PrintFatal(err)
	}
	defer con.Close()

	command := strings.Join(args, " ")

	_, err = con.Write([]byte(command))
	if err != nil {
		utils.PrintFatal(err)
	}

	con.SetReadDeadline(time.Now().Add(7 * time.Second))

	b, err := io.ReadAll(con)
	if errors.Is(err, os.ErrDeadlineExceeded) {
		utils.PrintFatalString("Timed out waiting for command response (7s)")
	}
	if err != nil {
		utils.PrintFatal(err)
	}

	commandRes := strings.TrimSpace(string(b))
	if commandRes != "" {
		fmt.Println(string(b))
	}
}
