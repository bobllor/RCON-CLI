package edit

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bobllor/rcon/app/utils"
	"github.com/bobllor/rcon/app/utils/paths"
	"github.com/bobllor/rcon/config"
	"github.com/spf13/cobra"
)

type EditCommand struct {
	Cmd  *cobra.Command
	Data EditData
	Path paths.AppPath
}

type EditData struct {
	Name          string
	Address       string
	Password      string
	NewDefault    bool
	RemoveDefault bool
}

func NewEditCommand(paths paths.AppPath) *EditCommand {
	cmd := EditCommand{
		Cmd: &cobra.Command{
			Use:   "edit [entry] [flags]",
			Short: "Edit the values of a RCON entry",
			Args: func(cmd *cobra.Command, args []string) error {
				// following conditions:
				//	1. no args are given and w/o --rm-default
				//	2. arg is given with --rm-default
				//	3. more than 1 arg is given
				// --rm-default can only be used alone
				if len(args) < 1 && !cmd.Flag("rm-default").Changed {
					return errors.New("missing RCON entry argument")
				}
				if len(args) == 1 && cmd.Flag("rm-default").Changed {
					return fmt.Errorf("--rm-default can only be used with no arguments (given %s)", args[0])
				}
				if len(args) > 1 {
					return errors.New("only one RCON entry argument is allowed")
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
	cfg, err := config.LoadConfigurationIfMissing(ec.Path.Config)
	if err != nil {
		utils.PrintFatal(err)
	}

	updates := []string{}
	// --rm-default is a non-argument based flag.
	if cmd.Flag("rm-default").Changed {
		if cfg.DefaultRcon == "" {
			fmt.Println("Default RCON is already empty, no changes were made")
			return
		}
		updates = append(updates, fmt.Sprintf(" - Removed default RCON %s", cfg.DefaultRcon))
		cfg.DefaultRcon = ""
	} else {
		newUpdates, err := ec.runNormal(cfg, cmd, args)
		if err != nil {
			utils.PrintFatal(err)
		}

		updates = append(updates, newUpdates...)
	}
	err = cfg.WriteFile(ec.Path.Config)
	if err != nil {
		utils.PrintFatal(err)
	}

	// args[0] will always exist, --rm-default appends one.
	fmt.Println("Updated configuration:")
	fmt.Printf("%s\n", strings.Join(updates, "\n"))
}

func (ec *EditCommand) InitFlags() {
	ec.Cmd.Flags().StringVarP(&ec.Data.Address, "address", "a", "", "The new address of the RCON entry")
	ec.Cmd.Flags().StringVarP(&ec.Data.Password, "password", "p", "", "The new password of the RCON entry")
	ec.Cmd.Flags().StringVar(&ec.Data.Name, "name", "", "The new name of the RCON entry")

	ec.Cmd.Flags().BoolVar(&ec.Data.NewDefault, "default", false, "Set the RCON entry as the new default entry")
	ec.Cmd.Flags().BoolVar(&ec.Data.RemoveDefault, "rm-default", false, "Removes the target from the default entry")

	ec.Cmd.MarkFlagsOneRequired("address", "password", "name", "default", "rm-default")
}

// runNormal is the normal run process for editing, without the use of the flag --rm-default.
// Upon success, it will modify the Configuration RCON entries map with the newly modified entry.
//
// Validation and input mutations are called within this method.
//
// If any errors occur it will return an error, otherwise it will return a slice of strings of
// the updates for the program.
func (ec *EditCommand) runNormal(cfg *config.Configuration, cmd *cobra.Command, args []string) ([]string, error) {
	// Cmd.Args checks length
	target := args[0]

	updates := []string{}

	if !cfg.HasEntry(target) {
		return nil, fmt.Errorf("entry %s does not exist", target)
	}

	currEntry := cfg.RconEntries[target]

	if cmd.Flag("name").Changed {
		err := ec.handleEditRconName(target, ec.Data.Name, cfg)
		if err != nil {
			return nil, err
		}

		updates = append(updates, fmt.Sprintf(" - RCON name %s -> %s", target, ec.Data.Name))
		// updates target name for adding the entry
		target = ec.Data.Name
	}

	if cmd.Flag("address").Changed {
		updates = append(updates, fmt.Sprintf(" - RCON address %s -> %s", currEntry.Address, ec.Data.Address))
		currEntry.Address = ec.Data.Address
	}

	if cmd.Flag("password").Changed {
		pw, err := ec.handleEditPassword(ec.Data.Password)
		if err != nil {
			return nil, err
		}

		currEntry.Password = pw
		updates = append(updates, fmt.Sprintf(" - %s password updated", target))
	}

	if cmd.Flag("default").Changed {
		cfg.DefaultRcon = target
		updates = append(updates, fmt.Sprintf(" - default RCON -> %s", target))
	}

	cfg.AddEntry(target, currEntry)

	return updates, nil
}

// handleEditRconName handles deleting the entry of the RCON target string.
// Prior to updating the configuration, it will validate the new name.
func (ec *EditCommand) handleEditRconName(target string, newName string, cfg *config.Configuration) error {
	if cfg.HasEntry(newName) {
		return fmt.Errorf("%s already exists as an entry, no changes were made", ec.Data.Name)
	}
	if strings.TrimSpace(newName) == "" {
		return errors.New("cannot have an empty name")
	}

	// gets added as a new entry at the end of the parent command
	cfg.DeleteEntry(target)
	// ensures that if the target is the default entry,
	// keep that entry as the default when switching over.
	if cfg.DefaultRcon == target {
		cfg.DefaultRcon = ec.Data.Name
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
