package client

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// HookClient sends hook events to the daemon and receives compressed results.
type HookClient struct {
	socketPath string
	timeout    time.Duration
}

// NewHookClient creates a new hook client.
func NewHookClient(socketPath string, timeout time.Duration) *HookClient {
	return &HookClient{
		socketPath: socketPath,
		timeout:    timeout,
	}
}

// ProcessHookEvent sends an event to the daemon and gets back a result.
// If the daemon is not running, it starts it.
// If communication fails, returns the original input (fail-open).
func (c *HookClient) ProcessHookEvent(event map[string]interface{}) (map[string]interface{}, error) {
	// Try to connect to daemon.
	conn, err := c.dialDaemon()
	if err != nil {
		// Daemon not running; try to start it.
		if err := c.ensureDaemon(); err != nil {
			// Failed to start daemon; fail open.
			return event, nil
		}

		// Retry connection.
		conn, err = c.dialDaemon()
		if err != nil {
			// Still can't connect; fail open.
			return event, nil
		}
	}
	defer conn.Close()

	// Set timeouts.
	if err := conn.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
		return event, nil // fail open
	}
	if err := conn.SetWriteDeadline(time.Now().Add(c.timeout)); err != nil {
		return event, nil // fail open
	}

	// Send event.
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(event); err != nil {
		return event, nil // fail open
	}

	// Receive result.
	decoder := json.NewDecoder(conn)
	var result map[string]interface{}
	if err := decoder.Decode(&result); err != nil {
		return event, nil // fail open
	}

	return result, nil
}

// dialDaemon connects to the daemon socket.
func (c *HookClient) dialDaemon() (net.Conn, error) {
	conn, err := net.DialTimeout("unix", c.socketPath, c.timeout)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// ensureDaemon starts the daemon if it's not running.
func (c *HookClient) ensureDaemon() error {
	// Find the slash binary.
	slashPath, err := exec.LookPath("slash")
	if err != nil {
		return fmt.Errorf("slash binary not found")
	}

	// Start daemon in background.
	cmd := exec.Command(slashPath, "daemon")
	cmd.Stdout = nil // suppress output
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	// Give it a moment to start and create the socket.
	time.Sleep(500 * time.Millisecond)

	return nil
}

// EnsureDaemon is a package-level function to ensure daemon is running.
func EnsureDaemon(socketPath string) error {
	client := NewHookClient(socketPath, 2*time.Second)
	return client.ensureDaemon()
}

// GetSocketPath returns the default slash daemon socket path.
func GetSocketPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".slash", "daemon.sock")
}
