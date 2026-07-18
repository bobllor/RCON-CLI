package root

import (
	"fmt"

	"github.com/bobllor/rcon-cli/app/utils/paths"
	"github.com/spf13/cobra"
)

// will be updated via go build -ldflags="-X 'root.ProgramVersion=version'"
var ProgramVersion = "v1.0.0"

// RootCommand is the entry point of the program. If no subcommands
// are used, then it will default to running a command to the
// default RCON server.
//
// If no default RCON server is found, it will prompt for inputs
// of the RCON entry to use. This can be bypassed with the -t flag
// if one exists.
type RootCommand struct {
	Cmd  *cobra.Command
	data RootData
	Path paths.AppPath
}

type RootData struct {
	// Version is used as a flag to show the version of the tool.
	Version bool
}

// NewRootCommand creates a new RootCommand and its initialization flags.
func NewRootCommand(appPaths paths.AppPath) *RootCommand {
	cmd := &RootCommand{
		Cmd: &cobra.Command{
			Use:   "gorcon <args>... [flags]",
			Short: "Execute a command with RCON",
		},
		data: RootData{},
		Path: appPaths,
	}

	cmd.Cmd.Run = cmd.Run
	cmd.Cmd.CompletionOptions.DisableDefaultCmd = true

	cmd.RootInitFlags()

	return cmd
}

// Run is the main entry point for the root CMD. This will run
// the execution of the command to the server.
func (r *RootCommand) Run(cmd *cobra.Command, args []string) {
	if r.data.Version {
		fmt.Println(ProgramVersion)
		return
	}
}

func (r *RootCommand) RootInitFlags() {
	r.Cmd.Flags().BoolVarP(&r.data.Version, "version", "v", false, "Displays the version")
}
