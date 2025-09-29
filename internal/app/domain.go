package app

import (
	"time"

	"irc-client/parser"
)

// ClientState represents the current state needed for command processing
type ClientState struct {
	CurrentChannel string
	Nick           string
	Connected      bool
	Server         string
	Port           int
}

// Command represents a processed IRC command ready to send
type Command struct {
	Raw         string            // The raw IRC protocol string
	Type        CommandType       // Type of command for routing
	Target      string            // Target channel/user (if applicable)
	Message     string            // Message content (if applicable)
	Parameters  map[string]string // Additional parameters
	StateChange *StateChange      // Expected state change (if any)
}

// CommandType represents different types of IRC commands
type CommandType int

const (
	CommandTypeUnknown CommandType = iota
	CommandTypeJoin
	CommandTypePart
	CommandTypePrivMsg
	CommandTypeNick
	CommandTypeQuit
	CommandTypePing
	CommandTypePong
	CommandTypeRaw // For direct IRC protocol commands
)

func (ct CommandType) String() string {
	switch ct {
	case CommandTypeJoin:
		return "JOIN"
	case CommandTypePart:
		return "PART"
	case CommandTypePrivMsg:
		return "PRIVMSG"
	case CommandTypeNick:
		return "NICK"
	case CommandTypeQuit:
		return "QUIT"
	case CommandTypePing:
		return "PING"
	case CommandTypePong:
		return "PONG"
	case CommandTypeRaw:
		return "RAW"
	default:
		return "UNKNOWN"
	}
}

// StateChange represents expected changes to session state
type StateChange struct {
	ChannelChange *ChannelChange
	NickChange    *NickChange
	ShouldQuit    bool
}

// ChannelChange represents a change in channel state
type ChannelChange struct {
	Action  ChannelAction
	Channel string
	Key     string // Optional channel key
}

// ChannelAction represents types of channel state changes
type ChannelAction int

const (
	ChannelActionJoin ChannelAction = iota
	ChannelActionPart
)

// NickChange represents a nickname change
type NickChange struct {
	NewNick string
}

// IRCEvent represents a parsed IRC event with context
type IRCEvent struct {
	Message   *parser.Message
	Timestamp time.Time
	Source    string // Simplified source (nick or "SERVER")
	Type      EventType
}

// EventType categorizes IRC events
type EventType int

const (
	EventTypeUnknown EventType = iota
	EventTypeMessage
	EventTypeJoin
	EventTypePart
	EventTypeQuit
	EventTypeNick
	EventTypePing
	EventTypeNotice
	EventTypeSystemMessage
)

func (et EventType) String() string {
	switch et {
	case EventTypeMessage:
		return "MESSAGE"
	case EventTypeJoin:
		return "JOIN"
	case EventTypePart:
		return "PART"
	case EventTypeQuit:
		return "QUIT"
	case EventTypeNick:
		return "NICK"
	case EventTypePing:
		return "PING"
	case EventTypeNotice:
		return "NOTICE"
	case EventTypeSystemMessage:
		return "SYSTEM"
	default:
		return "UNKNOWN"
	}
}

// Config holds application configuration
type Config struct {
	// IRC connection settings
	Server   string
	Port     int
	TLS      bool
	Insecure bool
	Timeout  time.Duration
	
	// User settings
	Nick     string
	RealName string
	
	// Application settings
	Verbose      bool
	EnableKawaii bool // Whether to show kawaii art
}

// MessageFormat represents different message display formats
type MessageFormat int

const (
	FormatDefault MessageFormat = iota
	FormatPrivMsg
	FormatSystemMessage
	FormatError
	FormatJoinPart
)

// UIMessage represents a formatted message for display
type UIMessage struct {
	Timestamp time.Time
	Format    MessageFormat
	Source    string
	Content   string
	Channel   string // Channel context (if applicable)
}

// ConnectionEvent represents connection state changes
type ConnectionEvent struct {
	Type      ConnectionEventType
	Timestamp time.Time
	Message   string
	Error     error
}

// ConnectionEventType represents types of connection events
type ConnectionEventType int

const (
	ConnectionEventConnecting ConnectionEventType = iota
	ConnectionEventConnected
	ConnectionEventDisconnected
	ConnectionEventError
	ConnectionEventReconnecting
)

func (cet ConnectionEventType) String() string {
	switch cet {
	case ConnectionEventConnecting:
		return "CONNECTING"
	case ConnectionEventConnected:
		return "CONNECTED"
	case ConnectionEventDisconnected:
		return "DISCONNECTED"
	case ConnectionEventError:
		return "ERROR"
	case ConnectionEventReconnecting:
		return "RECONNECTING"
	default:
		return "UNKNOWN"
	}
}