package client

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

func DefaultSocket() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".slash", "daemon.sock")
}

func SendEvent(raw []byte, socketPath string) ([]byte, error) {
	if socketPath == "" {
		socketPath = DefaultSocket()
	}

	conn, err := net.DialTimeout("unix", socketPath, 3*time.Second)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to daemon at %s: %w", socketPath, err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return nil, err
	}

	if _, err := conn.Write(raw); err != nil {
		return nil, fmt.Errorf("write to daemon failed: %w", err)
	}

	var result json.RawMessage
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("read from daemon failed: %w", err)
	}

	return []byte(result), nil
}

func RequestStats(socketPath string) (map[string]interface{}, error) {
	if socketPath == "" {
		socketPath = DefaultSocket()
	}

	conn, err := net.DialTimeout("unix", socketPath, 3*time.Second)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to daemon: %w", err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return nil, err
	}

	req := map[string]string{"type": "stats"}
	raw, _ := json.Marshal(req)
	if _, err := conn.Write(raw); err != nil {
		return nil, fmt.Errorf("write to daemon failed: %w", err)
	}

	var result map[string]interface{}
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("read from daemon failed: %w", err)
	}

	return result, nil
}
