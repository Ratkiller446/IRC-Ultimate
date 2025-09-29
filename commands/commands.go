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
func BuildIRCCommand(state ClientState, userInput string) ([]string, error) {
	// Sanitize input to prevent protocol injection
	userInput = Sanitize(userInput)
	
	// Handle empty input
	if userInput == "" {
		return []string{}, nil
	}
	
	// Check if it's a command (starts with /)
	if len(userInput) > 0 && userInput[0] == '/' {
		// Parse command and arguments
		fields := strings.Fields(userInput)
		if len(fields) == 0 {
			return []string{}, nil
		}
		
		// Extract command (remove the leading /)
		if len(fields[0]) <= 1 {
			return []string{}, nil // Just a slash
		}
		
		cmd := strings.ToLower(fields[0][1:])
		
		switch cmd {
		case "join":
			if len(fields) > 1 {
				// JOIN command can have optional key: JOIN #channel [key]
				if len(fields) > 2 {
					return []string{"JOIN " + fields[1] + " " + fields[2]}, nil
				}
				return []string{"JOIN " + fields[1]}, nil
			}
			return []string{}, nil
			
		case "part":
			if state.CurrentChannel != "" {
				// PART can have optional reason
				if len(fields) > 1 {
					reason := strings.Join(fields[1:], " ")
					return []string{"PART " + state.CurrentChannel + " :" + reason}, nil
				}
				return []string{"PART " + state.CurrentChannel}, nil
			}
			return []string{}, nil
			
		case "nick":
			if len(fields) > 1 {
				return []string{"NICK " + fields[1]}, nil
			}
			return []string{}, nil
			
		case "msg":
			if len(fields) > 2 {
				target := fields[1]
				message := strings.Join(fields[2:], " ")
				return []string{"PRIVMSG " + target + " :" + message}, nil
			}
			return []string{}, nil
			
		case "quit":
			if len(fields) > 1 {
				reason := strings.Join(fields[1:], " ")
				return []string{"QUIT :" + reason}, nil
			}
			return []string{"QUIT"}, nil
			
		default:
			// Unknown command
			return []string{}, nil
		}
	} else {
		// Not a command, treat as channel message
		if state.CurrentChannel != "" {
			return []string{"PRIVMSG " + state.CurrentChannel + " :" + userInput}, nil
		}
		return []string{}, nil
	}
}