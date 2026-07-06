package edit

import (
	"errors"
	"fmt"

	"github.com/bobllor/rcon/app/types"
	"github.com/bobllor/rcon/app/utils"
	"github.com/bobllor/rcon/config"
	"github.com/spf13/cobra"
)

type EditCommand struct {
	Cmd  *cobra.Command
	Data EditData
	Path types.AppPath
}

type EditData struct {
	Name       string
	Address    string
	Password   string
	NewDefault bool
}

func NewEditCommand(paths types.AppPath) *EditCommand {
	cmd := EditCommand{
		Cmd: &cobra.Command{
			Use:   "edit [entry] [flags]",
			Short: "Edit the values of a RCON entry",
			Args: func(cmd *cobra.Command, args []string) error {
				if len(args) < 1 {
					return errors.New("missing RCON entry argument")
				}
				if len(args) > 1 {
					return errors.New("only one RCON entry can be edited at a time")
				}

				return nil
			},
		},
		Path: paths,
	}

	cmd.Cmd.Run = cmd.Run
	cmd.InitFlags()

	return &cmd
}

func (ec *EditCommand) Run(cmd *cobra.Command, args []string) {
	// Cmd.Args checks length
	target := args[0]

	cfg, err := config.LoadConfigurationIfMissing(ec.Path.Config)
	if err != nil {
		utils.PrintFatal(err)
	}
	if !cfg.EntryExist(target) {
		utils.PrintFatalString(fmt.Sprintf("entry %s does not exist", target))
	}

	ogTarget := target
	currEntry := cfg.RconEntries[target]

	if cmd.Flag("name").Changed {
		err := ec.handleEditRconName(target, ec.Data.Name, cfg)
		if err != nil {
			utils.PrintFatal(err)
		}

		target = ec.Data.Name
	}

	if cmd.Flag("address").Changed {
		currEntry.Address = ec.Data.Address
	}

	if cmd.Flag("password").Changed {
		pw, err := ec.handleEditPassword(ec.Data.Password)
		if err != nil {
			utils.PrintFatal(err)
		}

		currEntry.Password = pw
	}

	if cmd.Flag("default").Changed {
		cfg.DefaultRcon = target
	}

	cfg.AddEntry(target, currEntry)
	err = cfg.WriteFile(ec.Path.Config)
	if err != nil {
		utils.PrintFatal(err)
	}

	fmt.Printf("Updated entry %s\n", ogTarget)
}

func (ec *EditCommand) InitFlags() {
	ec.Cmd.Flags().StringVarP(&ec.Data.Address, "address", "a", "", "The new address of the RCON entry")
	ec.Cmd.Flags().StringVarP(&ec.Data.Password, "password", "p", "", "The new password of the RCON entry")
	ec.Cmd.Flags().StringVar(&ec.Data.Name, "name", "", "The new name of the RCON entry")

	ec.Cmd.Flags().BoolVar(&ec.Data.NewDefault, "default", false, "Set the RCON entry as the new default entry")

	ec.Cmd.MarkFlagsOneRequired("address", "password", "name", "default")
}

// handleEditRconName handles deleting the entry of the RCON target string.
// Prior to updating the configuration, it will validate the new name.
func (ec *EditCommand) handleEditRconName(target string, newName string, cfg *config.Configuration) error {
	if cfg.EntryExist(newName) {
		return fmt.Errorf("%s already exists as an entry, no changes were made", ec.Data.Name)
	}

	// gets added as a new entry at the end of the command
	cfg.DeleteEntry(target)
	if cfg.DefaultRcon == target {
		cfg.DefaultRcon = ""
	}

	return nil
}

// handleEditPassword handles the password change. The given password value will
// be validated. Upon success, it will return the original password value.
//
// If the password is the character "-", then it will trigger an interactive prompt
// for the password.
func (ec *EditCommand) handleEditPassword(pwValue string) (string, error) {
	err := utils.ValidatePassword(pwValue)
	if err != nil {
		return "", err
	}

	if pwValue == "-" {
		fmt.Print("Enter the new RCON password: ")
		pw, err := utils.ReadInputHidden()
		if err != nil {
			return "", err
		}
		err = utils.ValidatePassword(pw)
		if err != nil {
			return "", err
		}

		pwValue = pw
	}

	return pwValue, nil
}
