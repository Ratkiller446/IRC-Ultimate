package conn

import (
        "context"
        "crypto/tls"
        "fmt"
        "log"
        "net"
        "time"
)

// Config holds connection configuration parameters
type Config struct {
        Server   string
        Port     int
        TLS      bool
        Timeout  time.Duration
        Insecure bool
        
        // Enhanced configuration options
        KeepAlive    time.Duration
        ReadTimeout  time.Duration
        WriteTimeout time.Duration
        RetryCount   int
        RetryDelay   time.Duration
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() Config {
        return Config{
                Server:       "irc.libera.chat",
                Port:         6697,
                TLS:          true,
                Timeout:      30 * time.Second,
                Insecure:     false,
                KeepAlive:    30 * time.Second,
                ReadTimeout:  60 * time.Second,
                WriteTimeout: 30 * time.Second,
                RetryCount:   3,
                RetryDelay:   5 * time.Second,
        }
}

// Connect establishes a connection to the IRC server with enhanced error handling
func Connect(cfg Config) (net.Conn, error) {
        return ConnectWithContext(context.Background(), cfg)
}

// ConnectWithContext establishes a connection with context support for cancellation
func ConnectWithContext(ctx context.Context, cfg Config) (net.Conn, error) {
        address := fmt.Sprintf("%s:%d", cfg.Server, cfg.Port)
        
        dialer := &net.Dialer{
                Timeout:   cfg.Timeout,
                KeepAlive: cfg.KeepAlive,
        }
        
        var conn net.Conn
        var err error
        
        if cfg.TLS {
                log.Printf("[conn] Establishing TLS connection to %s", address)
                tlsCfg := &tls.Config{
                        InsecureSkipVerify: cfg.Insecure,
                        ServerName:         cfg.Server,
                }
                conn, err = tls.DialWithDialer(dialer, "tcp", address, tlsCfg)
        } else {
                log.Printf("[conn] Establishing plain TCP connection to %s", address)
                conn, err = dialer.DialContext(ctx, "tcp", address)
        }
        
        if err != nil {
                return nil, fmt.Errorf("failed to connect to %s: %w", address, err)
        }
        
        // Configure connection timeouts if specified
        if tcpConn, ok := conn.(*net.TCPConn); ok {
                if cfg.KeepAlive > 0 {
                        if err := tcpConn.SetKeepAlive(true); err != nil {
                                log.Printf("[conn] Warning: failed to enable keepalive: %v", err)
                        }
                        if err := tcpConn.SetKeepAlivePeriod(cfg.KeepAlive); err != nil {
                                log.Printf("[conn] Warning: failed to set keepalive period: %v", err)
                        }
                }
        }
        
        log.Printf("[conn] Successfully connected to %s", address)
        return conn, nil
}

// ValidateConfig validates the connection configuration
func ValidateConfig(cfg Config) error {
        if cfg.Server == "" {
                return fmt.Errorf("server address is required")
        }
        if cfg.Port <= 0 || cfg.Port > 65535 {
                return fmt.Errorf("invalid port number: %d", cfg.Port)
        }
        if cfg.Timeout <= 0 {
                return fmt.Errorf("timeout must be positive")
        }
        return nil
}
