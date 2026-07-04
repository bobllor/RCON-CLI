package app

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/bobllor/rcon/config"
	"github.com/bobllor/rcon/rcon"
	"github.com/spf13/cobra"
)

type RootCommand struct {
	Cmd  *cobra.Command
	Data RootData
	Path AppPath
}

type RootData struct {
	Entry config.RconEntry
}

// NewRootCommand creates a new RootCommand and its initialization flags.
func NewRootCommand(appPaths AppPath) *RootCommand {
	cmd := &RootCommand{
		Cmd: &cobra.Command{
			Use: "mcron",
			Args: func(cmd *cobra.Command, args []string) error {
				if len(args) < 1 {
					PrintFatalString("missing command arguments")
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
}

// InitEntry initializes the RCON entry and validates it. If the data is already
// configured via the config file, then it will instead use the data from the file.
//
// The entry will be mutated in place. If an error occurs, it will return an error.
func (r *RootCommand) initEntry() error {
	// unlike the subcommands add and edit, this will
	// this does not create a config file.
	cfg, cfgErr := config.LoadConfiguration(r.Path.Config)
	// no errors, will fall back to terminal if an error occurs
	if errors.Is(cfgErr, os.ErrNotExist) {
		fmt.Println("mcrcon config not found")
	} else if cfgErr != nil {
		return cfgErr
	} else {
		if cfg.DefaultRcon != "" {
			cfgEntry, ok := cfg.RconEntries[cfg.DefaultRcon]
			if ok {
				r.Data.Entry = cfgEntry
			}
		}
	}

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
