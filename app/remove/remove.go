package remove

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bobllor/rcon-cli/app/utils"
	"github.com/bobllor/rcon-cli/app/utils/paths"
	"github.com/bobllor/rcon-cli/config"
	"github.com/spf13/cobra"
)

// RemoveCommand is the subcommand that handles removing RCON entries
// from the configuration.
type RemoveCommand struct {
	Cmd  *cobra.Command
	Path paths.AppPath
}

func NewRemoveCommand(paths paths.AppPath) *RemoveCommand {
	cmd := &RemoveCommand{
		Cmd: &cobra.Command{
			Use:     "remove <entry>...",
			Short:   "Remove an RCON entry",
			Aliases: []string{"rm"},
		},
		Path: paths,
	}

	cmd.Cmd.Run = cmd.Run
	cmd.Cmd.PreRunE = cmd.PreRunE

	return cmd
}

// Run is the main entrypoint to RemoveCommand. It uses args as a list of entries
// to remove from the configuration file.
//
// If there has been at least one valid entry that has been removed, it will write
// to the configuration file.
func (r *RemoveCommand) Run(cmd *cobra.Command, args []string) {
	cfg, err := config.LoadConfigurationIfMissing(r.Path.Config)
	if err != nil {
		utils.PrintFatal(err)
	}
	removedEntries, hasDeleted := r.remove(cfg, args)

	if hasDeleted {
		err := cfg.WriteFile(r.Path.Config)
		if err != nil {
			utils.PrintFatal(err)
		}
	}

	fmt.Println(strings.Join(removedEntries, "\n"))
}

func (r *RemoveCommand) PreRunE(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("must have at least one argument")
	}

	return nil
}

// remove removes the entries from the configuration. It will return a list of
// entries that have been removed and true if at least one entry has been
// deleted.
func (r *RemoveCommand) remove(cfg *config.Configuration, entries []string) ([]string, bool) {
	hasDeleted := false

	removed := []string{}
	for _, target := range entries {
		if cfg.DeleteEntry(target) {
			removed = append(removed, fmt.Sprintf("Removed entry %s", target))
			if !hasDeleted {
				hasDeleted = true
			}
		} else {
			removed = append(removed, fmt.Sprintf("Entry %s does not exist", target))
		}
	}

	return removed, hasDeleted
}
