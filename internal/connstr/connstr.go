package connstr

import (
	"fmt"
	"net/url"
	"strings"
)

// ConnInfo stores pass-auth connection details
type ConnInfo struct {
	User     string
	Password string
	Host     string
	Port     string
}

// ParseConnectionString parses the connection string using the `net/url` package
func ParseConnectionString(connStr string) (*ConnInfo, error) {
	if strings.TrimSpace(connStr) == "" {
		return nil, fmt.Errorf("empty connection string")
	}

	// Ensure the scheme exists (prefix with dummy scheme)
	if !strings.Contains(connStr, "://") {
		connStr = "ssh://" + connStr
	}

	// Parse using net/url
	u, err := url.Parse(connStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL format: %s", err)
	}

	// Extract username and password
	user := ""
	password := ""
	if u.User != nil {
		user = u.User.Username()
		password, _ = u.User.Password() // Password might be empty
	}

	// Extract host and port
	hostName := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "22" // Default SSH port
	}

	// Check required
	if user == "" {
		return nil, fmt.Errorf("user is required")
	}
	if hostName == "" {
		return nil, fmt.Errorf("hostname is required")
	}

	return &ConnInfo{
		User:     user,
		Password: password,
		Host:     hostName,
		Port:     port,
	}, nil
}
