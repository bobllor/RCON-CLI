package run

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bobllor/rcon/app/types"
	"github.com/bobllor/rcon/app/utils"
	"github.com/bobllor/rcon/config"
	"github.com/bobllor/rcon/rcon"
	"github.com/spf13/cobra"
)

type RunCommand struct {
	Cmd  *cobra.Command
	Data RunData
	Path types.AppPath
}

type RunData struct {
	// Target is the target RCON entry to run the command on.
	// This will overwrite the default entry of the config,
	// if there is one.
	Target string
	Entry  config.RconEntry
}

// NewRunCommand creates a new RunCommand for running commands with RCON.
//
// Args represents the command being sent to the server using RCON.
func NewRunCommand(appPaths types.AppPath) *RunCommand {
	cmd := &RunCommand{
		Cmd: &cobra.Command{
			Use:   "run [command] [flags]",
			Short: "Execute a command with RCON",
			Args: func(cmd *cobra.Command, args []string) error {
				if len(args) < 1 {
					return errors.New("must have at least one argument")
				}

				return nil
			},
		},
		Path: appPaths,
	}

	cmd.Cmd.Run = cmd.Run
	cmd.InitFlags()

	return cmd
}

func (rc *RunCommand) Run(cmd *cobra.Command, args []string) {
	cfg, err := config.LoadConfigurationIfMissing(rc.Path.Config)
	if err != nil {
		utils.PrintFatal(err)
	}
	if cfg.DefaultRcon != "" {
		if cfg.EntryExist(cfg.DefaultRcon) {
			rc.Data.Entry = cfg.RconEntries[cfg.DefaultRcon]
		} else {
			fmt.Println("Unable to find default entry")
		}
	} else {
		if rc.Data.Target == "" {
			fmt.Println("No default RCON entry found")
		}
	}

	if rc.Data.Target != "" {
		if cfg.EntryExist(rc.Data.Target) {
			rc.Data.Entry = cfg.RconEntries[rc.Data.Target]
		} else {
			// exit if target doesnt exist, default will not be used as a fallback for run since
			// this is a specific server being ran for.
			utils.PrintFatalString(fmt.Sprintf("RCON entry %s does not exist", rc.Data.Target))
		}
	}

	err = utils.InitEntry(&rc.Data.Entry)
	if err != nil {
		utils.PrintFatal(err)
	}

	err = utils.ValidateEntry(rc.Data.Entry)
	if err != nil {
		utils.PrintFatal(err)
	}

	con, err := rcon.NewRcon(rc.Data.Entry.Address)
	if err != nil {
		utils.PrintFatal(err)
	}

	err = con.Authenticate(rc.Data.Entry.Password)
	if err != nil {
		utils.PrintFatal(err)
	}

	command := strings.Join(args, " ")

	err = con.Command(command)
	if err != nil {
		utils.PrintFatal(err)
	}

	fmt.Println("Command sent")
}

func (rc *RunCommand) InitFlags() {
	rc.Cmd.Flags().StringVarP(&rc.Data.Target, "target", "t", "", "The target RCON entry to send the command")
}
