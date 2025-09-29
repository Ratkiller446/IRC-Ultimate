package parser

import (
        "strings"
)

type Message struct {
        Prefix  string
        Command string
        Params  []string
}

// skipSpaces skips one or more consecutive ASCII spaces as per RFC 1459
func skipSpaces(line string) string {
        for len(line) > 0 && line[0] == ' ' {
                line = line[1:]
        }
        return line
}

// findNextSpace finds the next space character, handling RFC 1459 space requirements
func findNextSpace(line string) int {
        return strings.IndexByte(line, ' ')
}

// ParseMessage parses a raw IRC message into its prefix, command, and parameters as per RFC 1459.
func ParseMessage(line string) Message {
        msg := Message{
                Params: []string{}, // Initialize with empty slice, not nil
        }
        if line == "" {
                return msg
        }
        
        // Handle prefix if present
        if line[0] == ':' {
                // Prefix present
                space := findNextSpace(line)
                if space == -1 {
                        // Malformed: colon without space - treat as regular command
                        msg.Command = line
                        return msg
                }
                msg.Prefix = line[1:space]
                line = line[space:]
                // Skip one or more spaces after prefix
                line = skipSpaces(line)
        }
        
        // Extract command
        if len(line) == 0 {
                return msg
        }
        space := findNextSpace(line)
        if space == -1 {
                msg.Command = line
                return msg
        }
        msg.Command = line[:space]
        line = line[space:]
        // Skip one or more spaces after command
        line = skipSpaces(line)
        
        // Parse parameters
        for len(line) > 0 {
                if line[0] == ':' {
                        // Trailing parameter - everything after ':'
                        msg.Params = append(msg.Params, line[1:])
                        break
                }
                space = findNextSpace(line)
                if space == -1 {
                        // Last parameter
                        if line != "" {
                                msg.Params = append(msg.Params, line)
                        }
                        break
                }
                // Middle parameter
                msg.Params = append(msg.Params, line[:space])
                line = line[space:]
                // Skip one or more spaces before next parameter
                line = skipSpaces(line)
        }
        return msg
}
