package app

import (
	"fmt"
	"strings"

	"github.com/bobllor/rcon/config"
	"github.com/spf13/cobra"
)

type AddCommand struct {
	Cmd  *cobra.Command
	Path AppPath
	Data AddData
}

type AddData struct {
	// Name is the RCON name entry. This must be unique, but can be overwritten
	// with a flag.
	Name string
	// Entry is the RCON entry data used to connect to and communicate to RCON.
	Entry config.RconEntry
	// Overwrite is used to overwrite an existing RCON entry in the
	// configuration file. If an entry does not exist, then this will do
	// nothing.
	Overwrite bool
	// SetDefault is used to set the new RCON entry as the default RCON entry.
	// This is primarily used with the root command, where it will instead use
	// the existing entry by default.
	SetDefault bool
}

func NewAddCommand(appPaths AppPath) *AddCommand {
	cmd := &AddCommand{
		Cmd: &cobra.Command{
			Use:   "add [entry] [flags]",
			Short: "Add a new RCON entry",
			Long:  "Add a new RCON entry into the configuration",
		},
		Path: appPaths,
	}

	cmd.Cmd.Run = cmd.AddRun
	cmd.AddInitFlags()

	return cmd
}

func (ac *AddCommand) AddRun(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		if strings.TrimSpace(ac.Data.Name) != "" {
			PrintFatalString("cannot have entry args and use the -n flag")
		}
		ac.Data.Name = strings.Join(args, " ")
	}
	err := ac.initData()
	if err != nil {
		PrintFatal(err)
	}

	cfg, err := config.LoadConfigurationIfMissing(ac.Path.Config)
	if err != nil {
		PrintFatal(err)
	}

	if !ac.Data.Overwrite {
		exist := cfg.EntryExist(ac.Data.Name)
		if exist {
			PrintFatal(fmt.Errorf(`RCON entry "%s" already exists`, ac.Data.Name))
		}
	}
	if ac.Data.SetDefault {
		cfg.DefaultRcon = ac.Data.Name
	}

	cfg.AddEntry(ac.Data.Name, ac.Data.Entry)
	writeErr := cfg.WriteFile(ac.Path.Config)
	if writeErr != nil {
		PrintFatal(writeErr)
	}

	// below are only used for stdout to the end user
	if ac.Data.Overwrite {
		fmt.Printf(`Replaced existing RCON entry %s%s`, ac.Data.Name, "\n")
	} else {
		fmt.Printf(`Added new RCON entry %s%s`, ac.Data.Name, "\n")
	}
	if ac.Data.SetDefault {
		fmt.Printf(`Set RCON entry %s as the default entry%s`, ac.Data.Name, "\n")
	}
}

func (ac *AddCommand) AddInitFlags() {
	ac.Cmd.Flags().StringVarP(&ac.Data.Name, "name", "n", "", "The unqiue name of the RCON entry")
	ac.Cmd.Flags().StringVarP(&ac.Data.Entry.Address, "address", "a", "", "The address of the RCON entry")
	ac.Cmd.Flags().StringVarP(&ac.Data.Entry.Password, "password", "p", "", "The password of the RCON entry")

	ac.Cmd.Flags().BoolVar(&ac.Data.Overwrite, "overwrite", false, "Overwrites the RCON entry if it already exists")
	ac.Cmd.Flags().BoolVar(&ac.Data.SetDefault, "default", false, "Sets the RCON entry as the default RCON entry")
}

// initData initializes and validates the AddCommand data.
func (ac *AddCommand) initData() error {
	if strings.TrimSpace(ac.Data.Name) == "" {
		name, nameErr := initRconName()
		if nameErr != nil {
			return nameErr
		}

		ac.Data.Name = name
	}

	initErr := initEntry(&ac.Data.Entry)
	if initErr != nil {
		return initErr
	}

	validErr := validateEntry(ac.Data.Entry)
	if validErr != nil {
		return validErr
	}

	return nil
}
