package edit

import (
	"errors"
	"fmt"
	"strings"

	rcondefault "github.com/bobllor/rcon-cli/app/edit/default"
	"github.com/bobllor/rcon-cli/app/edit/internal"
	"github.com/bobllor/rcon-cli/app/utils"
	"github.com/bobllor/rcon-cli/app/utils/paths"
	"github.com/bobllor/rcon-cli/config"
	"github.com/spf13/cobra"
)

type EditCommand struct {
	Cmd  *cobra.Command
	Data EditData
	Path paths.AppPath
}

type EditData struct {
	Name       string
	Address    string
	Password   string
	NewDefault bool
}

func NewEditCommand(paths paths.AppPath) *EditCommand {
	cmd := EditCommand{
		Cmd: &cobra.Command{
			Use:   "edit <entry> [flags]",
			Short: "Edit the values of a RCON entry",
			Args: func(cmd *cobra.Command, args []string) error {
				if len(args) == 0 {
					return errors.New("missing RCON entry")
				}
				if len(args) > 1 {
					return errors.New("only one RCON entry can be edited at a time")
				}

				return nil
			},
		},
		Path: paths,
	}

	defCmd := rcondefault.NewDefaultCommand(paths)

	cmd.Cmd.AddCommand(defCmd.Cmd)
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
	newUpdates, err := ec.runNormal(cfg, cmd, args)
	if err != nil {
		utils.PrintFatal(err)
	}

	updates = append(updates, newUpdates...)
	err = cfg.WriteFile(ec.Path.Config)
	if err != nil {
		utils.PrintFatal(err)
	}

	// checked in Cmd.Args
	fmt.Printf("Updated entry %s:\n", args[0])
	fmt.Printf("%s\n", strings.Join(updates, "\n"))
}

func (ec *EditCommand) InitFlags() {
	ec.Cmd.Flags().StringVarP(&ec.Data.Address, "address", "a", "", "The new address of the RCON entry")
	ec.Cmd.Flags().StringVarP(&ec.Data.Password, "password", "p", "", "The new password of the RCON entry")
	ec.Cmd.Flags().StringVarP(&ec.Data.Name, "name", "n", "", "The new name of the RCON entry")

	ec.Cmd.Flags().BoolVar(&ec.Data.NewDefault, "default", false, "Set the RCON entry as the new default entry")

	ec.Cmd.MarkFlagsOneRequired("address", "password", "name", "default")
}

// runNormal is the normal run process for editing.
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

	dncEntry := cfg.RconEntries[target]
	currEntry := cfg.RconEntries[target]

	if cmd.Flag("name").Changed {
		if ec.Data.Name == "-" {
			newName, err := internal.HandleInteractiveMode(internal.NAME_PROMPT, false)
			if err != nil {
				return nil, err
			}

			ec.Data.Name = newName
		}

		err := internal.HandleEditCfgRconName(cfg, target, ec.Data.Name)
		if err != nil {
			return nil, err
		}

		updates = append(updates, fmt.Sprintf(" - RCON name %s -> %s", target, ec.Data.Name))
		// updates target name for adding the entry
		target = ec.Data.Name
	}

	if cmd.Flag("address").Changed {
		if ec.Data.Address == "-" {
			newAddr, err := internal.HandleInteractiveMode(internal.ADDRESS_PROMPT, false)
			if err != nil {
				return nil, err
			}

			ec.Data.Address = newAddr
		}

		err := internal.HandleEditAddress(&currEntry, ec.Data.Address)
		if err != nil {
			return nil, err
		}

		updates = append(updates, fmt.Sprintf(" - RCON address %s -> %s", dncEntry.Address, ec.Data.Address))
	}

	if cmd.Flag("password").Changed {
		if ec.Data.Password == "-" {
			pw, err := internal.HandleInteractiveMode(internal.PASSWORD_PROMPT, true)
			if err != nil {
				return nil, err
			}

			ec.Data.Password = pw
		}

		err := internal.HandleEditPassword(&currEntry, ec.Data.Password)
		if err != nil {
			return nil, err
		}

		updates = append(updates, " - RCON password updated")
	}

	if cmd.Flag("default").Changed {
		cfg.DefaultRcon = target
		updates = append(updates, fmt.Sprintf(" - default RCON -> %s", target))
	}

	cfg.AddEntry(target, currEntry)

	return updates, nil
}
