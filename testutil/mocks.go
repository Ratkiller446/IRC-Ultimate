package testutil

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockConn implements net.Conn for testing IRC connections
type MockConn struct {
	mock.Mock
	readData    *bytes.Buffer
	writeData   *bytes.Buffer
	closed      bool
	mu          sync.RWMutex
	readErr     error
	writeErr    error
	closeErr    error
	localAddr   net.Addr
	remoteAddr  net.Addr
}

// NewMockConn creates a new mock connection with optional initial read data
func NewMockConn(readData string) *MockConn {
	return &MockConn{
		readData:  bytes.NewBufferString(readData),
		writeData: &bytes.Buffer{},
	}
}

func (m *MockConn) Read(b []byte) (int, error) {
	args := m.Called(b)
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.closed {
		return 0, fmt.Errorf("connection closed")
	}
	
	if m.readErr != nil {
		return 0, m.readErr
	}
	
	// Read from buffer if data is available
	if m.readData.Len() > 0 {
		n, err := m.readData.Read(b)
		return n, err
	}
	
	// If no mock expectations are set, use the buffer behavior
	if len(args) == 0 {
		return 0, fmt.Errorf("EOF")
	}
	
	return args.Int(0), args.Error(1)
}

func (m *MockConn) Write(b []byte) (int, error) {
	args := m.Called(b)
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.closed {
		return 0, fmt.Errorf("connection closed")
	}
	
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	
	// Write to buffer for inspection
	n, err := m.writeData.Write(b)
	
	// If no mock expectations are set, use the buffer behavior
	if len(args) == 0 {
		return n, err
	}
	
	return args.Int(0), args.Error(1)
}

func (m *MockConn) Close() error {
	args := m.Called()
	
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	
	if m.closeErr != nil {
		return m.closeErr
	}
	
	if len(args) > 0 {
		return args.Error(0)
	}
	return nil
}

func (m *MockConn) LocalAddr() net.Addr {
	args := m.Called()
	if len(args) > 0 {
		return args.Get(0).(net.Addr)
	}
	return m.localAddr
}

func (m *MockConn) RemoteAddr() net.Addr {
	args := m.Called()
	if len(args) > 0 {
		return args.Get(0).(net.Addr)
	}
	return m.remoteAddr
}

func (m *MockConn) SetDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockConn) SetReadDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockConn) SetWriteDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

// Mock helper methods for testing

func (m *MockConn) GetWrittenData() []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.writeData.Bytes()
}

func (m *MockConn) GetWrittenString() string {
	return string(m.GetWrittenData())
}

func (m *MockConn) AddReadData(data string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readData.WriteString(data)
}

func (m *MockConn) SetReadError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readErr = err
}

func (m *MockConn) SetWriteError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writeErr = err
}

func (m *MockConn) SetCloseError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closeErr = err
}

func (m *MockConn) IsClosed() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.closed
}

func (m *MockConn) ClearWrittenData() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writeData.Reset()
}

// MockAddr implements net.Addr for testing
type MockAddr struct {
	NetworkString string
	AddressString string
}

func (m MockAddr) Network() string {
	return m.NetworkString
}

func (m MockAddr) String() string {
	return m.AddressString
}

// MockReader implements io.Reader for testing
type MockReader struct {
	mock.Mock
	data *bytes.Buffer
}

func NewMockReader(data string) *MockReader {
	return &MockReader{
		data: bytes.NewBufferString(data),
	}
}

func (m *MockReader) Read(p []byte) (int, error) {
	args := m.Called(p)
	
	if len(args) == 0 {
		return m.data.Read(p)
	}
	
	return args.Int(0), args.Error(1)
}

// MockWriter implements io.Writer for testing
type MockWriter struct {
	mock.Mock
	data *bytes.Buffer
}

func NewMockWriter() *MockWriter {
	return &MockWriter{
		data: &bytes.Buffer{},
	}
}

func (m *MockWriter) Write(p []byte) (int, error) {
	args := m.Called(p)
	
	// Always write to buffer for inspection
	m.data.Write(p)
	
	if len(args) == 0 {
		return len(p), nil
	}
	
	return args.Int(0), args.Error(1)
}

func (m *MockWriter) GetWritten() string {
	return m.data.String()
}

func (m *MockWriter) Clear() {
	m.data.Reset()
}

// MockBufioReader wraps a MockReader with bufio functionality
func NewMockBufioReader(data string) *bufio.Reader {
	return bufio.NewReader(NewMockReader(data))
}

// MockBufioWriter wraps a MockWriter with bufio functionality
func NewMockBufioWriter() (*bufio.Writer, *MockWriter) {
	mockWriter := NewMockWriter()
	return bufio.NewWriter(mockWriter), mockWriter
}