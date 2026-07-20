<div align="center"> 
    <h1>GO RCON CLI</h1>
</div>

# Overview

A cross-platform CLI application for executing commands over the RCON protocol, featuring a persistent background
RCON client that maintains an authenticated connection between consecutive commands.

It features:
- A direct mode for one-off command execution
- A service mode for a persistent RCON client listening in the background for commands
- IPC communication for command execution over a persistent RCON client
- Supports named server profiles
- Execute commands to servers by targeting a server profile name
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
    - [Linux/macOS](#linuxmacos)
    - [Windows](#windows)
    - [Manual](#manual)
- [Uninstall](#uninstall)
    - [Linux/macOS](#uninstall-linuxmacos)
    - [Windows](#uninstall-windows)
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
- [Development](#development)
    - [Requirements](#requirements)
    - [Getting started](#getting-started)
- [Current Limitations](#current-limitations)

# Installation

Installation scripts are used to automatically install and setup the
application based on your OS. There are certain software requirements that are
required for the installation script:
- Linux/macOS: `curl`, `tar`
- Windows: `curl`/`Invoke-WebRequest`

On modern OSes, these will already be included.
Simply copy and paste the commands below to run the installation script.

## Linux/macOS

```bash
bash <(curl -s https://raw.githubusercontent.com/bobllor/rcon/refs/heads/main/install.sh)
```

## Windows

```powershell
# WIP
```

## Manual

If manual is preferred, you can:
1. Download and extract the latest binaries related to your OS
2. Clone the repository and run the build commands manually (requires Go >= 1.24)
3. Clone the repository and run the `install.sh`/`install.ps1` script

Usage of `gorcon` is expected to be *through a terminal session*. How to use it is
dependent on your preference, such as setting the PATH variable or storing
it in a folder to use.
- It is recommended to use the installation scripts as this is handled for you

# Uninstall

Similar to the installation scripts, there is an *automatic uninstall script* that
will be used. The same requirements are needed to run the commands below:
- Linux/macOS: `curl`, `tar`
- Windows: `curl`/`Invoke-WebRequest`

Simply copy and paste the command in your respective terminal.

## Uninstall Linux/macOS

WIP

## Uninstall Windows

WIP


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
- The name `default` is reserved, and *cannot be used* as a profile name

*Named conflicts* can occur if attempting to add an existing profile entry. By default, it
will *prevent duplicate profiles* from being added.
This can be bypassed with the `--overwrite` flag, which will replace the original profile
in the configuration.

> `--overwrite` is a destructive action and will permanently remove the original profile.

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

Only one edit can occur at a time. Additionally, there must be *at least one flag
used* with the command.
If *interactive mode* is preferred, pass in the string `-` to a flag value to trigger
the mode.
- This is recommended to be used for changing passwords

There is a subcommand `gorcon edit default` that works similiar to `gorcon edit`, except this
command *automatically targets the default profile*. There must be a default profile entry for this to be used.
- Useful for if you need to modify the default RCON profile without having to remember the name

> If you are changing the RCON entry profile name and it is also the default RCON profile,
> *it will automatically update the default profile to the new name chosen*.
>
> For example, `defaultRCON = defaultName` -> `newName -> aNewName` -> `defaultRCON = aNewName`.

For more flags, run `gorcon edit -h`.

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

The flags `-t/--target` and any of the two `-a/--address`/`-p/--password` cannot be used together.

```bash
# default usage, will always send to default profile otherwise interactive mode is started
gorcon exec say hello world!

# target server2 for the command
gorcon exec deop Notch -t server2

# overwrites the default values and enters interactive mode for the password
gorcon exec op Jeb -a 10.0.0.0:12345

# the flags -t and -p are not allowed to be ran together. this also includes -t and -a
gorcon exec say wowzers -p somepasswordhere -t server01 # ERROR
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

# Development

Due to the program being *cross-platform*, it is expected that both Windows and a Unix environment
are used.

## Requirements

Required:
- Go >= 1.24
- Docker >= 28.3.2
- Bash / Git Bash

Optional (only for the scripts):
- curl
- tar
- zip

## Getting Started

Clone the repository:

```bash
git clone https://github.com/bobllor/RCON-CLI
```

The structure of the project and what each directory is used for:

```
.
└── root/
    ├── app (CLI)
    ├── config (program configuration logic)
    ├── docker (vanilla server setup for integration testing)
    ├── listener (daemon service)
    ├── packet (RCON packets)
    └── rcon (RCON communication) 
```

# Current Limitations

- RCON passwords are stored in plain text
- Multi-packet responses are not currently supported
- macOS is not tested, however since it shares a Unix environment it should work
- Service mode only supports one instance at a time