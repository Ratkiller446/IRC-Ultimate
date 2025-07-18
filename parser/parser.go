package parser

import (
	"strings"
)

type Message struct {
	Prefix  string
	Command string
	Params  []string
}

// ParseMessage parses a raw IRC message into its prefix, command, and parameters as per RFC 1459.
func ParseMessage(line string) Message {
	msg := Message{}
	if line == "" {
		return msg
	}
	if line[0] == ':' {
		// Prefix present
		space := strings.IndexByte(line, ' ')
		if space == -1 {
			return msg
		}
		msg.Prefix = line[1:space]
		line = line[space+1:]
	}
	// Command
	space := strings.IndexByte(line, ' ')
	if space == -1 {
		msg.Command = line
		return msg
	}
	msg.Command = line[:space]
	line = line[space+1:]
	// Params
	params := []string{}
	for len(line) > 0 {
		if line[0] == ':' {
			params = append(params, line[1:])
			break
		}
		space = strings.IndexByte(line, ' ')
		if space == -1 {
			params = append(params, line)
			break
		}
		params = append(params, line[:space])
		line = line[space+1:]
	}
	msg.Params = params
	return msg
} 