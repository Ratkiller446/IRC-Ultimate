package commands

import (
	"fmt"
	"strings"

	"irc-client/internal/app"
)

// Router implements app.CommandRouter interface
type Router struct {
	handlers map[string]app.CommandHandler
}

// NewRouter creates a new command router with default handlers
func NewRouter() *Router {
	router := &Router{
		handlers: make(map[string]app.CommandHandler),
	}
	
	// Register default handlers
	router.RegisterHandler("join", &JoinHandler{})
	router.RegisterHandler("part", &PartHandler{})
	router.RegisterHandler("nick", &NickHandler{})
	router.RegisterHandler("msg", &MsgHandler{})
	router.RegisterHandler("quit", &QuitHandler{})
	
	return router
}

// Route processes user input and returns IRC commands to send
func (r *Router) Route(state app.ClientState, input string) ([]app.Command, error) {
	// Sanitize input to prevent protocol injection
	input = Sanitize(input)
	
	// Handle empty input
	if input == "" {
		return []app.Command{}, nil
	}
	
	// Check if it's a command (starts with /)
	if len(input) > 0 && input[0] == '/' {
		return r.handleSlashCommand(state, input)
	}
	
	// Not a command, treat as channel message
	return r.handleChannelMessage(state, input)
}

// handleSlashCommand processes commands that start with /
func (r *Router) handleSlashCommand(state app.ClientState, input string) ([]app.Command, error) {
	// Parse command and arguments
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return []app.Command{}, nil
	}
	
	// Extract command (remove the leading /)
	if len(fields[0]) <= 1 {
		return []app.Command{}, nil // Just a slash
	}
	
	cmd := strings.ToLower(fields[0][1:])
	args := fields[1:]
	
	// Look up handler
	handler, exists := r.handlers[cmd]
	if !exists {
		// Unknown command, return empty result (no error)
		return []app.Command{}, nil
	}
	
	// Execute handler
	rawCommands, err := handler.Handle(state, args)
	if err != nil {
		return nil, fmt.Errorf("command %s failed: %w", cmd, err)
	}
	
	// Convert raw commands to app.Command objects
	commands := make([]app.Command, 0, len(rawCommands))
	for _, raw := range rawCommands {
		command := r.parseCommand(raw, cmd)
		commands = append(commands, command)
	}
	
	return commands, nil
}

// handleChannelMessage processes regular messages to be sent to current channel
func (r *Router) handleChannelMessage(state app.ClientState, input string) ([]app.Command, error) {
	if state.CurrentChannel == "" {
		return []app.Command{}, nil
	}
	
	raw := fmt.Sprintf("PRIVMSG %s :%s", state.CurrentChannel, input)
	command := app.Command{
		Raw:     raw,
		Type:    app.CommandTypePrivMsg,
		Target:  state.CurrentChannel,
		Message: input,
	}
	
	return []app.Command{command}, nil
}

// parseCommand converts raw IRC command string to app.Command object
func (r *Router) parseCommand(raw, cmdType string) app.Command {
	command := app.Command{
		Raw: raw,
	}
	
	// Determine command type and extract relevant information
	if strings.HasPrefix(raw, "JOIN ") {
		command.Type = app.CommandTypeJoin
		parts := strings.Fields(raw)
		if len(parts) >= 2 {
			command.Target = parts[1]
			command.StateChange = &app.StateChange{
				ChannelChange: &app.ChannelChange{
					Action:  app.ChannelActionJoin,
					Channel: parts[1],
				},
			}
			if len(parts) >= 3 {
				command.StateChange.ChannelChange.Key = parts[2]
			}
		}
	} else if strings.HasPrefix(raw, "PART ") {
		command.Type = app.CommandTypePart
		parts := strings.Fields(raw)
		if len(parts) >= 2 {
			command.Target = parts[1]
			command.StateChange = &app.StateChange{
				ChannelChange: &app.ChannelChange{
					Action:  app.ChannelActionPart,
					Channel: parts[1],
				},
			}
		}
	} else if strings.HasPrefix(raw, "NICK ") {
		command.Type = app.CommandTypeNick
		parts := strings.Fields(raw)
		if len(parts) >= 2 {
			command.Target = parts[1]
			command.StateChange = &app.StateChange{
				NickChange: &app.NickChange{
					NewNick: parts[1],
				},
			}
		}
	} else if strings.HasPrefix(raw, "PRIVMSG ") {
		command.Type = app.CommandTypePrivMsg
		parts := strings.SplitN(raw, " :", 2)
		if len(parts) >= 2 {
			command.Message = parts[1]
			msgParts := strings.Fields(parts[0])
			if len(msgParts) >= 2 {
				command.Target = msgParts[1]
			}
		}
	} else if strings.HasPrefix(raw, "QUIT") {
		command.Type = app.CommandTypeQuit
		command.StateChange = &app.StateChange{
			ShouldQuit: true,
		}
		if strings.Contains(raw, ":") {
			parts := strings.SplitN(raw, ":", 2)
			if len(parts) >= 2 {
				command.Message = parts[1]
			}
		}
	} else {
		command.Type = app.CommandTypeRaw
	}
	
	return command
}

// RegisterHandler adds a new command handler
func (r *Router) RegisterHandler(command string, handler app.CommandHandler) error {
	if command == "" {
		return fmt.Errorf("command name cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}
	
	r.handlers[strings.ToLower(command)] = handler
	return nil
}

// UnregisterHandler removes a command handler
func (r *Router) UnregisterHandler(command string) error {
	if command == "" {
		return fmt.Errorf("command name cannot be empty")
	}
	
	cmd := strings.ToLower(command)
	if _, exists := r.handlers[cmd]; !exists {
		return fmt.Errorf("handler for command %q not found", command)
	}
	
	delete(r.handlers, cmd)
	return nil
}

// GetRegisteredCommands returns a list of all registered command names
func (r *Router) GetRegisteredCommands() []string {
	commands := make([]string, 0, len(r.handlers))
	for cmd := range r.handlers {
		commands = append(commands, cmd)
	}
	return commands
}

// HasHandler returns true if a handler is registered for the given command
func (r *Router) HasHandler(command string) bool {
	_, exists := r.handlers[strings.ToLower(command)]
	return exists
}