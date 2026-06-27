package store

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// CCRCache is the content-addressed retrieval cache using SQLite.
// It stores compressed content's original for on-demand retrieval.
type CCRCache struct {
	db     *sql.DB
	dir    string
	ttl    time.Duration
	maxMB  int64
	mu     sync.RWMutex
}

// CacheEntry represents a cached item.
type CacheEntry struct {
	Handle              string
	Original            []byte
	Tool                string
	Path                string
	CompressionType     string
	CreatedAt           time.Time
	AccessedAt          time.Time
	SizeOriginal        int64
	SizeCompressed      int64
}

// NewCCRCache creates a new cache.
func NewCCRCache(dir string, ttl time.Duration, maxMB int64) (*CCRCache, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create cache dir: %w", err)
	}

	dbPath := filepath.Join(dir, "cache.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open cache database: %w", err)
	}

	cache := &CCRCache{
		db:    db,
		dir:   dir,
		ttl:   ttl,
		maxMB: maxMB,
	}

	// Initialize schema.
	if err := cache.initSchema(); err != nil {
		db.Close()
		return nil, err
	}

	return cache, nil
}

// initSchema creates the cache table if it doesn't exist.
func (c *CCRCache) initSchema() error {

	for _, stmt := range []string{
		`CREATE TABLE IF NOT EXISTS ccr_cache (
			handle TEXT PRIMARY KEY,
			original BLOB,
			tool TEXT,
			path TEXT,
			compression_type TEXT,
			created_at TIMESTAMP,
			accessed_at TIMESTAMP,
			size_original INTEGER,
			size_compressed INTEGER
		)`,
		`CREATE INDEX IF NOT EXISTS idx_path_tool ON ccr_cache(path, tool)`,
		`CREATE INDEX IF NOT EXISTS idx_created_at ON ccr_cache(created_at)`,
	} {
		if _, err := c.db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to initialize schema: %w", err)
		}
	}

	return nil
}

// Insert stores a cache entry.
func (c *CCRCache) Insert(original []byte, tool, path, compressionType string, sizeCompressed int64) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	handle := generateHandle(original)
	now := time.Now().UTC()

	query := `
	INSERT OR REPLACE INTO ccr_cache
	(handle, original, tool, path, compression_type, created_at, accessed_at, size_original, size_compressed)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := c.db.Exec(query, handle, original, tool, path, compressionType, now, now, len(original), sizeCompressed)
	if err != nil {
		return "", fmt.Errorf("failed to insert cache entry: %w", err)
	}

	return handle, nil
}

// Get retrieves a cache entry by handle.
func (c *CCRCache) Get(handle string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var original []byte
	err := c.db.QueryRow(`SELECT original FROM ccr_cache WHERE handle = ?`, handle).Scan(&original)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("handle not found: %s", handle)
		}
		return nil, fmt.Errorf("failed to retrieve cache entry: %w", err)
	}

	_, _ = c.db.Exec(`UPDATE ccr_cache SET accessed_at = ? WHERE handle = ?`, time.Now().UTC(), handle)
	return original, nil
}

// GetRange retrieves a portion of cached content (for slicing).
func (c *CCRCache) GetRange(handle string, start, end int64) ([]byte, error) {
	original, err := c.Get(handle)
	if err != nil {
		return nil, err
	}

	if start < 0 || end < 0 || start > end || int64(len(original)) < end {
		return nil, fmt.Errorf("invalid range: [%d, %d] for content of size %d", start, end, len(original))
	}

	return original[start:end], nil
}

// InvalidateByPath removes all cache entries for a given file path.
func (c *CCRCache) InvalidateByPath(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	query := `DELETE FROM ccr_cache WHERE path = ?`
	_, err := c.db.Exec(query, path)
	return err
}

// Cleanup removes expired entries and enforces size limits.
func (c *CCRCache) Cleanup() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove expired entries (older than TTL).
	expireTime := time.Now().UTC().Add(-c.ttl)
	query := `DELETE FROM ccr_cache WHERE created_at < ?`
	_, err := c.db.Exec(query, expireTime)
	if err != nil {
		return fmt.Errorf("failed to clean expired entries: %w", err)
	}

	// Enforce size limit (LRU eviction).
	if c.maxMB > 0 {
		sizeMB, err := c.currentSize()
		if err != nil {
			return err
		}

		if sizeMB > c.maxMB {
			// Delete oldest 10% of entries.

			query := `
			DELETE FROM ccr_cache WHERE handle IN (
				SELECT handle FROM ccr_cache
				ORDER BY accessed_at ASC
				LIMIT (SELECT COUNT(*) FROM ccr_cache) / 10
			)
			`
			_, err := c.db.Exec(query)
			if err != nil {
				return fmt.Errorf("failed to evict old entries: %w", err)
			}
		}
	}

	return nil
}

// currentSize returns the current cache size in MB.
func (c *CCRCache) currentSize() (int64, error) {
	var totalBytes int64
	query := `SELECT COALESCE(SUM(size_original), 0) FROM ccr_cache`
	err := c.db.QueryRow(query).Scan(&totalBytes)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate cache size: %w", err)
	}
	return totalBytes / (1024 * 1024), nil
}

// SizeMB returns the current cache size in MB.
func (c *CCRCache) SizeMB() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	size, _ := c.currentSize()
	return size
}

// EntryCount returns the number of cached entries.
func (c *CCRCache) EntryCount() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var count int64
	_ = c.db.QueryRow(`SELECT COUNT(*) FROM ccr_cache`).Scan(&count)
	return count
}

// List returns all cache entries (for audit/debugging).
func (c *CCRCache) List() ([]CacheEntry, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	query := `
	SELECT handle, tool, path, compression_type, created_at, accessed_at, size_original, size_compressed
	FROM ccr_cache
	ORDER BY accessed_at DESC
	`

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list cache entries: %w", err)
	}
	defer rows.Close()

	var entries []CacheEntry
	for rows.Next() {
		var e CacheEntry
		if err := rows.Scan(&e.Handle, &e.Tool, &e.Path, &e.CompressionType, &e.CreatedAt, &e.AccessedAt, &e.SizeOriginal, &e.SizeCompressed); err != nil {
			return nil, fmt.Errorf("failed to scan cache entry: %w", err)
		}
		entries = append(entries, e)
	}

	return entries, rows.Err()
}

// Purge deletes all cache entries.
func (c *CCRCache) Purge() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.db.Exec(`DELETE FROM ccr_cache`)
	return err
}

// Close closes the database.
func (c *CCRCache) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// generateHandle creates a content-addressed handle from content.
func generateHandle(content []byte) string {
	h := sha256.Sum256(content)
	return fmt.Sprintf("h_%x", h[:4]) // first 4 bytes, 8 hex chars
}
