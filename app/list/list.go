package list

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/bobllor/rcon-cli/app/utils"
	"github.com/bobllor/rcon-cli/app/utils/paths"
	"github.com/bobllor/rcon-cli/config"
	"github.com/spf13/cobra"
)

type ListCommand struct {
	Cmd  *cobra.Command
	Path paths.AppPath
	Data ListData
}

type ListData struct {
	ShowPassword bool
	ShowDefault  bool
}

func NewListCommand(paths paths.AppPath) *ListCommand {
	listCmd := &ListCommand{
		Cmd: &cobra.Command{
			Use:     "list [entry]... [flags]",
			Short:   "Lists RCON entries and its information",
			Aliases: []string{"ls"},
		},
		Path: paths,
	}

	listCmd.Cmd.Run = listCmd.Run
	listCmd.Cmd.PreRunE = listCmd.PreRunE
	listCmd.InitFlags()

	return listCmd
}

// Run is the main entry point of the ListCommand.
func (lc *ListCommand) Run(cmd *cobra.Command, args []string) {
	cfg, err := config.LoadConfigurationIfMissing(lc.Path.Config)
	if err != nil {
		utils.PrintFatal(err)
	}

	if len(args) < 1 && !lc.Data.ShowDefault {
		str := lc.listAllString(cfg)

		fmt.Println("RCON Entry -> RCON Address")
		fmt.Println(str)
	} else {
		details := lc.getDetailedEntries(cfg, args, lc.Data.ShowDefault)

		fmt.Println(details)
	}
}

// InitFlags initializes the flags for the command.
func (lc *ListCommand) InitFlags() {
	lc.Cmd.Flags().BoolVar(&lc.Data.ShowPassword, "show-password", false, "Shows the password of RCON entries")
	lc.Cmd.Flags().BoolVar(&lc.Data.ShowDefault, "default", false, "Show the default RCON entry")
}

func (lc *ListCommand) PreRunE(cmd *cobra.Command, args []string) error {
	if len(args) > 0 && lc.Data.ShowDefault {
		return errors.New("flag --default must be used with no arguments")
	}

	return nil
}

// getDetailedEntries retrieves the string of the detailed information
// of RCON entries from the configuration file.
//
// If a default RCON is set and the default entry is given as an argument, then
// it will always be the first entry in the slice during the string build.
// The slice will be sorted in ascending order. Missing values will bubble down
// to the bottom.
func (lc *ListCommand) getDetailedEntries(cfg *config.Configuration, entries []string, showDefault bool) string {
	targets := []string{}

	// already handled in PreRunE but for sanity check
	if showDefault {
		if cfg.DefaultRcon == "" {
			fmt.Println("No default RCON entry found")
			return "No default RCON entry found"
		}
		targets = append(targets, cfg.DefaultRcon)
	} else {
		defaultRcon := ""
		for _, entry := range entries {
			if entry == cfg.DefaultRcon {
				defaultRcon = entry
			} else {
				targets = append(targets, entry)
			}
		}

		slices.Sort(targets)
		if defaultRcon != "" {
			// ensures default rcon will always be first in the slice
			targets = slices.Concat([]string{defaultRcon}, targets)
		}
	}

	targetStrings := []string{}
	invalidStrings := []string{}
	for _, target := range targets {
		str, found := lc.listTargetString(target, lc.Data.ShowPassword, cfg)

		if found {
			targetStrings = append(targetStrings, str)
		} else {
			invalidStrings = append(invalidStrings, str)
		}
	}

	// formatting
	if len(targetStrings) > 0 && len(invalidStrings) > 0 {
		invalidStrings[0] = "\n" + invalidStrings[0]
	}

	targetStrings = slices.Concat(targetStrings, invalidStrings)

	return strings.Join(targetStrings, "\n")
}

// listAllString returns a string of RCON entries with their address. It does not
// display the password.
func (lc *ListCommand) listAllString(cfg *config.Configuration) string {
	entries := []string{}

	for k, v := range cfg.RconEntries {
		str := fmt.Sprintf("%s -> %s", k, v.Address)
		if k == cfg.DefaultRcon {
			// default goes to the top
			str = fmt.Sprintf("%s (default)", str)
			entries = slices.Insert(entries, 0, str)
		} else {
			entries = append(entries, str)
		}
	}

	if len(entries) == 0 {
		entries = append(entries, "No entries found, to add an entry run the command: gorcon add")
	}

	return strings.Join(entries, "\n")
}

// listTargetString returns a string of an target RCON entry. It does not display
// the password.
//
// If the target does not exist, then it will return an invalid string and a false status.
func (lc *ListCommand) listTargetString(target string, showPassword bool, cfg *config.Configuration) (string, bool) {
	entry, ok := cfg.RconEntries[target]
	if !ok {
		return fmt.Sprintf("Entry %s does not exist", target), false
	}

	padding := len("RCON Entry") + 8
	pw := "********"
	if showPassword {
		pw = entry.Password
	}

	if target == cfg.DefaultRcon {
		target = fmt.Sprintf("%s (default)", target)
	}

	str := fmt.Sprintf(
		"%-*s %s\n%-*s %s\n%-*s %s",
		padding,
		"RCON Entry:",
		target,
		padding,
		"Address:",
		entry.Address,
		padding,
		"Password:",
		pw,
	)

	return str, true
}
