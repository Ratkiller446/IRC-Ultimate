package conn

import (
        "bufio"
        "context"
        "fmt"
        "net"
        "sync"
        "time"
)

// WriteRequest represents a write operation to be performed by the connection manager
type WriteRequest struct {
        Data     string
        Response chan error // Channel to send back the result of the write operation
}

// ConnectionState represents the current state of the IRC connection
type ConnectionState int

const (
        StateDisconnected ConnectionState = iota
        StateConnecting
        StateConnected
        StateReconnecting
        StateClosing
        StateClosed
)

func (s ConnectionState) String() string {
        switch s {
        case StateDisconnected:
                return "Disconnected"
        case StateConnecting:
                return "Connecting"
        case StateConnected:
                return "Connected"
        case StateReconnecting:
                return "Reconnecting"
        case StateClosing:
                return "Closing"
        case StateClosed:
                return "Closed"
        default:
                return "Unknown"
        }
}

// Manager provides thread-safe IRC connection management with single-writer pattern
type Manager struct {
        config Config
        conn   net.Conn
        writer *bufio.Writer
        reader *bufio.Reader

        // Single-writer pattern
        writeQueue   chan WriteRequest
        writerDone   chan struct{}
        readerDone   chan struct{}
        
        // State management
        state     ConnectionState
        stateMu   sync.RWMutex
        
        // Cancellation and cleanup
        ctx        context.Context
        cancel     context.CancelFunc
        closeOnce  sync.Once
        writeMu    sync.Mutex  // protects write operations from races during shutdown
        
        // Error handling
        errorCh    chan error
        
        // Configuration
        writeTimeout time.Duration
        readTimeout  time.Duration
}

// NewManager creates a new connection manager with the given configuration
func NewManager(config Config) *Manager {
        ctx, cancel := context.WithCancel(context.Background())
        
        return &Manager{
                config:       config,
                writeQueue:   make(chan WriteRequest, 100), // Buffered channel for write requests
                writerDone:   make(chan struct{}),
                readerDone:   make(chan struct{}),
                state:        StateDisconnected,
                ctx:          ctx,
                cancel:       cancel,
                errorCh:      make(chan error, 10), // Buffered error channel
                writeTimeout: 30 * time.Second,
                readTimeout:  60 * time.Second,
        }
}

// Connect establishes a connection to the IRC server
func (m *Manager) Connect() error {
        m.stateMu.Lock()
        defer m.stateMu.Unlock()
        
        if m.state != StateDisconnected {
                return fmt.Errorf("connection already in progress or established")
        }
        
        m.setState(StateConnecting)
        
        // Use ConnectWithContext to respect manager's context for cancellation
        conn, err := ConnectWithContext(m.ctx, m.config)
        if err != nil {
                m.setState(StateDisconnected)
                return fmt.Errorf("failed to connect: %w", err)
        }
        
        m.conn = conn
        m.writer = bufio.NewWriter(conn)
        m.reader = bufio.NewReader(conn)
        m.setState(StateConnected)
        
        // Start the single writer goroutine
        go m.writerLoop()
        
        return nil
}

// Write sends data to the IRC server using the single-writer pattern
func (m *Manager) Write(data string) error {
        // First check if context is cancelled to prevent race with Close()
        select {
        case <-m.ctx.Done():
                return fmt.Errorf("connection manager is shutting down")
        default:
        }
        
        // Check connection state while holding read lock
        m.stateMu.RLock()
        state := m.state
        m.stateMu.RUnlock()
        
        if state != StateConnected {
                return fmt.Errorf("connection not established")
        }
        
        // Create a response channel to get the result
        responseCh := make(chan error, 1)
        
        writeReq := WriteRequest{
                Data:     data,
                Response: responseCh,
        }
        
        // Protect the write operation with a mutex to prevent race with shutdown
        // This ensures atomicity between checking context and sending to channel
        m.writeMu.Lock()
        defer m.writeMu.Unlock()
        
        // Check context while holding the write lock
        select {
        case <-m.ctx.Done():
                return fmt.Errorf("connection manager is shutting down")
        default:
        }
        
        // Send write request to the single writer goroutine with context protection
        select {
        case m.writeQueue <- writeReq:
                // Successfully queued write request
        case <-m.ctx.Done():
                return fmt.Errorf("connection manager is shutting down")
        case <-time.After(m.writeTimeout):
                return fmt.Errorf("write timeout")
        }
        
        // Wait for the write result
        select {
        case err := <-responseCh:
                return err
        case <-m.ctx.Done():
                return fmt.Errorf("connection manager is shutting down")
        case <-time.After(m.writeTimeout):
                return fmt.Errorf("write response timeout")
        }
}

// Reader returns a reader for the connection
func (m *Manager) Reader() *bufio.Reader {
        m.stateMu.RLock()
        defer m.stateMu.RUnlock()
        return m.reader
}

// GetState returns the current connection state
func (m *Manager) GetState() ConnectionState {
        m.stateMu.RLock()
        defer m.stateMu.RUnlock()
        return m.state
}

// setState updates the connection state (internal use)
func (m *Manager) setState(state ConnectionState) {
        m.state = state
}

// ErrorChannel returns a channel that receives connection errors
func (m *Manager) ErrorChannel() <-chan error {
        return m.errorCh
}

// Close gracefully shuts down the connection manager
func (m *Manager) Close() error {
        var closeErr error
        
        m.closeOnce.Do(func() {
                // Acquire write lock to ensure no writes are in progress
                m.writeMu.Lock()
                
                m.stateMu.Lock()
                m.setState(StateClosing)
                m.stateMu.Unlock()
                
                // Cancel context to signal shutdown
                m.cancel()
                
                // Release write lock after context cancellation
                m.writeMu.Unlock()
                
                // Wait for writer goroutine to finish and close the queue itself
                // This eliminates the race condition since only writerLoop closes writeQueue
                select {
                case <-m.writerDone:
                case <-time.After(5 * time.Second):
                        // Timeout waiting for writer to finish
                }
                
                // Close the underlying connection
                if m.conn != nil {
                        closeErr = m.conn.Close()
                }
                
                m.stateMu.Lock()
                m.setState(StateClosed)
                m.stateMu.Unlock()
                
                // Close error channel
                close(m.errorCh)
        })
        
        return closeErr
}

// writerLoop is the single writer goroutine that handles all write operations
func (m *Manager) writerLoop() {
        defer func() {
                // Close writeQueue on shutdown to ensure no more writes can be queued
                // This is safe since only writerLoop closes the queue, eliminating race conditions
                close(m.writeQueue)
                close(m.writerDone)
        }()
        
        for {
                select {
                case writeReq, ok := <-m.writeQueue:
                        if !ok {
                                // Write queue closed, shutdown
                                return
                        }
                        
                        // Perform the write operation
                        err := m.performWrite(writeReq.Data)
                        
                        // Send result back to caller
                        select {
                        case writeReq.Response <- err:
                        case <-m.ctx.Done():
                                return
                        default:
                                // If caller is not waiting for response, continue
                        }
                        
                case <-m.ctx.Done():
                        // Context cancelled, drain remaining writes and shutdown gracefully
                        m.drainWriteQueue()
                        return
                }
        }
}

// drainWriteQueue processes any remaining write requests during shutdown
func (m *Manager) drainWriteQueue() {
        for {
                select {
                case writeReq, ok := <-m.writeQueue:
                        if !ok {
                                return
                        }
                        // Send shutdown error to any pending writes
                        select {
                        case writeReq.Response <- fmt.Errorf("connection manager is shutting down"):
                        default:
                        }
                default:
                        // No more pending writes
                        return
                }
        }
}

// performWrite actually writes data to the connection
func (m *Manager) performWrite(data string) error {
        if m.conn == nil {
                return fmt.Errorf("connection not established")
        }
        
        // Set write deadline
        if err := m.conn.SetWriteDeadline(time.Now().Add(m.writeTimeout)); err != nil {
                m.reportError(fmt.Errorf("failed to set write deadline: %w", err))
                return err
        }
        
        // Write data to buffer
        if _, err := m.writer.WriteString(data); err != nil {
                m.reportError(fmt.Errorf("failed to write to buffer: %w", err))
                return err
        }
        
        // Flush buffer to connection
        if err := m.writer.Flush(); err != nil {
                m.reportError(fmt.Errorf("failed to flush buffer: %w", err))
                return err
        }
        
        return nil
}

// reportError sends an error to the error channel if possible
func (m *Manager) reportError(err error) {
        select {
        case m.errorCh <- err:
        default:
                // Error channel full, drop error
        }
}

// Context returns the manager's context for cancellation
func (m *Manager) Context() context.Context {
        return m.ctx
}

// IsConnected returns true if the connection is established
func (m *Manager) IsConnected() bool {
        return m.GetState() == StateConnected
}