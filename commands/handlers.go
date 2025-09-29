package commands

import (
	"fmt"
	"strings"

	"irc-client/internal/app"
)

// JoinHandler handles JOIN commands
type JoinHandler struct{}

func (h *JoinHandler) Handle(state app.ClientState, args []string) ([]string, error) {
	if len(args) == 0 {
		return []string{}, nil
	}
	
	channel := args[0]
	if len(args) >= 2 {
		// JOIN with key
		key := args[1]
		return []string{fmt.Sprintf("JOIN %s %s", channel, key)}, nil
	}
	
	// JOIN without key
	return []string{fmt.Sprintf("JOIN %s", channel)}, nil
}

func (h *JoinHandler) Description() string {
	return "Join a channel. Usage: /join #channel [key]"
}

// PartHandler handles PART commands
type PartHandler struct{}

func (h *PartHandler) Handle(state app.ClientState, args []string) ([]string, error) {
	if state.CurrentChannel == "" {
		return []string{}, nil
	}
	
	if len(args) == 0 {
		// PART without reason
		return []string{fmt.Sprintf("PART %s", state.CurrentChannel)}, nil
	}
	
	// PART with reason
	reason := strings.Join(args, " ")
	return []string{fmt.Sprintf("PART %s :%s", state.CurrentChannel, reason)}, nil
}

func (h *PartHandler) Description() string {
	return "Leave the current channel. Usage: /part [reason]"
}

// NickHandler handles NICK commands
type NickHandler struct{}

func (h *NickHandler) Handle(state app.ClientState, args []string) ([]string, error) {
	if len(args) == 0 {
		return []string{}, nil
	}
	
	newNick := args[0]
	return []string{fmt.Sprintf("NICK %s", newNick)}, nil
}

func (h *NickHandler) Description() string {
	return "Change your nickname. Usage: /nick newnick"
}

// MsgHandler handles MSG/PRIVMSG commands
type MsgHandler struct{}

func (h *MsgHandler) Handle(state app.ClientState, args []string) ([]string, error) {
	if len(args) < 2 {
		return []string{}, nil
	}
	
	target := args[0]
	message := strings.Join(args[1:], " ")
	return []string{fmt.Sprintf("PRIVMSG %s :%s", target, message)}, nil
}

func (h *MsgHandler) Description() string {
	return "Send a private message. Usage: /msg target message"
}

// QuitHandler handles QUIT commands
type QuitHandler struct{}

func (h *QuitHandler) Handle(state app.ClientState, args []string) ([]string, error) {
	if len(args) == 0 {
		return []string{"QUIT"}, nil
	}
	
	reason := strings.Join(args, " ")
	return []string{fmt.Sprintf("QUIT :%s", reason)}, nil
}

func (h *QuitHandler) Description() string {
	return "Quit the IRC server. Usage: /quit [reason]"
}

// HelpHandler provides help information about commands
type HelpHandler struct {
	router *Router
}

func NewHelpHandler(router *Router) *HelpHandler {
	return &HelpHandler{router: router}
}

func (h *HelpHandler) Handle(state app.ClientState, args []string) ([]string, error) {
	// Help is displayed locally, not sent to server
	return []string{}, nil
}

func (h *HelpHandler) Description() string {
	return "Show help for commands. Usage: /help [command]"
}

func (h *HelpHandler) GetHelpText(command string) string {
	if command == "" {
		// General help
		commands := h.router.GetRegisteredCommands()
		var helpText strings.Builder
		helpText.WriteString("Available commands:\n")
		for _, cmd := range commands {
			if handler, exists := h.router.handlers[cmd]; exists {
				helpText.WriteString(fmt.Sprintf("  /%s - %s\n", cmd, handler.Description()))
			}
		}
		helpText.WriteString("Type /help <command> for detailed help on a specific command.")
		return helpText.String()
	}
	
	// Specific command help
	handler, exists := h.router.handlers[strings.ToLower(command)]
	if !exists {
		return fmt.Sprintf("Unknown command: %s", command)
	}
	
	return handler.Description()
}

// WhoHandler handles WHO commands
type WhoHandler struct{}

func (h *WhoHandler) Handle(state app.ClientState, args []string) ([]string, error) {
	if len(args) == 0 {
		// WHO for current channel
		if state.CurrentChannel != "" {
			return []string{fmt.Sprintf("WHO %s", state.CurrentChannel)}, nil
		}
		return []string{}, nil
	}
	
	target := args[0]
	return []string{fmt.Sprintf("WHO %s", target)}, nil
}

func (h *WhoHandler) Description() string {
	return "List users in a channel. Usage: /who [channel]"
}

// TopicHandler handles TOPIC commands
type TopicHandler struct{}

func (h *TopicHandler) Handle(state app.ClientState, args []string) ([]string, error) {
	if state.CurrentChannel == "" {
		return []string{}, nil
	}
	
	if len(args) == 0 {
		// Get current topic
		return []string{fmt.Sprintf("TOPIC %s", state.CurrentChannel)}, nil
	}
	
	// Set new topic
	topic := strings.Join(args, " ")
	return []string{fmt.Sprintf("TOPIC %s :%s", state.CurrentChannel, topic)}, nil
}

func (h *TopicHandler) Description() string {
	return "View or set channel topic. Usage: /topic [new topic]"
}

// MeHandler handles ACTION commands (/me)
type MeHandler struct{}

func (h *MeHandler) Handle(state app.ClientState, args []string) ([]string, error) {
	if state.CurrentChannel == "" || len(args) == 0 {
		return []string{}, nil
	}
	
	action := strings.Join(args, " ")
	// ACTION is sent as CTCP ACTION
	return []string{fmt.Sprintf("PRIVMSG %s :\x01ACTION %s\x01", state.CurrentChannel, action)}, nil
}

func (h *MeHandler) Description() string {
	return "Send an action message. Usage: /me action"
}