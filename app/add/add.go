package add

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bobllor/rcon-cli/app/utils"
	"github.com/bobllor/rcon-cli/app/utils/paths"
	"github.com/bobllor/rcon-cli/config"
	"github.com/spf13/cobra"
)

// AddCommand is the subcommand used to handle adding new RCON entries
// into the configuration.
type AddCommand struct {
	Cmd  *cobra.Command
	Path paths.AppPath
	Data AddData
}

type AddData struct {
	// Name is the RCON name entry. This must be unique, but can be overwritten
	// with a flag.
	Name string
	// Entry is the RCON entry data used to connect and communicate to RCON.
	Entry config.RconEntry
	// Overwrite is used to overwrite an existing RCON entry in the
	// configuration file. If an entry does not exist, then this will do
	// nothing.
	Overwrite bool
	// SetDefault is used to set the new RCON entry as the default RCON entry.
	// The default RCON is used when running a command, it will automatically
	// run the command to this default RCON.
	SetDefault bool
}

func NewAddCommand(appPaths paths.AppPath) *AddCommand {
	cmd := &AddCommand{
		Cmd: &cobra.Command{
			Use:   "add [entry]... [flags]",
			Short: "Add a new RCON entry",
			Long:  "Add a new RCON entry into the configuration",
		},
		Path: appPaths,
	}

	cmd.Cmd.Run = cmd.Run
	cmd.Cmd.PreRunE = cmd.PreRunE
	cmd.AddInitFlags()

	return cmd
}

// Run is the main entry point of the AddCommand. It adds a new entry
// to the config.
func (ac *AddCommand) Run(cmd *cobra.Command, args []string) {
	// enables space in entries
	if len(args) > 0 {
		ac.Data.Name = strings.Join(args, " ")
	}

	// default/DEFAULT or whatever combination is reserved and cannot be used
	if strings.TrimSpace(strings.ToLower(ac.Data.Name)) == "default" {
		utils.PrintFatalString("cannot use reserved word 'default' as an RCON entry name")
	}

	err := ac.initData()
	if err != nil {
		utils.PrintFatal(err)
	}

	cfg, err := config.LoadConfigurationIfMissing(ac.Path.Config)
	if err != nil {
		utils.PrintFatal(err)
	}

	if !ac.Data.Overwrite {
		exist := cfg.HasEntry(ac.Data.Name)
		if exist {
			utils.PrintFatal(fmt.Errorf(`RCON entry "%s" already exists`, ac.Data.Name))
		}
	}
	if ac.Data.SetDefault {
		cfg.DefaultRcon = ac.Data.Name
	}

	cfg.AddEntry(ac.Data.Name, ac.Data.Entry)
	writeErr := cfg.WriteFile(ac.Path.Config)
	if writeErr != nil {
		utils.PrintFatal(writeErr)
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

// PreRunE validates data prior to Run.
func (ac *AddCommand) PreRunE(cobra *cobra.Command, args []string) error {
	if len(args) > 0 && strings.TrimSpace(ac.Data.Name) != "" {
		return errors.New("cannot have entry args and use the -n name flag")
	}

	return nil
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
		name, nameErr := utils.InitRconName()
		if nameErr != nil {
			return nameErr
		}

		ac.Data.Name = name
	}

	initErr := utils.InitEntry(&ac.Data.Entry)
	if initErr != nil {
		return initErr
	}

	validErr := utils.ValidateEntry(ac.Data.Entry)
	if validErr != nil {
		return validErr
	}

	return nil
}
