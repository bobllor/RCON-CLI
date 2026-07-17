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
			Use:   "list [entry]... [flags]",
			Short: "Lists RCON entries and its information",
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
		targets := []string{}

		// already handled in PreRunE but for sanity check
		if lc.Data.ShowDefault {
			if cfg.DefaultRcon == "" {
				fmt.Println("No default RCON entry found")
				return
			}
			targets = append(targets, cfg.DefaultRcon)
		} else {
			targets = append(targets, args...)
		}

		targetStrings := []string{}
		for _, target := range targets {
			str := lc.listTargetString(target, lc.Data.ShowPassword, cfg)

			targetStrings = append(targetStrings, str)
		}

		fmt.Println(strings.Join(targetStrings, "\n"))
	}
}

// InitFlags initializes the flags for the command.
func (lc *ListCommand) InitFlags() {
	lc.Cmd.Flags().BoolVar(&lc.Data.ShowPassword, "show-password", false, "Shows the password of RCON entries")
	lc.Cmd.Flags().BoolVar(&lc.Data.ShowDefault, "default", false, "Show the default RCON entry")
}

func (lc *ListCommand) PreRunE(cmd *cobra.Command, args []string) error {
	if len(args) > 0 && lc.Data.ShowDefault {
		return errors.New("flag --default must be used with no additional arguments")
	}

	return nil
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
// If the target does not exist, then it will return an invalid string.
func (lc *ListCommand) listTargetString(target string, showPassword bool, cfg *config.Configuration) string {
	entry, ok := cfg.RconEntries[target]
	if !ok {
		return fmt.Sprintf("Entry %s does not exist", target)
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

	return str
}
