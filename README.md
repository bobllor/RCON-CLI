# RCON

A Go-based CLI for executing commands over the RCON protocol, featuring a persistent background
RCON client that maintains an authenticated connection between consecutive commands.

It features:
- A direct mode for one-off command execution
- A service mode for a persistent RCON client listening in the background for commands
- IPC communication for command execution over sockets
- Supports named server profiles
- Send commands to servers by targeting a server profile name
- Interactive configuration
- Linux and Windows support

To jump straight to how to execute commands for both modes:
- [Direct mode](#executing-direct-commands)
- [Service mode](#executing-service-mode-commands)


# What is Service Mode?

Service mode creates an authenticated RCON client in the background. This client
listens for incoming commands to send to the server.

The mode addresses an issue that direct mode does not do: *authenticating once and reusing the connection*.
This removes the TCP handshake overhead for sending each command due to reauthentication.

This is primarily used for sending consecutive commands within a timeframe, but for nearly all use cases
the direct mode should be the *preferred option*.


# Before You Start

Ensure RCON is enabled for the server the commands are being sent to. 
How to enable RCON varies from game to game. 

In Minecraft's case, the following values must be set in `server.properties`:

```properties
enable-rcon=true
rcon.port=YOUR_PORT_HERE
rcon.password=YOUR_PASSWORD_HERE
```

The RCON port *cannot be the same port* as the *server port*. 
It is *expected* that the RCON protocol is done through a *local network*.
It is *not recommended* to expose RCON to the internet due to RCON being a plaintext protocol.
- In other words, all data sent over RCON is unencrypted by default


## Named Server Profiles

`rcon` supports storing named server profiles in its configurations. 
This is used to automate the authentication process, cache multiple servers, and target
specific servers for commands.
- The profile names *must be unique*
- The profile names are *case sensitive*
- Spaces are allowed

A special profile is considered a `default` profile, which the `rcon` command
will automatically target this profile with commands without *specifying the target server*.
- This must be defined with `rcon edit` or `rcon add` to set a default profile

The profile is structured into three parts:
1. The unqiue name identifier
2. The RCON address, which *must be in the format `address:port`*
3. The RCON password


### Adding Profiles

To add a profile:

```bash
# interactive: profile name, RCON address, and RCON password
rcon add

# interactive: RCON address and RCON password.
rcon add -n MyNewServer

# interactive: RCON password
rcon add MyNewServer -a 127.0.0.1:50123

# adds to the entry with no interactive mode
rcon add MyNewServer -a 127:0.0.01:50123 -p ExamplePassword

# interactive: RCON address and RCON password. the new entry is set as the default profile
rcon add MyNewServer --default
```

Some caveats:
- Arguments and the `-n/--name` flag cannot be used together 
- The *password cannot be empty*, but interactive mode can be triggered by 
not using `-p/--password` or passing the string `-`

*Named conflicts* can occur if attempting to add an existing profile entry. By default, it
will *prevent duplicate profiles* from being added.
This can be bypassed with the `--overwrite` flag, which will replace the original profile
in the configuration.
- This will do nothing if the new profiles are unique

> This is a destructive action and will permanently remove the original profile.

```bash
# upon success it will overwrite the original entry if it exists
rcon add NamedServerConflict --overwrite
```


### Editing Profiles

The following values are allowed for profile editing:
- Name of the entry (must be unique)
- Setting a profile as default
- RCON address
- RCON password

Basic use: `rcon edit <entry> [flags...]`

The flag `--rm-default` is used to remove the current default profile. This must be used
without any arguments, otherwise an error will occur.

For more flags, run `rcon edit -h`.

At least one flag must be used, otherwise an error will occur. 
*Interactive mode is not available* for editing profiles.

```bash
# name MyServer -> YourServer
rcon edit MyServer -n YourServer

# new address, password, and new default profile
rcon edit Server01 -a 127.0.0.1:28383 -p NewPassword --default

# removes the current default profile
rcon edit --rm-default
```

Similar to adding an entry, names *must be unique*. If a new name conflicts with
an existing one, it will abort the operation.


### Removing Profiles

To remove a profile:

```bash
# removes MyServer1
rcon rm MyServer1

# removes all servers
rcon rm MyServer2 "A Server Here" bigdawg
```


# Usage

There are two mode types used with `rcon`:
1. Direct mode: Authenticates and sends a command once, before exiting
2. Service mode: Spawns a background process that authenticates once and listens for commands for a set duration

*Direct mode* is the default mode used with the root command.

For full list of commands, run `rcon -h`.


## Direct Mode

### Executing Direct Commands

By default if there is no *default server profile*, it will go interactive for the address and password
to the server, depending on which value wasn't given.
- The flags `-a/--address` and `-p/--password` are used to pass the RCON info in
- The target flag `-t/--target` can be used to target server profiles, bypassing the default profile requirement

To execute a command to a server, arguments can be given or a string argument is given
with the `-c/--command` flag. This is ran with the root command `rcon`.
- Using both arguments and the flag is not allowed

```bash
# arg based usage
rcon say hello world!

# flag based usage
rcon -c "say hello world!"

# target server2 for the command
rcon -t server2 deop Notch
```

If the command has a response, it will display the response to the terminal.


## Service Mode

By default the server will run for *5 minutes* if no duration is given. This can be changed
with the `--duration` flag, which takes any valid integer.
- The value of `--duration` is in minutes

Before the service mode can be ran, a *server profile is required*. By default, it will use
the *default server profile*, but can use a different server profile with the `-t/--target` option.


### Starting the Service

To start the service:

```bash
# starts the service for the default profile for 5 minutes
rcon serve start

# starts the service for the default profile for 1 minute
rcon serve start --duration 1

# starts the service for the profile "Server05"
rcon serve start -t Server05
```

There may be a case where the server fails to start. A built-in command can be used
to clear up the service cache to get the service to run again:

```bash
# clears the service cache
rcon serve --clean
```


### Stopping the Service

To force stop the service, there are two ways to do so (instead of natural expiration):

```bash
# built-in command
rcon serve stop

# normal unix command
kill -15 PID_HERE
kill -9 PID_HERE
```

It is recommended to use the built-in `rcon serve stop`, as it has clean up functionality
in the removal of the service.


### Executing Service Mode Commands

Unlike the *direct mode command execution*, service mode uses `rcon serve exec` to run commands.
Similar to the direct mode, it takes *any number of arguments* as the command string, but does
not support any flags.

```bash
# send a command to an existing service
rcon serve exec say hello world!
rcon serve exec op Notch
```


# Current Limitations

- RCON passwords are stored in plain text
- Multi-packet responses are not currently supported
- MacOS is not tested, however since it shares a Unix environment it should work
- Service mode only supports one instance at a time