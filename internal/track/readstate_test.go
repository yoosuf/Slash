package track

import (
	"testing"
)

func TestNewReadState(t *testing.T) {
	rs := NewReadState()
	if rs == nil {
		t.Fatal("expected non-nil ReadState")
	}
}

func TestRecordRead(t *testing.T) {
	rs := NewReadState()
	rs.RecordRead("test.go", "package main\nfunc main() {}")

	if !rs.WasRead("test.go") {
		t.Fatal("expected test.go to be marked as read")
	}

	if rs.WasRead("nonexistent.go") {
		t.Fatal("nonexistent.go should not be marked as read")
	}
}

func TestWasEdited(t *testing.T) {
	rs := NewReadState()
	rs.RecordRead("test.go", "original content")
	rs.MarkEdited("test.go")

	if !rs.WasEdited("test.go") {
		t.Fatal("expected test.go to be marked as edited")
	}
}

func TestIsChanged(t *testing.T) {
	rs := NewReadState()
	rs.RecordRead("test.go", "original")

	if rs.IsChanged("test.go", "original") {
		t.Fatal("expected unchanged content to not be flagged as changed")
	}

	if !rs.IsChanged("test.go", "modified") {
		t.Fatal("expected modified content to be flagged as changed")
	}
}

func TestGetPreviousContent(t *testing.T) {
	rs := NewReadState()
	rs.RecordRead("test.go", "previous content")

	content, ok := rs.GetPreviousContent("test.go")
	if !ok {
		t.Fatal("expected to find previous content")
	}
	if content != "previous content" {
		t.Fatalf("expected 'previous content', got '%s'", content)
	}

	_, ok = rs.GetPreviousContent("unknown.go")
	if ok {
		t.Fatal("expected not to find content for unknown file")
	}
}

func TestRecordRead_Overwrites(t *testing.T) {
	rs := NewReadState()
	rs.RecordRead("test.go", "first")
	rs.RecordRead("test.go", "second")

	content, _ := rs.GetPreviousContent("test.go")
	if content != "second" {
		t.Fatalf("expected 'second', got '%s'", content)
	}
}

func TestMarkEdited_ThenRecordRead_ResetsEdited(t *testing.T) {
	rs := NewReadState()
	rs.RecordRead("test.go", "content")
	rs.MarkEdited("test.go")

	if !rs.WasEdited("test.go") {
		t.Fatal("expected file to be edited after MarkEdited")
	}

	rs.RecordRead("test.go", "new content")

	if rs.WasEdited("test.go") {
		t.Fatal("expected edited flag to be reset after RecordRead")
	}
}

func TestGetReReadStrategy_FirstRead(t *testing.T) {
	rs := NewReadState()
	strategy, _ := rs.GetReReadStrategy("new.go", "content")
	if strategy != "first_read" {
		t.Fatalf("expected 'first_read', got '%s'", strategy)
	}
}

func TestGetReReadStrategy_Unchanged(t *testing.T) {
	rs := NewReadState()
	rs.RecordRead("test.go", "content")
	strategy, _ := rs.GetReReadStrategy("test.go", "content")
	if strategy != "unchanged" {
		t.Fatalf("expected 'unchanged', got '%s'", strategy)
	}
}

func TestGetReReadStrategy_Changed(t *testing.T) {
	rs := NewReadState()
	rs.RecordRead("test.go", "old content")
	strategy, prevContent := rs.GetReReadStrategy("test.go", "new content")
	if strategy != "changed" {
		t.Fatalf("expected 'changed', got '%s'", strategy)
	}
	if prevContent != "old content" {
		t.Fatalf("expected 'old content', got '%s'", prevContent)
	}
}

func TestNewReadStateTracker(t *testing.T) {
	tracker := NewReadStateTracker()
	if tracker == nil {
		t.Fatal("expected non-nil tracker")
	}
}

func TestGetOrCreate(t *testing.T) {
	tracker := NewReadStateTracker()

	rs1 := tracker.GetOrCreate("session-1")
	if rs1 == nil {
		t.Fatal("expected non-nil ReadState")
	}

	rs2 := tracker.GetOrCreate("session-1")
	if rs1 != rs2 {
		t.Fatal("expected same ReadState for same session ID")
	}

	rs3 := tracker.GetOrCreate("session-2")
	if rs1 == rs3 {
		t.Fatal("expected different ReadState for different session IDs")
	}
}

func TestDelete(t *testing.T) {
	tracker := NewReadStateTracker()
	tracker.GetOrCreate("session-1")
	tracker.Delete("session-1")

	rs := tracker.GetOrCreate("session-1")
	if rs == nil {
		t.Fatal("expected a new ReadState after deletion")
	}
}
