# IRC Ultimate Client - Project Documentation

## Overview
IRC Ultimate is a minimal, modular IRC client with kawaii ASCII art, built in Go following the UNIX philosophy and RFC 1459. The project has been successfully imported and configured to run in the Replit environment.

## Recent Changes (September 29, 2025)
- Successfully imported from GitHub repository
- Installed Go toolchain and verified compatibility (Go 1.21.13)
- Built the IRC client application successfully (./irc-client binary created)
- Configured console workflow for interactive terminal application
- Set up VM deployment configuration for production use
- Application tested and running properly in development environment

## Project Architecture
This is a Go-based console application with the following modular structure:

### Core Modules:
- **main.go** - Main CLI entry point with kawaii ASCII art display
- **conn/** - Connection logic handling TLS/TCP connections and configuration
- **parser/** - IRC message parser implementing RFC 1459 protocol
- **asciiart/** - Kawaii ASCII art and display utilities
  - **art/** - Art collections (cats, faces, etc.)

### Key Features:
- 🐱 Kawaii ASCII art on startup
- 🔒 Secure by default (TLS support) 
- 🚀 Fast and lightweight
- 📦 No external dependencies
- 🎨 Customizable kawaii art collection

## Development Environment
- **Language**: Go 1.21.13
- **Build Command**: `go build -o irc-client`
- **Run Command**: `./irc-client`
- **Workflow**: Console application (interactive terminal)
- **Port Usage**: None (IRC client connects outbound to IRC servers)

## Deployment Configuration
- **Target**: VM (maintains state, always running)
- **Run Command**: `./irc-client`
- **Type**: Console/Terminal Application

## User Commands
Once connected to an IRC server, users can:
- `/join #channel` - Join a channel
- `/msg <target> <message>` - Send private message
- `/part` - Leave current channel  
- `/nick <newnick>` - Change nickname
- `/quit` - Disconnect from server
- Type messages to send to current channel

## Connection Configuration
Default settings:
- Server: irc.libera.chat (or user specified)
- Port: 6697 (TLS) / 6667 (plain)
- TLS: Enabled by default
- Nickname: System username or user specified

## Testing Status
✅ Application builds successfully
✅ Workflow configured and running
✅ Console output displaying correctly
✅ Interactive prompts working
✅ Deployment configuration set