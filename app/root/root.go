package root

import (
	"github.com/bobllor/rcon-cli/app/utils/paths"
	"github.com/spf13/cobra"
)

// will be updated via go build -ldflags="-X 'root.ProgramVersion=version'"
var ProgramVersion = "N/A"

// RootCommand is the entry point of the program. If no subcommands
// are used, then it will default to running a command to the
// default RCON server.
//
// If no default RCON server is found, it will prompt for inputs
// of the RCON entry to use. This can be bypassed with the -t flag
// if one exists.
type RootCommand struct {
	Cmd  *cobra.Command
	Path paths.AppPath
}

// NewRootCommand creates a new RootCommand and its initialization flags.
func NewRootCommand(appPaths paths.AppPath) *RootCommand {
	cmd := &RootCommand{
		Cmd: &cobra.Command{
			Use:   "gorcon",
			Short: "Execute a command with RCON",
		},
		Path: appPaths,
	}

	cmd.Cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.Cmd.Version = ProgramVersion

	return cmd
}
