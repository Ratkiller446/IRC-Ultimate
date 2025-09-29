package app

import (
        "bufio"
        "context"
        "strings"
        "testing"
        "time"

        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/mock"

        "irc-client/parser"
)

// MockSessionManager implements SessionManager for testing
type MockSessionManager struct {
        mock.Mock
}

func (m *MockSessionManager) GetCurrentNick() string {
        args := m.Called()
        return args.String(0)
}

func (m *MockSessionManager) SetCurrentNick(nick string) {
        m.Called(nick)
}

func (m *MockSessionManager) GetCurrentChannel() string {
        args := m.Called()
        return args.String(0)
}

func (m *MockSessionManager) JoinChannel(channel string) {
        m.Called(channel)
}

func (m *MockSessionManager) PartChannel() string {
        args := m.Called()
        return args.String(0)
}

func (m *MockSessionManager) HandleMessage(msg *parser.Message) error {
        args := m.Called(msg)
        return args.Error(0)
}

func (m *MockSessionManager) GetClientState() ClientState {
        args := m.Called()
        return args.Get(0).(ClientState)
}

// MockCommandRouter implements CommandRouter for testing
type MockCommandRouter struct {
        mock.Mock
}

func (m *MockCommandRouter) Route(state ClientState, input string) ([]Command, error) {
        args := m.Called(state, input)
        return args.Get(0).([]Command), args.Error(1)
}

func (m *MockCommandRouter) RegisterHandler(command string, handler CommandHandler) error {
        args := m.Called(command, handler)
        return args.Error(0)
}

func (m *MockCommandRouter) UnregisterHandler(command string) error {
        args := m.Called(command)
        return args.Error(0)
}

// MockCommandHandler implements CommandHandler for testing
type MockCommandHandler struct {
        mock.Mock
}

func (m *MockCommandHandler) Handle(state ClientState, args []string) ([]string, error) {
        mockArgs := m.Called(state, args)
        return mockArgs.Get(0).([]string), mockArgs.Error(1)
}

func (m *MockCommandHandler) Description() string {
        args := m.Called()
        return args.String(0)
}

// MockConnection implements Connection for testing
type MockConnection struct {
        mock.Mock
        reader    *bufio.Reader
        errorCh   chan error
        ctx       context.Context
        cancel    context.CancelFunc
}

func NewMockConnection() *MockConnection {
        ctx, cancel := context.WithCancel(context.Background())
        return &MockConnection{
                reader:  bufio.NewReader(strings.NewReader("")),
                errorCh: make(chan error, 1),
                ctx:     ctx,
                cancel:  cancel,
        }
}

func (m *MockConnection) Connect() error {
        args := m.Called()
        return args.Error(0)
}

func (m *MockConnection) Write(data string) error {
        args := m.Called(data)
        return args.Error(0)
}

func (m *MockConnection) Reader() *bufio.Reader {
        args := m.Called()
        if len(args) > 0 {
                return args.Get(0).(*bufio.Reader)
        }
        return m.reader
}

func (m *MockConnection) ErrorChannel() <-chan error {
        args := m.Called()
        if len(args) > 0 {
                return args.Get(0).(<-chan error)
        }
        return m.errorCh
}

func (m *MockConnection) Close() error {
        args := m.Called()
        m.cancel()
        close(m.errorCh)
        return args.Error(0)
}

func (m *MockConnection) Context() context.Context {
        args := m.Called()
        if len(args) > 0 {
                return args.Get(0).(context.Context)
        }
        return m.ctx
}

func (m *MockConnection) IsConnected() bool {
        args := m.Called()
        return args.Bool(0)
}

// MockUI implements UI for testing
type MockUI struct {
        mock.Mock
        inputCh chan string
        ctx     context.Context
        cancel  context.CancelFunc
}

func NewMockUI() *MockUI {
        ctx, cancel := context.WithCancel(context.Background())
        return &MockUI{
                inputCh: make(chan string, 10),
                ctx:     ctx,
                cancel:  cancel,
        }
}

func (m *MockUI) DisplayWelcome(message string) {
        m.Called(message)
}

func (m *MockUI) DisplayMessage(timestamp time.Time, source, message string) {
        m.Called(timestamp, source, message)
}

func (m *MockUI) DisplaySystemMessage(timestamp time.Time, source, command string, params []string) {
        m.Called(timestamp, source, command, params)
}

func (m *MockUI) DisplayConnected(message string) {
        m.Called(message)
}

func (m *MockUI) DisplayError(message string) {
        m.Called(message)
}

func (m *MockUI) PromptInput(label, defaultValue string) string {
        args := m.Called(label, defaultValue)
        return args.String(0)
}

func (m *MockUI) InputChannel() <-chan string {
        args := m.Called()
        if len(args) > 0 {
                return args.Get(0).(<-chan string)
        }
        return m.inputCh
}

func (m *MockUI) StartInputLoop(ctx context.Context) error {
        args := m.Called(ctx)
        return args.Error(0)
}

func (m *MockUI) Close() error {
        args := m.Called()
        m.cancel()
        close(m.inputCh)
        return args.Error(0)
}

// Test basic interface contract compliance
func TestInterfaceContracts(t *testing.T) {
        // Verify our mocks implement the interfaces
        var _ SessionManager = (*MockSessionManager)(nil)
        var _ CommandRouter = (*MockCommandRouter)(nil)
        var _ CommandHandler = (*MockCommandHandler)(nil)
        var _ Connection = (*MockConnection)(nil)
        var _ UI = (*MockUI)(nil)
}

// Test domain struct creation and validation
func TestClientState(t *testing.T) {
        state := ClientState{
                CurrentChannel: "#test",
                Nick:           "testnick",
                Connected:      true,
                Server:         "irc.example.com",
                Port:           6667,
        }

        assert.Equal(t, "#test", state.CurrentChannel)
        assert.Equal(t, "testnick", state.Nick)
        assert.True(t, state.Connected)
        assert.Equal(t, "irc.example.com", state.Server)
        assert.Equal(t, 6667, state.Port)
}

func TestCommand(t *testing.T) {
        cmd := Command{
                Raw:    "JOIN #test",
                Type:   CommandTypeJoin,
                Target: "#test",
                StateChange: &StateChange{
                        ChannelChange: &ChannelChange{
                                Action:  ChannelActionJoin,
                                Channel: "#test",
                        },
                },
        }

        assert.Equal(t, "JOIN #test", cmd.Raw)
        assert.Equal(t, CommandTypeJoin, cmd.Type)
        assert.Equal(t, "#test", cmd.Target)
        assert.NotNil(t, cmd.StateChange)
        assert.NotNil(t, cmd.StateChange.ChannelChange)
        assert.Equal(t, ChannelActionJoin, cmd.StateChange.ChannelChange.Action)
}

func TestCommandTypeString(t *testing.T) {
        tests := []struct {
                cmdType  CommandType
                expected string
        }{
                {CommandTypeJoin, "JOIN"},
                {CommandTypePart, "PART"},
                {CommandTypePrivMsg, "PRIVMSG"},
                {CommandTypeNick, "NICK"},
                {CommandTypeQuit, "QUIT"},
                {CommandTypePing, "PING"},
                {CommandTypePong, "PONG"},
                {CommandTypeRaw, "RAW"},
                {CommandTypeUnknown, "UNKNOWN"},
        }

        for _, tt := range tests {
                t.Run(tt.expected, func(t *testing.T) {
                        assert.Equal(t, tt.expected, tt.cmdType.String())
                })
        }
}

func TestIRCEvent(t *testing.T) {
        msg := &parser.Message{
                Command: "PRIVMSG",
                Prefix:  "nick!user@host",
                Params:  []string{"#channel", "Hello world"},
        }

        event := IRCEvent{
                Message:   msg,
                Timestamp: time.Now(),
                Source:    "nick",
                Type:      EventTypeMessage,
        }

        assert.Equal(t, msg, event.Message)
        assert.Equal(t, "nick", event.Source)
        assert.Equal(t, EventTypeMessage, event.Type)
}

func TestConfig(t *testing.T) {
        config := Config{
                Server:       "irc.libera.chat",
                Port:         6697,
                TLS:          true,
                Nick:         "testnick",
                RealName:     "Test User",
                Verbose:      false,
                EnableKawaii: true,
        }

        assert.Equal(t, "irc.libera.chat", config.Server)
        assert.Equal(t, 6697, config.Port)
        assert.True(t, config.TLS)
        assert.Equal(t, "testnick", config.Nick)
        assert.True(t, config.EnableKawaii)
}

// Test adapter behavior with mocks
func TestMockSessionManager(t *testing.T) {
        mockSession := &MockSessionManager{}
        
        // Set up expectations
        mockSession.On("GetCurrentNick").Return("testnick")
        mockSession.On("SetCurrentNick", "newnick").Return()
        mockSession.On("GetCurrentChannel").Return("#test")
        
        // Test calls
        assert.Equal(t, "testnick", mockSession.GetCurrentNick())
        mockSession.SetCurrentNick("newnick")
        assert.Equal(t, "#test", mockSession.GetCurrentChannel())
        
        // Verify expectations
        mockSession.AssertExpectations(t)
}

func TestMockCommandRouter(t *testing.T) {
        mockRouter := &MockCommandRouter{}
        
        state := ClientState{Nick: "test", CurrentChannel: "#test"}
        expectedCommands := []Command{
                {Raw: "PRIVMSG #test :hello", Type: CommandTypePrivMsg},
        }
        
        // Set up expectations
        mockRouter.On("Route", state, "hello").Return(expectedCommands, nil)
        
        // Test call
        commands, err := mockRouter.Route(state, "hello")
        assert.NoError(t, err)
        assert.Equal(t, expectedCommands, commands)
        
        // Verify expectations
        mockRouter.AssertExpectations(t)
}

func TestMockConnection(t *testing.T) {
        mockConn := NewMockConnection()
        
        // Set up expectations
        mockConn.On("Connect").Return(nil)
        mockConn.On("Write", "NICK test\r\n").Return(nil)
        mockConn.On("IsConnected").Return(true)
        mockConn.On("Close").Return(nil)
        
        // Test calls
        assert.NoError(t, mockConn.Connect())
        assert.NoError(t, mockConn.Write("NICK test\r\n"))
        assert.True(t, mockConn.IsConnected())
        assert.NoError(t, mockConn.Close())
        
        // Verify expectations
        mockConn.AssertExpectations(t)
}

func TestMockUI(t *testing.T) {
        mockUI := NewMockUI()
        
        // Set up expectations
        mockUI.On("DisplayWelcome", "Welcome!").Return()
        mockUI.On("PromptInput", "Nick", "default").Return("testnick")
        mockUI.On("DisplayMessage", mock.AnythingOfType("time.Time"), "source", "message").Return()
        mockUI.On("Close").Return(nil)
        
        // Test calls
        mockUI.DisplayWelcome("Welcome!")
        nick := mockUI.PromptInput("Nick", "default")
        assert.Equal(t, "testnick", nick)
        mockUI.DisplayMessage(time.Now(), "source", "message")
        assert.NoError(t, mockUI.Close())
        
        // Verify expectations
        mockUI.AssertExpectations(t)
}