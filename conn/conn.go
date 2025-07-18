package conn

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"time"
)

type Config struct {
	Server   string
	Port     int
	TLS      bool
	Timeout  time.Duration
	Insecure bool
}

func Connect(cfg Config) (net.Conn, error) {
	address := fmt.Sprintf("%s:%d", cfg.Server, cfg.Port)
	dialer := net.Dialer{Timeout: cfg.Timeout}
	if cfg.TLS {
		log.Printf("[conn] Establishing TLS connection to %s", address)
		tlsCfg := &tls.Config{InsecureSkipVerify: cfg.Insecure}
		return tls.DialWithDialer(&dialer, "tcp", address, tlsCfg)
	}
	log.Printf("[conn] Establishing plain TCP connection to %s", address)
	return dialer.Dial("tcp", address)
} 