package ipcstart

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/bobllor/rcon/app/serve/internal"
	"github.com/bobllor/rcon/app/utils"
	"github.com/bobllor/rcon/app/utils/paths"
	"github.com/bobllor/rcon/config"
	"github.com/bobllor/rcon/listener"
	"github.com/bobllor/rcon/rcon"
	"github.com/spf13/cobra"
)

// IpcStartCommand is a struct used to start the IPC listener
// and connection to run RCON commands without needing
// to reauthenticate.
//
// The IpcStartCommand will run a listener and detach itself
// from the process.
//
// The process will timeout after 5 mintues by default,
// however a subcommand to force stop can be used.
//
// It does not take any commands to run. The root and run
// subcommands are responsible for running values.
type IpcStartCommand struct {
	Cmd  *cobra.Command
	Path paths.AppPath
	data IpcStartData
}

type IpcStartData struct {
	// StartService is the flag th
	StartService bool
	// Address is the socket address.
	Address string
	// Target is the target RCON entry to start the service on. This must
	// exist, and must have a valid address and password.
	Target string
	// Duration is the service uptime before it shuts down in minutes.
	// By default it will be 5 minutes via ipc.InitFlag.
	Duration int
	// PidFile is the path to the PID file for the service.
	PidFile string
}

const internalFlag = "internal-bg-service"
const usageTemplate = `Starts the RCON service in the background.

This requires an RCON entry to exist, either given with the target or 
having a valid default RCON entry.`

func NewIpcStartCommand(addr, pidFilePath string, paths paths.AppPath) *IpcStartCommand {
	cmd := &IpcStartCommand{
		Cmd: &cobra.Command{
			Use:   "start [flags]",
			Short: "Starts the RCON service in the background",
			Long:  usageTemplate,
			Args:  cobra.NoArgs,
		},
		data: IpcStartData{
			Address: addr,
			PidFile: pidFilePath,
		},
		Path: paths,
	}

	cmd.Cmd.Run = cmd.Run
	cmd.Cmd.PreRunE = cmd.PreRunE

	cmd.initFlags()

	// not meant to be used by the user. this is ran with os.Exec
	// if it is used it will return an error to the user.
	cmd.Cmd.Flag(internalFlag).Hidden = true

	return cmd
}

// Run is the main entry to the start subcommand.
func (isc *IpcStartCommand) Run(cmd *cobra.Command, args []string) {
	cfg, err := config.LoadConfigurationIfMissing(isc.Path.Config)
	if err != nil {
		utils.PrintFatal(err)
	}

	var entry *config.RconEntry
	var target string
	if cfg.DefaultRcon != "" {
		if cfg.HasEntry(cfg.DefaultRcon) {
			cfgEntry := cfg.RconEntries[cfg.DefaultRcon]
			entry = &cfgEntry

			target = cfg.DefaultRcon
		}
	} else if cfg.DefaultRcon == "" || !cfg.HasEntry(cfg.DefaultRcon) {
		fmt.Println("No default RCON entry found in config")
	}

	// overwrites default RCON
	if isc.data.Target != "" {
		if cfg.HasEntry(isc.data.Target) {
			cfgEntry := cfg.RconEntries[isc.data.Target]

			entry = &cfgEntry
			target = isc.data.Target
		} else {
			// due to this overwriting the default entry, it will exit
			// if not found.
			utils.PrintFatalf("Entry %s is not found", isc.data.Target)
		}
	}

	if entry == nil {
		utils.PrintFatalString("No entry found")
	}

	isc.runProcess(target, *entry)
}

func (isc *IpcStartCommand) PreRunE(cmd *cobra.Command, args []string) error {
	if strings.TrimSpace(isc.data.Target) == "" && cmd.Flag("target").Changed {
		return errors.New("cannot have an empty RCON target entry")
	}

	return nil
}

// runProcess is the main method of the command. It is responsible for starting the service by using
// a child process. The parent and child will communicate over a pipe.
//
// Files will be written to the runtime path directory.
//
// It will run in the background with the default command run. The flag will be used by an
// exec.Command to start the service.
//
// It will not return errors, instead it will exit if errors occur. This is due to the IPC
// processes.
func (isc *IpcStartCommand) runProcess(rconTarget string, entry config.RconEntry) {
	// NOTE: this gets most edge cases except a major one:
	//	- service running -> rcon file missing; previous service will remain open.
	isc.checkAndRemoveStaleService(isc.data.PidFile, isc.data.Address)

	if !isc.data.StartService {
		err := isc.initRun(rconTarget)
		if err != nil {
			utils.PrintFatal(err)
		}
	} else {
		err := isc.serviceRunHandler(entry)
		if err != nil {
			utils.PrintFatal(err)
		}
	}
}

// initRun is the initial run process without the internal flag. It will start the
// child process with the flag.
//
// The method will do the following:
//   - Spawn the child process
//   - Pipe creation for IPC
//   - Check the pipe for successful start
//   - Write PID to a file on the disk
//
// After the command is started, any errors will remove the created files.
func (isc *IpcStartCommand) initRun(rconTarget string) error {
	execCmd := exec.Command(
		os.Args[0],
		"serve",
		"start",
		"--"+internalFlag,
		"--duration",
		fmt.Sprintf("%d", isc.data.Duration),
		"--target",
		rconTarget,
	)
	execCmd.Stderr = os.Stderr

	// nil checks arent required due to the switch statement
	var pipeR io.ReadCloser

	switch runtime.GOOS {
	case "linux", "darwin":
		// debug only
		// execCmd.Stdout = os.Stdout

		pipe, pipeW, err := os.Pipe()
		if err != nil {
			return err
		}
		defer pipeW.Close()

		execCmd.ExtraFiles = []*os.File{pipeW}
		pipeR = pipe
	case "windows":
		pipe, err := execCmd.StdoutPipe()
		if err != nil {
			return err
		}
		pipeR = pipe

		envkeyval := fmt.Sprintf("%s=%s", internal.WindowsEnvKey, "somevalue")

		// resets the variables inside the process, this is used to ensure
		// its ran from the parent process
		execCmd.Env = append(execCmd.Env, envkeyval)
		// USERPROFILE key is missing from the env in the child process
		execCmd.Env = append(execCmd.Env, "USERPROFILE="+os.Getenv("USERPROFILE"))
	default:
		return errors.New("OS not supported")
	}

	defer pipeR.Close()

	err := execCmd.Start()
	if err != nil {
		internal.RemoveFiles(isc.data.PidFile, isc.data.Address)
		return err
	}

	// windows will print json format to Stdout
	var pipeProcess internal.PipeProcess
	err = json.NewDecoder(pipeR).Decode(&pipeProcess)
	if err != nil {
		return err
	}

	if pipeProcess.OK {
		fmt.Printf("Starting RCON service (%s: %d)\n", rconTarget, execCmd.Process.Pid)
		err := isc.initRunWritePid(execCmd)
		if err != nil {
			return err
		}
	}

	return nil
}

// initRunWritePid writes the PID file to the disk. If it fails to write to the disk,
// it will kill the process, remove the files, and return an error.
//
// This should only be called in initRun.
func (isc *IpcStartCommand) initRunWritePid(cmd *exec.Cmd) error {
	err := isc.writeFile(isc.data.PidFile, fmt.Appendf([]byte{}, "%d", cmd.Process.Pid))
	if err != nil {
		internal.RemoveFiles(isc.data.PidFile, isc.data.Address)
		cmd.Process.Kill()

		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// serviceRunHandler is used to run the service mode based on the
// OS the program is being ran from.
//
// OSes that aren't supported will return an error.
func (isc *IpcStartCommand) serviceRunHandler(entry config.RconEntry) error {
	proc := internal.PipeProcess{OK: true}
	switch runtime.GOOS {
	case "darwin", "linux":
		// this should already be created from initRun, serviceRunHandler
		// is called in runProcess from the parent
		fi := os.NewFile(uintptr(3), "pipe")

		unixpipe := internal.NewUnixPipeProcess(proc, fi)
		// will not close until serviceRun exists since it blocks
		defer fi.Close()

		return isc.serviceRun(unixpipe, entry)
	case "windows":
		winpipe := internal.NewWindowsPipeProcess(proc)

		return isc.serviceRun(winpipe, entry)
	default:
		return errors.New("OS not supported")
	}
}

// serviceRun is the method used to start the service and listen for connections.
// The connections will send command payloads to the socket in order to execute them
// with RCON. This is intended to authenticate once and send any amount of commands.
//
// This is a blocking call.
//
// In order to prevent multiple services from being ran and having the PID file
// being rewritten, a pipe is used to communicate back to the parent process.
// The pipe is created in the initRun call.
//
// The method will do the following:
//   - Set up the listener for RCON commands
//   - Write to the pipe during the listener set up
//   - Authenticate into RCON
//   - Listen for and execute commands
//   - Cleanup file removal of the socket and PID files
//
// When the function ends, it will always remove the files even on an error.
//
// This method handles both Unix and Windows OSes.
func (isc *IpcStartCommand) serviceRun(process internal.Process, entry config.RconEntry) error {
	err := process.ValidateHandshake()
	if err != nil {
		return err
	}

	serv, err := isc.newRconListener(isc.data.Address)
	if err != nil {
		process.SetError(err.Error())

		err := process.Report()
		if err != nil {
			return err
		}

		return process.ToError()
	}
	defer serv.Close()
	go serv.Stop(time.Duration(isc.data.Duration) * time.Minute)

	defer internal.RemoveFiles(isc.data.PidFile, isc.data.Address)

	rconn, err := isc.newRconClient(entry)
	if err != nil {
		process.SetError(err.Error())

		err := process.Report()
		if err != nil {
			return err
		}

		return process.ToError()
	}
	defer rconn.Close()

	// if no errors occur after the init and auth, it is considered successful.
	// write back to parent
	err = process.Report()
	if err != nil {
		return err
	}

	err = serv.HandleConnection(rconn)
	if err != nil && !errors.Is(err, net.ErrClosed) {
		return err
	}

	return nil
}

// newRconClient creates a new authenticated RCON client.
//
// It is the caller's responsibility to close the connection.
func (isc *IpcStartCommand) newRconClient(entry config.RconEntry) (*rcon.Rcon, error) {
	rconn, err := rcon.NewRcon(entry.Address)
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			return nil, fmt.Errorf("failed to start RCON service: connection refused for %s", entry.Address)
		}

		return nil, err
	}

	err = rconn.Authenticate(entry.Password)
	if err != nil {
		if errors.Is(err, rcon.ErrAuthFail) {
			return nil, errors.New("failed to authenticate RCON: password is incorrect")
		}

		return nil, err
	}

	return rconn, nil
}

// newRconListener creates a new RCON listener for the commmands.
//
// It is the caller's responsibility to close the listener.
func (isc *IpcStartCommand) newRconListener(addr string) (*listener.RconListener, error) {
	serv, err := listener.NewRconListener(addr)
	if errors.Is(err, syscall.EADDRINUSE) {
		return nil, errors.New("RCON service is already running")
	} else if err != nil {
		return nil, err
	}

	return serv, nil
}

// writeFile writes to a given file path with bytes.
//
// The file will be created if it does not exist. If it exists,
// the contents will be oevrwritten.
func (isc *IpcStartCommand) writeFile(path string, b []byte) error {
	return os.WriteFile(path, b, 0o744)
}

// checkAndRemoveStaleService checks if the service is running via reading the PID file.
// If it is not running, then it will remove the stale files on the disk.
//
// If the PID is not in use, the PID file does not exist, or an error
// it will remove the files for the service.
//
// If the PID is in use, this will do nothing.
func (isc *IpcStartCommand) checkAndRemoveStaleService(pidFile string, addr string) {
	// any errors will just return nil, including non-number errors.
	pid, err := internal.ReadPID(pidFile)
	if err != nil {
		internal.RemoveFiles(pidFile, addr)

		return
	}

	isRunning := internal.CheckProcessRunning(pid)
	if !isRunning {
		internal.RemoveFiles(pidFile, addr)

		return
	}
}

func (isc *IpcStartCommand) initFlags() {
	isc.Cmd.Flags().BoolVar(&isc.data.StartService, internalFlag, false, "Runs the IPC service in the background")
	isc.Cmd.Flags().IntVar(&isc.data.Duration, "duration", 5, "The service duration uptime in minutes")

	isc.Cmd.Flags().StringVarP(&isc.data.Target, "target", "t", "", "The target RCON entry to start the service on")
}
