package rcondefault

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bobllor/rcon-cli/app/edit/internal"
	"github.com/bobllor/rcon-cli/app/utils"
	"github.com/bobllor/rcon-cli/app/utils/paths"
	"github.com/bobllor/rcon-cli/config"
	"github.com/spf13/cobra"
)

// DefaultCommand is used as a subcommand for EditCommand. It is responsible
// for handling changes to the default RCON entry, if one exists.
type DefaultCommand struct {
	Cmd  *cobra.Command
	path paths.AppPath
	data DefaultData
}

type DefaultData struct {
	// Name is the new name of the default RCON entry. This will change the
	// value and also the entry in the configuration file.
	Name  string
	Entry config.RconEntry
	// Remove is a flag used to handle the removing the default RCON entry.
	// This does not remove the actual entry, but rather the default RCON entry value.
	Remove bool
	// updates is used to track what was updated in the edit. It will be displayed at
	// the end of the call.
	updates []string
}

func NewDefaultCommand(appPaths paths.AppPath) *DefaultCommand {
	cmd := &DefaultCommand{
		Cmd: &cobra.Command{
			Use:   "default [flags]",
			Short: "Edit the default RCON entry",
		},
		path: appPaths,
		data: DefaultData{
			updates: make([]string, 0),
		},
	}

	cmd.Cmd.Run = cmd.Run
	cmd.Cmd.PreRunE = cmd.PreRunE
	cmd.initFlags()

	return cmd
}

func (dc *DefaultCommand) Run(cmd *cobra.Command, args []string) {
	cfg, err := config.LoadConfigurationIfMissing(dc.path.Config)
	if err != nil {
		utils.PrintFatal(err)
	}

	if cfg.DefaultRcon == "" {
		fmt.Println("No default RCON entry found (default entry is empty)")
		return
	}

	dncDefaultRcon := cfg.DefaultRcon
	dncDefaultEntry := cfg.RconEntries[dncDefaultRcon]

	rconName := cfg.DefaultRcon
	baseEntry := cfg.RconEntries[rconName]

	if !dc.data.Remove {
		if cmd.Flag("name").Changed {
			if dc.data.Name == "-" {
				newName, err := internal.HandleInteractiveMode(internal.NAME_PROMPT, false)
				if err != nil {
					utils.PrintFatal(err)
				}

				dc.data.Name = newName
			}

			err := internal.HandleEditCfgRconName(cfg, rconName, dc.data.Name)
			if err != nil {
				utils.PrintFatal(err)
			}

			dc.data.updates = append(dc.data.updates, fmt.Sprintf(" - RCON name %s -> %s", rconName, dc.data.Name))
			// new name can still be the original default rcon name
			rconName = dc.data.Name
		}

		if cmd.Flag("address").Changed {
			if dc.data.Entry.Address == "-" {
				newAddr, err := internal.HandleInteractiveMode(internal.ADDRESS_PROMPT, false)
				if err != nil {
					utils.PrintFatal(err)
				}

				dc.data.Entry.Address = newAddr
			}

			err = internal.HandleEditAddress(&baseEntry, dc.data.Entry.Address)
			if err != nil {
				utils.PrintFatal(err)
			}

			dc.data.updates = append(dc.data.updates, fmt.Sprintf(" - RCON address %s -> %s", dncDefaultEntry.Address, dc.data.Entry.Address))
		}

		if cmd.Flag("password").Changed {
			if dc.data.Entry.Password == "-" {
				newPw, err := internal.HandleInteractiveMode(internal.PASSWORD_PROMPT, true)
				if err != nil {
					utils.PrintFatal(err)
				}

				dc.data.Entry.Password = newPw
			}
			err = internal.HandleEditPassword(&baseEntry, dc.data.Entry.Password)
			if err != nil {
				utils.PrintFatal(err)
			}

			dc.data.updates = append(dc.data.updates, " - RCON password updated")
		}

		cfg.AddEntry(rconName, baseEntry)
	} else {
		cfg.DefaultRcon = ""

		dc.data.updates = append(dc.data.updates, " - Removed as default RCON")
	}

	err = cfg.WriteFile(dc.path.Config)
	if err != nil {
		utils.PrintFatal(err)
	}

	fmt.Printf("Updated entry %s:\n", dncDefaultRcon)
	fmt.Println(strings.Join(dc.data.updates, "\n"))
}

func (dc *DefaultCommand) PreRunE(cmd *cobra.Command, args []string) error {
	if cmd.Flag("name").Changed && strings.TrimSpace(dc.data.Name) == "" {
		return errors.New("cannot have an empty RCON name")
	}

	return nil
}

func (dc *DefaultCommand) initFlags() {
	dc.Cmd.Flags().StringVarP(&dc.data.Entry.Address, "address", "a", "", "New RCON address")
	dc.Cmd.Flags().StringVarP(&dc.data.Entry.Password, "password", "p", "", "New RCON password")
	dc.Cmd.Flags().StringVarP(&dc.data.Name, "name", "n", "", "New RCON default name")

	dc.Cmd.Flags().BoolVar(&dc.data.Remove, "remove", false, "Remove the default RCON entry")

	dc.Cmd.MarkFlagsMutuallyExclusive("remove", "address")
	dc.Cmd.MarkFlagsMutuallyExclusive("remove", "password")
	dc.Cmd.MarkFlagsMutuallyExclusive("remove", "name")
	dc.Cmd.MarkFlagsOneRequired("address", "password", "remove", "name")
}
