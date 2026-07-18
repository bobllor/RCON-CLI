# GO RCON CLI

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

<details>
    <summary>Before you start</summary>

Ensure RCON is enabled for the server the commands are being sent to. 
How to enable RCON varies from game to game. 

In Minecraft's case, the following values must be set in `server.properties`:

```properties
enable-rcon=true
rcon.port=YOUR_PORT_HERE
rcon.password=YOUR_PASSWORD_HERE
```

> WARNING
>
> Do not expose the RCON port unless you know what you are doing.

</details>

<details>
    <summary>What is Service Mode?</summary>

Service mode creates an *authenticated RCON client* in the background. This client
listens for incoming commands to send to the server.

The mode addresses an issue that direct mode does not do: *authenticating once and reusing the connection*.
This removes the TCP handshake overhead for sending each command due to reauthentication.

This is primarily used for sending consecutive commands within a timeframe, but for nearly all use cases
the direct mode should be the *preferred option*.

</details>

<details>
    <summary>Quickstart</summary>

```bash
# direct mode
gorcon exec say Hello world!
gorcon exec op Notch
gorcon exec ban Notch

# service mode
gorcon serve exec say sorry!
gorcon serve exec unban Notch
```
</details>

# Table of contents

- [Installation](#installation)
- [Profiles](#profiles)
    - [Named Server Profiles](#named-server-profiles)
    - [Adding Profiles](#adding-profiles)
    - [Editing Profiles](#editing-profiles)
    - [Removing Profiles](#removing-profiles)
- [Usage](#usage)
    - [Direct Mode](#direct-mode)
        - [Executing Direct Commands](#executing-direct-commands)
    - [Service Mode](#service-mode)
        - [Starting the Service](#starting-the-service)
        - [Stopping the Service](#stopping-the-service)
        - [Executing Service Commands](#executing-service-commands)
- [Current Limitations](#current-limitations)

# Installation

There are two scripts to install, a *Shell* and a *PowerShell* script. It is *recommended to use
scripts* to install as it will *initialize the PATH variables* in order to use them in
the terminal.

For ***Linux and macOS*** installation:
`bash <(curl -s https://raw.githubusercontent.com/bobllor/rcon/refs/heads/main/install.sh)`

For ***Windows*** installation: `WIP`

If manual is preferred, cloning the repo and running the commands based on your OS (requires Go >= 1.24).
- Must run `go mod download` to get all dependencies prior to building
- Must setup PATH variables manually (or this can be ignored)

```bash
# linux
GOOS=linux GOARCH=amd64 go build -o build/linux/gorcon \ 
    -ldflags="-X 'github.com/bobllor/rcon-cli/app/root.ProgramVersion=$(git tag | tail -1)'"

# macOS
GOOS=darwin GOARCH=arm64 go build -o build/darwin/gorcon \ 
    -ldflags="-X 'github.com/bobllor/rcon-cli/app/root.ProgramVersion=$(git tag | tail -1)'"

# windows
GOOS=windows GOARCH=amd64 go build -o build/windows/gorcon.exe \ 
    -ldflags="-X 'github.com/bobllor/rcon-cli/app/root.ProgramVersion=$(git tag | tail -1)'"
```


# Profiles

## Named Server Profiles

`gorcon` supports storing named server profiles in its configurations. 
This is used to automate the authentication process, cache multiple servers, and target
specific servers for commands.
- The profile names *must be unique*
- The profile names are *case sensitive*
- Spaces are allowed

A special profile is considered a `default` profile, which the `gorcon` command
will automatically target this profile with commands without *specifying the target server*.
- This must be defined with `gorcon edit` or `gorcon add` to set a default profile

The profile is structured into three parts:
1. The unique name identifier
2. The RCON address, which *must be in the format `address:port`*
3. The RCON password

Viewing profiles can be done with the command: `gorcon list` or `gorcon ls`.


## Adding Profiles

To add a profile:

```bash
# interactive: profile name, RCON address, and RCON password
gorcon add

# interactive: RCON address and RCON password.
gorcon add -n MyNewServer

# interactive: RCON password
gorcon add MyNewServer -a 127.0.0.1:50123

# adds to the entry with no interactive mode
gorcon add MyNewServer -a 127.0.0.01:50123 -p ExamplePassword

# interactive: RCON address and RCON password. the new entry is set as the default profile
gorcon add MyNewServer --default
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
gorcon add NamedServerConflict --overwrite
```


## Editing Profiles

The following values are allowed for profile editing:
- Name of the entry (must be unique)
- Setting a profile as default
- RCON address
- RCON password

Basic use: `gorcon edit <entry> [flags...]`

The flag `--rm-default` is used to remove the current default profile. This must be used
without any arguments, otherwise an error will occur.

For more flags, run `gorcon edit -h`.

At least one flag must be used, otherwise an error will occur. 
*Interactive mode is not available* for editing profiles.

```bash
# name MyServer -> YourServer
gorcon edit MyServer -n YourServer

# new address, password, and new default profile
gorcon edit Server01 -a 127.0.0.1:28383 -p NewPassword --default

# removes the current default profile
gorcon edit --rm-default
```

Similar to adding an entry, names *must be unique*. If a new name conflicts with
an existing one, it will abort the operation.


## Removing Profiles

To remove a profile:

```bash
# removes MyServer1
gorcon rm MyServer1

# removes all servers
gorcon rm MyServer2 "A Server Here" bigdawg
```


# Usage

There are two mode types:
1. Direct mode, which authenticates and sends a command and exits
2. Service mode, which spawns a background process that authenticates once and listens for commands for a set duration

*Direct mode* is the default mode used with the root command.

For full list of commands, run `gorcon -h`.


## Direct Mode

### Executing Direct Commands

To run a command: `gorcon exec <command>...`

By default, if there is no *default server profile*, it will enter an interactive mode for the 
address and password to the server, depending on which value wasn't given.

To send a command to a non-default profile entry, there are two ways:
1. Using either flag `-a/--address` and `-p/--password` for the RCON info. The default values
will not be used if at least one flag is given
2. The target flag `-t/--target` to target use the information from a specific target profile

> The flags `-t/--target` and any of the two `-a/--address`/`-p/--password` cannot be used together.

```bash
# default usage, will always send to default profile otherwise interactive mode is started
gorcon exec say hello world!

# target server2 for the command
gorcon exec deop Notch -t server2

# overwrites the default values and enters interactive mode for the password
gorcon exec op Jeb -a 10.0.0.0:12345

# the flags -t and -p are not allowed to be ran together. this also includes -t and -a
gorcon exec say wowzers -p somepasswordhere -t server01
```

If the *command has an output response*, it displays the response to the terminal.


## Service Mode

By default the server will run for *5 minutes* if no duration is given. This can be changed
with the `--duration` flag, which takes any valid integer.
- The value of `--duration` is in minutes

Before the service mode can be run, a *server profile is required*. By default, it will use
the *default server profile*, but can use a different server profile with the `-t/--target` option.


### Starting the Service

To start the service:

```bash
# starts the service for the default profile for 5 minutes
gorcon serve start

# starts the service for the default profile for 1 minute
gorcon serve start --duration 1

# starts the service for the profile "Server05"
gorcon serve start -t Server05
```

There may be a case where the server fails to start. A built-in command can be used
to clear up the service cache to get the service to run again:

```bash
# clears the service cache
gorcon serve --clean
```


### Stopping the Service

To force stop the service, there are two ways to do so (instead of natural expiration):

```bash
# built-in command
gorcon serve stop

# normal unix command, should not be preferred but can be a last resort
kill -15 PID_HERE
kill -9 PID_HERE
```

It is recommended to use the built-in `gorcon serve stop`, as it runs a clean up
in the removal of the service.


### Executing Service Commands

Unlike the *direct mode command execution*, service mode uses `gorcon serve exec` to run commands.
Similar to the direct mode, it takes *any number of arguments* as the command string, but does
not support any flags.

```bash
# send a command to an existing service
gorcon serve exec say hello world!
gorcon serve exec op Notch
```


# Current Limitations

- RCON passwords are stored in plain text
- Multi-packet responses are not currently supported
- macOS is not tested, however since it shares a Unix environment it should work
- Service mode only supports one instance at a time