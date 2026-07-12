package paths

type AppPath struct {
	// Home is the default home path of the logged in user on the device.
	Home string
	// Config is the path where the configuration file is stored in.
	Config string
	// Log is the path where logging is stored in.
	Log string
	// Runtime is the path for files created during the program's runtime.
	Runtime string
}
