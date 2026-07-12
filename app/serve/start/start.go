package ipcstart

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/bobllor/rcon/app/serve/internal"
	"github.com/bobllor/rcon/app/utils"
	"github.com/bobllor/rcon/app/utils/paths"
	"github.com/bobllor/rcon/config"
	"github.com/bobllor/rcon/rcon"
	"github.com/bobllor/rcon/service"
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

func NewIpcStartCommand(addr, pidFile string, paths paths.AppPath) *IpcStartCommand {
	cmd := &IpcStartCommand{
		Cmd: &cobra.Command{
			Use:   "start",
			Short: "Starts the IPC RCON service",
			Args:  cobra.NoArgs,
		},
		data: IpcStartData{
			Address: addr,
			PidFile: pidFile,
		},
		Path: paths,
	}

	cmd.Cmd.Run = cmd.Run
	cmd.initFlags()

	// not meant to be used by the user. this is ran with os.Exec
	// if it is used it will return an error to the user.
	cmd.Cmd.Flag(internalFlag).Hidden = true

	return cmd
}

func (isc *IpcStartCommand) Run(cmd *cobra.Command, args []string) {
	cfg, err := config.LoadConfigurationIfMissing(isc.Path.Config)
	if err != nil {
		utils.PrintFatal(err)
	}
	var entry config.RconEntry
	if cfg.DefaultRcon != "" {
		entry = cfg.RconEntries[cfg.DefaultRcon]
	}

	if isc.data.Target != "" {

	}

	isc.runProcess(entry)
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
func (isc *IpcStartCommand) runProcess(entry config.RconEntry) {
	// NOTE: this gets most edge cases except a major one:
	//	- service running -> rcon file missing; previous service will remain open.
	isc.checkAndRemoveStaleService(isc.data.PidFile, isc.data.Address)

	if !isc.data.StartService {
		err := isc.initRun()
		if err != nil {
			utils.PrintFatal(err)
		}
	} else {
		err := isc.serviceRun(entry)
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
func (isc *IpcStartCommand) initRun() error {
	execCmd := exec.Command(os.Args[0], "serve", "start", "--"+internalFlag, "--duration", fmt.Sprintf("%d", isc.data.Duration))
	execCmd.Stderr = os.Stderr
	// only used for debugging
	execCmd.Stdout = os.Stdout

	pipeR, pipeW, err := os.Pipe()
	if err != nil {
		return err
	}
	defer pipeW.Close()
	defer pipeR.Close()

	execCmd.ExtraFiles = []*os.File{pipeW}

	err = execCmd.Start()
	if err != nil {
		internal.RemoveFiles(isc.data.PidFile, isc.data.Address)
		return err
	}

	var data internal.PipeProcessError
	// io.ReadAll/ReadFull does not work, it blocks the terminal
	err = json.NewDecoder(pipeR).Decode(&data)
	if err != nil {
		return err
	}

	if data.OK {
		fmt.Printf("Starting RCON service (%d)\n", execCmd.Process.Pid)
		err := isc.initWritePid(execCmd)
		if err != nil {
			return err
		}
	}

	return nil
}

// initWritePid writes the PID file to the disk. If it fails to write to the disk,
// it will kill the process, remove the files, and return an error.
func (isc *IpcStartCommand) initWritePid(cmd *exec.Cmd) error {
	err := isc.writeFile(isc.data.PidFile, fmt.Appendf([]byte{}, "%d", cmd.Process.Pid))
	if err != nil {
		internal.RemoveFiles(isc.data.PidFile, isc.data.Address)
		cmd.Process.Kill()

		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
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
func (isc *IpcStartCommand) serviceRun(entry config.RconEntry) error {
	// used to communicate to the parent call
	processErr := internal.PipeProcessError{OK: true}

	fi := os.NewFile(uintptr(3), "pipe")
	defer fi.Close()

	_, err := fi.Stat()
	if err != nil {
		processErr.SetError("Must run mcrcon serve start")

		return processErr.ToError()
	}

	serv, err := service.NewRconListener(isc.data.Address)
	if errors.Is(err, syscall.EADDRINUSE) {
		processErr.SetError("RCON service is already running")
		processErr.Encode(fi)

		return processErr.ToError()
	} else if err != nil {
		processErr.SetError(err.Error())
		processErr.Encode(fi)

		return err
	}
	defer serv.Close()
	go serv.Stop(time.Duration(isc.data.Duration) * time.Minute)

	defer internal.RemoveFiles(isc.data.PidFile, isc.data.Address)

	rconn, err := rcon.NewRcon(entry.Address)
	if err != nil {
		processErr.SetError(err.Error())
		if errors.Is(err, syscall.ECONNREFUSED) {
			processErr.SetErrorf("Failed to start RCON service: connection refused for %s", entry.Address)
		}

		processErr.Encode(fi)

		return processErr.ToError()
	}
	defer rconn.Close()

	err = rconn.Authenticate(entry.Password)
	if err != nil {
		processErr.SetError(err.Error())
		errReturn := errors.New(err.Error())

		if errors.Is(err, rcon.ErrAuthFail) {
			processErr.SetErrorf("Failed to authenticate RCON: password is incorrect (%s)", entry.Address)
			errReturn = processErr.ToError()
		}

		processErr.Encode(fi)

		return errReturn
	}

	// if no errors occur after the init and auth, it is considered successful.
	// write back to parent
	processErr.Encode(fi)

	err = serv.HandleConnection(rconn)
	if err != nil && !errors.Is(err, net.ErrClosed) {
		return err
	}

	return nil
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
}
