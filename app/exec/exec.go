package rconexec

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bobllor/rcon-cli/app/utils"
	"github.com/bobllor/rcon-cli/app/utils/paths"
	"github.com/bobllor/rcon-cli/config"
	"github.com/bobllor/rcon-cli/rcon"
	"github.com/spf13/cobra"
)

type ExecCommand struct {
	Cmd  *cobra.Command
	Path paths.AppPath
	data ExecData
}

type ExecData struct {
	// Entry is the RCON entry information. By default it will be empty values, but
	// will be populated by an existing entry found in the config file. If not found
	// or the values remain empty, the program will go into interactive mode.
	Entry config.RconEntry
	// Target is the target RCON entry to send the command to. This
	// will overwrite the default RCON entry.
	Target string
}

func NewExecCommand(appPaths paths.AppPath) *ExecCommand {
	cmd := ExecCommand{
		Cmd: &cobra.Command{
			Use:   "exec <args>...",
			Short: "Execute a command to an RCON server",
		},
		Path: appPaths,
	}

	cmd.Cmd.Run = cmd.Run
	cmd.Cmd.PreRunE = cmd.PreRunE
	cmd.initFlags()

	return &cmd
}

func (e *ExecCommand) Run(cmd *cobra.Command, args []string) {
	cfg, err := e.loadConfig(e.Path.Config)
	if err != nil {
		utils.PrintFatal(err)
	}

	entry, err := e.getEntry(cfg)
	if err != nil {
		utils.PrintFatal(err)
	}

	entry, err = e.initAndValidateEntry(entry)
	if err != nil {
		utils.PrintFatal(err)
	}

	res, err := e.execCommand(entry, args)
	if err != nil {
		utils.PrintFatal(err)
	}

	fmt.Println(res)
}

// PreRunE is used to validate the command data prior to running the main command.
func (e *ExecCommand) PreRunE(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("cannot have empty arguments")
	}

	return nil
}

func (e *ExecCommand) initFlags() {
	e.Cmd.Flags().StringVarP(&e.data.Entry.Address, "address", "a", "", "RCON address")
	e.Cmd.Flags().StringVarP(&e.data.Entry.Password, "password", "p", "", "RCON password")

	e.Cmd.Flags().StringVarP(&e.data.Target, "target", "t", "", "The target RCON entry to execute commands on")

	e.Cmd.MarkFlagsMutuallyExclusive("target", "address")
	e.Cmd.MarkFlagsMutuallyExclusive("target", "password")
}

// loadConfig loads the configuration and returns the configuration data.
//
// This does not create the configuration file due to the interactive mode.
func (e *ExecCommand) loadConfig(root string) (*config.Configuration, error) {
	// this does not create a cfg file.
	cfg, cfgErr := utils.LoadConfiguration(root)
	if cfgErr != nil {
		return nil, cfgErr
	}

	return cfg, nil
}

// execCommand executes the command from the arguments. It returns the
// string response of the command, if it has one, otherwise a default string
// that the command has been executed.
func (e *ExecCommand) execCommand(entry config.RconEntry, args []string) (string, error) {
	con, err := rcon.NewRcon(entry.Address)
	if err != nil {
		return "", fmt.Errorf("Failed to establish connection: %v", err)
	}
	defer con.Close()

	loginErr := con.Authenticate(entry.Password)
	if loginErr != nil {
		return "", fmt.Errorf("Failed to authenticate: %v", loginErr)
	}

	command := e.getCommandString(args)
	cmdRes, cmdErr := con.Command(command)
	if cmdErr != nil {
		return "", fmt.Errorf("Failed to run command: %v", cmdErr)
	}

	if strings.TrimSpace(cmdRes) == "" {
		cmdRes = "Command executed"
	}

	return cmdRes, nil
}

// getEntry retrieves the entry for use. This will be a zeroed RconEntry,
// the default RconEntry, a target RconEntry, or the given entry values.
//
// The only error that is returned is if the target does not exist in the entries.
func (e *ExecCommand) getEntry(cfg *config.Configuration) (config.RconEntry, error) {
	var def config.RconEntry
	if cfg.DefaultRcon != "" {
		cfgEntry, ok := cfg.RconEntries[cfg.DefaultRcon]
		if ok {
			def = cfgEntry
		}
	}

	// if a target is given or any non-empty values are given, then default
	// will always be overwritten.

	if e.data.Target != "" {
		cfgEntry, ok := cfg.RconEntries[e.data.Target]
		if ok {
			return cfgEntry, nil
		} else {
			return def, fmt.Errorf("RCON entry target %s is not found", e.data.Target)
		}
	}

	// target and address/password flags cannot be used together.
	// this is handled in PreRunE, so it does not matter here.
	// this ensures defaults do not get overwritten unless a value
	// is expliclity given.
	if e.data.Entry.Address != "" || e.data.Entry.Password != "" {
		def.Address = e.data.Entry.Address
		def.Password = e.data.Entry.Password
	}

	return def, nil
}

// initAndValidateEntry is used to initialize the entry with an interactive mode
// if the entry contains empty values and validates the entry afterwards.
//
// The interactive mode is dependent on the empty field values, which will prompt for
// the missing value.
//
// It will return the either the original entry, an updated version, or a zeroed version if
// an error occurs.
func (e *ExecCommand) initAndValidateEntry(entry config.RconEntry) (config.RconEntry, error) {
	var zero config.RconEntry
	err := utils.InitEntry(&entry)
	if err != nil {
		return zero, err
	}

	validErr := utils.ValidateEntry(entry)
	if validErr != nil {
		return zero, validErr
	}

	return entry, nil
}

// getCommandString retrieves the command string from the args.
func (e *ExecCommand) getCommandString(args []string) string {
	command := strings.Join(args, " ")

	return command
}
