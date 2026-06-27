package track

import (
	"crypto/md5"
	"fmt"
	"sync"
)

type FileRecord struct {
	Content string
	Hash    string
	Edited  bool
}

type ReadState struct {
	mu          sync.RWMutex
	files       map[string]*FileRecord
}

func NewReadState() *ReadState {
	return &ReadState{
		files: make(map[string]*FileRecord),
	}
}

func (rs *ReadState) RecordRead(path, content string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	hash := md5Hash(content)
	if existing, ok := rs.files[path]; ok {
		existing.Content = content
		existing.Hash = hash
		existing.Edited = false
	} else {
		rs.files[path] = &FileRecord{
			Content: content,
			Hash:    hash,
			Edited:  false,
		}
	}
}

func (rs *ReadState) MarkEdited(path string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if record, ok := rs.files[path]; ok {
		record.Edited = true
	}
}

func (rs *ReadState) WasRead(path string) bool {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	_, ok := rs.files[path]
	return ok
}

func (rs *ReadState) WasEdited(path string) bool {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	if record, ok := rs.files[path]; ok {
		return record.Edited
	}
	return false
}

func (rs *ReadState) IsChanged(path string, newContent string) bool {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	if record, ok := rs.files[path]; ok {
		return record.Hash != md5Hash(newContent)
	}
	return false
}

func (rs *ReadState) GetPreviousContent(path string) (string, bool) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	if record, ok := rs.files[path]; ok {
		return record.Content, true
	}
	return "", false
}

func (rs *ReadState) GetReReadStrategy(path string, newContent string) (string, string) {
	rs.mu.RLock()
	record, seen := rs.files[path]
	rs.mu.RUnlock()

	if !seen {
		return "first_read", ""
	}

	hash := md5Hash(newContent)
	changed := record.Hash != hash

	if !changed {
		return "unchanged", record.Content
	}

	return "changed", record.Content
}

type ReadStateTracker struct {
	mu       sync.RWMutex
	sessions map[string]*ReadState
}

func NewReadStateTracker() *ReadStateTracker {
	return &ReadStateTracker{
		sessions: make(map[string]*ReadState),
	}
}

func (t *ReadStateTracker) GetOrCreate(sessionID string) *ReadState {
	t.mu.Lock()
	defer t.mu.Unlock()

	if rs, ok := t.sessions[sessionID]; ok {
		return rs
	}

	rs := NewReadState()
	t.sessions[sessionID] = rs
	return rs
}

func (t *ReadStateTracker) Delete(sessionID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.sessions, sessionID)
}

func md5Hash(s string) string {
	h := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", h)
}
