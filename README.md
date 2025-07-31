# IRC-Ultimate

A minimal, modular IRC client with kawaii ASCII art, following the UNIX philosophy and RFC 1459. (◕‿◕✿)

## Project Layout

This project uses a modular Go layout:

- `main.go` — Main CLI entry point
- `conn/` — Connection logic (TLS, TCP, config)
- `parser/` — IRC message parser and tests
- `asciiart/` — Kawaii ASCII art and display utilities
- `go.mod` — Go module definition

**All commands below assume you are in the `irc` directory.**

## ✨ Features

- 🐱 Kawaii ASCII art on startup
- 🔒 Secure by default (TLS support)
- 🚀 Fast and lightweight
- 📦 No external dependencies
- 🎨 Customizable kawaii art collection

## 🛠️ Installation

### Prerequisites
- Go 1.16 or later
- Git

### Build from source

```bash
git clone https://github.com/yourusername/IRC-Ultimate.git
cd IRC-Ultimate
go build -o irc-client
```

## 🚀 Usage

### Basic Usage
```bash
# Run with default settings
./irc-client

# Connect to a specific server
./irc-client -server irc.libera.chat -port 6697 -tls -nick YourNickname

# Disable TLS (not recommended)
./irc-client -tls=false -port 6667
```

### Commands
- `/join #channel` - Join a channel
- `/msg <target> <message>` - Send a private message
- `/part [message]` - Leave the current channel
- `/nick <newnick>` - Change your nickname
- `/quit [message]` - Disconnect from the server
- Type a message to send to the current channel

### Kawaii Features
- Enjoy random kawaii ASCII art on startup
- Customize your experience with different art styles
- All kawaii art is terminal-friendly and works in most environments

## 🐛 Troubleshooting

### Terminal Display Issues
If you see broken characters in the kawaii art:
1. **Windows Users**: Use [Windows Terminal](https://aka.ms/terminal) for better Unicode support
2. **Font Issues**: Install a Nerd Font (like [Cascadia Code](https://www.nerdfonts.com/font-downloads))
3. **Terminal Settings**: Ensure your terminal is set to UTF-8 encoding

### Common Problems
- **Missing packages**: Run `go mod tidy`
- **Build errors**: Try `go clean -modcache` and rebuild
- **Connection issues**: Check your network and firewall settings

## Usage
- Type `/join #channel` to join a channel
- Type `/msg <target> <message>` to send a private message
- Type `/part` to leave the current channel
- Type `/nick <newnick>` to change nickname
- Type `/quit` to disconnect
- Type a message to send to the current channel

## 📦 Project Structure

```
.
├── asciiart/      # Kawaii ASCII art and display utilities
│   └── art/       # Art collections (cats, faces, etc.)
├── conn/          # Connection handling (TLS/TCP)
├── parser/        # IRC message parser
├── main.go        # Main application entry point
└── README.md      # This file
```

## 🤝 Contributing

Contributions are welcome! Feel free to submit issues and pull requests.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the BSD 2-Clause License - see the [LICENSE](LICENSE) file for details.

[![License](https://img.shields.io/badge/License-BSD_2--Clause-orange.svg)](https://opensource.org/licenses/BSD-2-Clause)

## 🙏 Acknowledgments

- Inspired by the simplicity of UNIX philosophy
- Kawaii art by the community
- Built with ❤️ and Go

---

<div align="center">
  Made with (◕‿◕✿) and Go
</div>


