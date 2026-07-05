package app

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bobllor/rcon/config"
	"github.com/bobllor/rcon/rcon"
	"github.com/spf13/cobra"
)

// RootCommand is the entry point of the program. If no subcommands
// are used, then it will default to running a command to the
// default RCON server.
//
// If no default RCON server is found, it will prompt for inputs
// of the RCON entry to use. This can be bypassed with the -t flag
// if one exists.
type RootCommand struct {
	Cmd  *cobra.Command
	Data RootData
	Path AppPath
}

type RootData struct {
	Entry config.RconEntry
	// Target is the target RCON entry to send the command to. This
	// will overwrite the default RCON entry.
	Target string
}

// NewRootCommand creates a new RootCommand and its initialization flags.
func NewRootCommand(appPaths AppPath) *RootCommand {
	cmd := &RootCommand{
		Cmd: &cobra.Command{
			Use: "mcron",
			// required due to subcommands changing arg parsing rules
			Args: func(cmd *cobra.Command, args []string) error {
				if len(args) < 1 {
					return errors.New("must have at least one argument")
				}

				return nil
			},
		},
		Data: RootData{},
		Path: appPaths,
	}

	cmd.Cmd.Run = cmd.RootRun
	cmd.RootInitFlags()

	return cmd
}

// RootRun is the main entry point for the root CMD. This will run
// the execution of the command to the server.
func (r *RootCommand) RootRun(cmd *cobra.Command, args []string) {
	// this does not create a cfg file.
	cfg, cfgErr := loadConfiguration(r.Path.Config)
	if cfgErr != nil {
		PrintFatal(cfgErr)
	} else {
		if cfg.DefaultRcon != "" {
			cfgEntry, ok := cfg.RconEntries[cfg.DefaultRcon]
			if ok {
				r.Data.Entry = cfgEntry
			}
		}
	}

	// target will overwrite the default
	if r.Data.Target != "" {
		cfgEntry, ok := cfg.RconEntries[r.Data.Target]
		if ok {
			r.Data.Entry = cfgEntry
		}
	}

	initErr := r.initEntry()
	if initErr != nil {
		PrintFatal(initErr)
	}

	con, err := rcon.NewRcon(r.Data.Entry.Address)
	if err != nil {
		PrintFatal(err)
	}

	loginErr := con.Authenticate(r.Data.Entry.Password)
	if loginErr != nil {
		PrintFatal(loginErr)
	}

	command := strings.Join(args, " ")
	cmdErr := con.Command(command)
	if cmdErr != nil {
		PrintFatal(cmdErr)
	}

	fmt.Println("Executed command")
}

func (r *RootCommand) RootInitFlags() {
	r.Cmd.Flags().StringVarP(&r.Data.Entry.Address, "address", "a", "", "RCON address")
	r.Cmd.Flags().StringVarP(&r.Data.Entry.Password, "password", "p", "", "RCON password")

	r.Cmd.Flags().StringVarP(&r.Data.Target, "target", "t", "", "RCON entry target to run the command on")
}

// InitEntry initializes the RCON entry and validates it. If the data is already
// configured via the config file, then it will instead use the data from the file.
//
// The entry will be mutated in place. If an error occurs, it will return an error.
func (r *RootCommand) initEntry() error {
	err := initEntry(&r.Data.Entry)
	if err != nil {
		return err
	}

	validErr := validateEntry(r.Data.Entry)
	if validErr != nil {
		return validErr
	}

	return nil
}
