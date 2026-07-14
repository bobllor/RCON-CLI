package root

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bobllor/rcon/app/utils"
	"github.com/bobllor/rcon/app/utils/paths"
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
	Path paths.AppPath
}

type RootData struct {
	// Entry is the RCON entry information. By default it will be empty values, but
	// will be populated by an existing entry found in the config file. If not found
	// or the values remain empty, the program will go into interactive mode.
	Entry config.RconEntry
	// Target is the target RCON entry to send the command to. This
	// will overwrite the default RCON entry.
	Target string
	// Command is the command to send. Command and args cannot be used together,
	// and will return an error.
	Command string
}

// NewRootCommand creates a new RootCommand and its initialization flags.
func NewRootCommand(appPaths paths.AppPath) *RootCommand {
	cmd := &RootCommand{
		Cmd: &cobra.Command{
			Use:   "rcon <args>... [flags]",
			Short: "Execute a command with RCON",
		},
		Data: RootData{},
		Path: appPaths,
	}

	cmd.Cmd.Run = cmd.RootRun
	cmd.Cmd.CompletionOptions.DisableDefaultCmd = true
	// required due to subcommands changing arg parsing rules
	cmd.Cmd.Args = cmd.Args
	cmd.Cmd.PreRunE = cmd.PreRunE

	cmd.RootInitFlags()

	return cmd
}

// RootRun is the main entry point for the root CMD. This will run
// the execution of the command to the server.
func (r *RootCommand) RootRun(cmd *cobra.Command, args []string) {
	// this does not create a cfg file.
	cfg, cfgErr := utils.LoadConfiguration(r.Path.Config)
	if cfgErr != nil {
		utils.PrintFatal(cfgErr)
	} else {
		if cfg.DefaultRcon != "" {
			cfgEntry, ok := cfg.RconEntries[cfg.DefaultRcon]
			if ok {
				r.Data.Entry = cfgEntry
			}
		}
	}

	// target will overwrite the default if it exists
	if r.Data.Target != "" {
		cfgEntry, ok := cfg.RconEntries[r.Data.Target]
		if ok {
			r.Data.Entry = cfgEntry
		} else {
			utils.PrintFatal(fmt.Errorf("RCON entry target %s is not found", r.Data.Target))
		}
	}

	initErr := r.initEntry()
	if initErr != nil {
		utils.PrintFatal(initErr)
	}

	con, err := rcon.NewRcon(r.Data.Entry.Address)
	if err != nil {
		utils.PrintFatalf("Failed to establish connection: %v", err)
	}
	defer con.Close()

	loginErr := con.Authenticate(r.Data.Entry.Password)
	if loginErr != nil {
		utils.PrintFatalf("Failed to authenticate: %v", loginErr)
	}

	command := r.getCommandString(args)
	cmdRes, cmdErr := con.Command(command)
	if cmdErr != nil {
		utils.PrintFatalf("Failed to run command: %v", cmdErr)
	}

	if strings.TrimSpace(cmdRes) == "" {
		cmdRes = "Command executed"
	}

	fmt.Println(cmdRes)
}

// Args is the function used to check the arguments and certain flags.
func (r *RootCommand) Args(cmd *cobra.Command, args []string) error {
	if len(args) < 1 && r.Data.Command == "" {
		return errors.New("must have at least one argument")
	}
	if len(args) > 0 && r.Data.Command != "" {
		return errors.New("cannot use -c/--command flag and have arguments")
	}

	return nil
}

// PreRunE is used to validate the certain data prior to running the command.
func (r *RootCommand) PreRunE(cmd *cobra.Command, args []string) error {
	if strings.TrimSpace(r.Data.Command) == "" && len(args) < 1 {
		return errors.New("cannot have an empty command")
	}

	return nil
}

func (r *RootCommand) RootInitFlags() {
	r.Cmd.Flags().StringVarP(&r.Data.Entry.Address, "address", "a", "", "RCON address")
	r.Cmd.Flags().StringVarP(&r.Data.Entry.Password, "password", "p", "", "RCON password")

	r.Cmd.Flags().StringVarP(&r.Data.Target, "target", "t", "", "RCON entry target to run the command on")
	r.Cmd.Flags().StringVarP(&r.Data.Command, "command", "c", "", "The command to send via RCON")
}

// getCommandString retrieves the command string. It handles
// both the flag and the args.
func (r *RootCommand) getCommandString(args []string) string {
	var command string
	if r.Data.Command != "" {
		command = r.Data.Command
	} else {
		command = strings.Join(args, " ")
	}

	return command
}

// InitEntry initializes the RCON entry and validates it. If the data is already
// configured via the config file, then it will instead use the data from the file.
//
// The entry will be mutated in place. If an error occurs, it will return an error.
func (r *RootCommand) initEntry() error {
	err := utils.InitEntry(&r.Data.Entry)
	if err != nil {
		return err
	}

	validErr := utils.ValidateEntry(r.Data.Entry)
	if validErr != nil {
		return validErr
	}

	return nil
}
