package remove

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bobllor/rcon/app/types"
	"github.com/bobllor/rcon/app/utils"
	"github.com/bobllor/rcon/config"
	"github.com/spf13/cobra"
)

// RemoveCommand is the subcommand that handles removing RCON entries
// from the configuration.
type RemoveCommand struct {
	Cmd  *cobra.Command
	Path types.AppPath
}

func NewRemoveCommand(paths types.AppPath) *RemoveCommand {
	cmd := &RemoveCommand{
		Cmd: &cobra.Command{
			Use:   "rm [entry...]",
			Short: "Remove a RCON entry",
		},
		Path: paths,
	}

	cmd.Cmd.Run = cmd.Run
	cmd.Cmd.PreRunE = cmd.PreRunE

	return cmd
}

func (dc *RemoveCommand) Run(cmd *cobra.Command, args []string) {
	cfg, err := config.LoadConfigurationIfMissing(dc.Path.Config)
	if err != nil {
		utils.PrintFatal(err)
	}
	var hasDeleted bool

	filesRemoved := []string{}
	for _, target := range args {
		if cfg.DeleteEntry(target) {
			filesRemoved = append(filesRemoved, fmt.Sprintf("Removed entry %s", target))
			if !hasDeleted {
				hasDeleted = true
			}
		} else {
			filesRemoved = append(filesRemoved, fmt.Sprintf("Entry %s does not exist", target))
		}
	}

	if hasDeleted {
		err := cfg.WriteFile(dc.Path.Config)
		if err != nil {
			utils.PrintFatal(err)
		}
	}

	fmt.Println(strings.Join(filesRemoved, "\n"))
}

func (dc *RemoveCommand) PreRunE(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("must have at least one argument")
	}

	return nil
}
