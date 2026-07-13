# RCON

A Go-based CLI and persistent service for managing servers using the RCON protocol.

# Features

- Persistent background service that reuses an authenticated RCON connection
- IPC communication for command execution over sockets
- Supports named server profiles
- Send commands to servers by targeting a server profile name
- Interactive configuration

# Current Limitations

- RCON passwords are stored in plain text
- Windows is currently not supported (as of 7-13-2026)
- Multi-packet responses are not currently supported