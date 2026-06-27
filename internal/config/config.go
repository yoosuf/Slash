package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Daemon      DaemonConfig      `json:"daemon,omitempty"`
	Compression CompressionConfig `json:"compression,omitempty"`
	Cache       CacheConfig       `json:"cache,omitempty"`
	Telemetry   TelemetryConfig   `json:"telemetry,omitempty"`
}

type DaemonConfig struct {
	Socket   string `json:"socket,omitempty"`
	LogLevel string `json:"log_level,omitempty"`
}

type CompressionConfig struct {
	Enabled        bool `json:"enabled,omitempty"`
	DiffOnlyReads  bool `json:"diff_only_reads,omitempty"`
	OutputCompress bool `json:"output_compress,omitempty"`
	RepoMapInject  bool `json:"repo_map_inject,omitempty"`
}

type CacheConfig struct {
	Dir     string `json:"dir,omitempty"`
	TTL     string `json:"ttl_hours,omitempty"`
	MaxSize int64  `json:"max_size_mb,omitempty"`
}

type TelemetryConfig struct {
	Enabled       bool   `json:"enabled,omitempty"`
	AggregateOnly bool   `json:"aggregate_only,omitempty"`
	Events        string `json:"events,omitempty"`
}

func DefaultConfig() *Config {
	return &Config{
		Daemon: DaemonConfig{LogLevel: "warn"},
		Compression: CompressionConfig{
			Enabled:        true,
			DiffOnlyReads:  true,
			OutputCompress: true,
			RepoMapInject:  false,
		},
		Cache: CacheConfig{TTL: "24", MaxSize: 1024},
		Telemetry: TelemetryConfig{
			Enabled:       false,
			AggregateOnly: true,
		},
	}
}

func Load() *Config {
	cfg := DefaultConfig()

	home, err := os.UserHomeDir()
	if err != nil {
		return cfg
	}

	path := filepath.Join(home, ".slash", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	var fileCfg Config
	if err := json.Unmarshal(data, &fileCfg); err != nil {
		fmt.Fprintf(os.Stderr, "[slash] warning: failed to parse %s: %v\n", path, err)
		return cfg
	}

	merge(cfg, &fileCfg)
	return cfg
}

func merge(dst *Config, src *Config) {
	if src.Daemon.Socket != "" {
		dst.Daemon.Socket = src.Daemon.Socket
	}
	if src.Daemon.LogLevel != "" {
		dst.Daemon.LogLevel = src.Daemon.LogLevel
	}
	if src.Compression.Enabled {
		dst.Compression.Enabled = src.Compression.Enabled
	}
	if src.Cache.Dir != "" {
		dst.Cache.Dir = src.Cache.Dir
	}
	if src.Cache.TTL != "" {
		dst.Cache.TTL = src.Cache.TTL
	}
	if src.Cache.MaxSize > 0 {
		dst.Cache.MaxSize = src.Cache.MaxSize
	}
}

func (c *Config) CacheTTL() time.Duration {
	h := 24
	if c.Cache.TTL != "" {
		_, _ = fmt.Sscanf(c.Cache.TTL, "%d", &h)
	}
	return time.Duration(h) * time.Hour
}
