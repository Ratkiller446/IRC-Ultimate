package commands

import (
        "strings"
)

// Sanitize removes carriage returns, newlines, and null bytes from input strings
// This ensures IRC protocol compliance and prevents injection attacks
func Sanitize(s string) string {
        s = strings.ReplaceAll(s, "\r", "")
        s = strings.ReplaceAll(s, "\n", "")
        s = strings.ReplaceAll(s, "\x00", "")
        return s
}

// ClientState represents the state needed for IRC command processing
type ClientState struct {
        CurrentChannel string
        Nick           string
}

// BuildIRCCommand converts user input into IRC protocol commands
// Returns a slice of IRC commands to send to the server
// Handles all standard IRC commands: JOIN, PART, NICK, PRIVMSG, QUIT
// Returns empty slice for invalid or unhandled commands
//
// NOTE: This function now uses the new router internally while maintaining
// backward compatibility with the existing interface.
func BuildIRCCommand(state ClientState, userInput string) ([]string, error) {
        // Use the new router system internally
        commands, err := ConvertToNewRouter(state, userInput)
        if err != nil {
                return nil, err
        }
        
        // Convert back to raw string format for backward compatibility
        rawCommands := make([]string, len(commands))
        for i, cmd := range commands {
                rawCommands[i] = cmd.Raw
        }
        
        return rawCommands, nil
}