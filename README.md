# IRC Client (Go)

A minimal, modular IRC client for UNIX-like systems, following the UNIX philosophy and RFC 1459.

## Project Layout

This project uses a modular Go layout:

- `main.go` — Main CLI entry point
- `conn/` — Connection logic (TLS, TCP, config)
- `parser/` — IRC message parser and tests
- `go.mod` — Go module definition

**All commands below assume you are in the `irc` directory.**

## Build

```sh
cd IRC-ULTIMATE

go build -o irc-client
```

## Run

```sh
./irc-client
```

You will be prompted for your nickname and server. By default, the client connects with SSL (port 6697). Use `--insecure` to disable TLS verification, or `--tls=false` and `--port=6667` for plain connections.

## Troubleshooting
- If you see errors about missing packages or modules, make sure you are running all commands from the `irc` directory (where `go.mod` is).
- If you move files or rename directories, run `go mod tidy` to update dependencies.
- If Go still complains, try `go clean -modcache` and rebuild.

## Usage
- Type `/join #channel` to join a channel
- Type `/msg <target> <message>` to send a private message
- Type `/part` to leave the current channel
- Type `/nick <newnick>` to change nickname
- Type `/quit` to disconnect
- Type a message to send to the current channel

## Features
- Modular Go packages
- Plain text I/O for UNIX pipelines
- No external dependencies
- Secure by default

See `irc_client_rules.markdown` for full requirements.
