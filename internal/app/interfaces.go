package app

import (
        "bufio"
        "context"
        "time"

        "irc-client/parser"
)

// SessionManager handles IRC session state including nick and channel tracking
type SessionManager interface {
        // GetCurrentNick returns the current nickname
        GetCurrentNick() string
        
        // SetCurrentNick updates the current nickname
        SetCurrentNick(nick string)
        
        // GetCurrentChannel returns the currently active channel
        GetCurrentChannel() string
        
        // JoinChannel updates state when joining a channel
        JoinChannel(channel string)
        
        // PartChannel updates state when leaving a channel
        PartChannel() string // returns the channel that was left
        
        // HandleMessage processes incoming IRC messages and updates session state
        HandleMessage(msg *parser.Message) error
        
        // GetClientState returns current state for command processing
        GetClientState() ClientState
}

// CommandRouter handles conversion of user input into IRC commands
type CommandRouter interface {
        // Route processes user input and returns IRC commands to send
        Route(state ClientState, input string) ([]Command, error)
        
        // RegisterHandler adds a new command handler (for extensibility)
        RegisterHandler(command string, handler CommandHandler) error
        
        // UnregisterHandler removes a command handler
        UnregisterHandler(command string) error
}

// CommandHandler processes a specific IRC command
type CommandHandler interface {
        // Handle processes the command and returns IRC protocol strings
        Handle(state ClientState, args []string) ([]string, error)
        
        // Description returns help text for the command
        Description() string
}

// Connection abstracts network connection management
type Connection interface {
        // Connect establishes connection to IRC server
        Connect() error
        
        // Write sends data to the IRC server
        Write(data string) error
        
        // Reader returns a reader for incoming data
        Reader() *bufio.Reader
        
        // ErrorChannel returns channel for connection errors
        ErrorChannel() <-chan error
        
        // Close gracefully shuts down the connection
        Close() error
        
        // Context returns context for coordinated shutdown
        Context() context.Context
        
        // IsConnected returns true if connection is established
        IsConnected() bool
}

// UI handles user interface operations
type UI interface {
        // DisplayWelcome shows welcome message and art
        DisplayWelcome(message string)
        
        // DisplayMessage shows a formatted message to user
        DisplayMessage(timestamp time.Time, source, message string)
        
        // DisplaySystemMessage shows system/server messages
        DisplaySystemMessage(timestamp time.Time, source, command string, params []string)
        
        // DisplayConnected shows connection success message
        DisplayConnected(message string)
        
        // DisplayError shows error messages
        DisplayError(message string)
        
        // PromptInput prompts user for input with default value
        PromptInput(label, defaultValue string) string
        
        // InputChannel returns channel for user input
        InputChannel() <-chan string
        
        // StartInputLoop begins reading user input asynchronously
        StartInputLoop(ctx context.Context) error
        
        // Close shuts down UI resources
        Close() error
}

// EventHandler processes parsed IRC events
type EventHandler interface {
        // HandlePing responds to server ping
        HandlePing(params []string) error
        
        // HandleMessage processes incoming messages
        HandleMessage(msg *IRCEvent) error
        
        // HandleJoin processes join events
        HandleJoin(event *IRCEvent) error
        
        // HandlePart processes part events
        HandlePart(event *IRCEvent) error
        
        // HandleQuit processes quit events
        HandleQuit(event *IRCEvent) error
        
        // HandleNick processes nick change events
        HandleNick(event *IRCEvent) error
}

// Application orchestrates all components
type Application interface {
        // Initialize sets up all components
        Initialize(config Config) error
        
        // Start begins the application main loop
        Start(ctx context.Context) error
        
        // Stop gracefully shuts down the application
        Stop() error
        
        // GetSession returns the session manager
        GetSession() SessionManager
        
        // GetRouter returns the command router
        GetRouter() CommandRouter
        
        // GetConnection returns the network connection
        GetConnection() Connection
        
        // GetUI returns the user interface
        GetUI() UI
}