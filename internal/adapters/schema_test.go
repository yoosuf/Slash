package adapters

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestClaudeCodeAdapterPreCall validates that the adapter can decode
// a real Claude Code pre-call event and encode a result back.
func TestClaudeCodeAdapterPreCall(t *testing.T) {
	adapter := NewClaudeCodeAdapter()

	// Load fixture.
	fixture, err := os.ReadFile(filepath.Join("fixtures", "claudecode_precall.json"))
	if err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}

	// Decode.
	event, err := adapter.DecodeHookEvent(fixture)
	if err != nil {
		t.Fatalf("DecodeHookEvent failed: %v", err)
	}

	// Validate.
	if event.HostType != HostClaudeCode {
		t.Errorf("HostType: expected %s, got %s", HostClaudeCode, event.HostType)
	}
	if event.EventKind != EventPreToolUse {
		t.Errorf("EventKind: expected %s, got %s", EventPreToolUse, event.EventKind)
	}
	if event.Tool != "read" {
		t.Errorf("Tool: expected read, got %s", event.Tool)
	}
	if event.EventID != "evt_550e8400e29b41d4a716446655440000" {
		t.Errorf("EventID mismatch")
	}

	// Check ToolInput.
	if path, ok := event.ToolInput["path"]; !ok || path != "/Users/dev/myproject/src/main.go" {
		t.Errorf("ToolInput path mismatch")
	}

	// Encode a result.
	result := &HookResult{
		PermissionDecision: PermissionAllow,
	}
	encoded, err := adapter.EncodeHookResult(result)
	if err != nil {
		t.Fatalf("EncodeHookResult failed: %v", err)
	}

	// Verify it's valid JSON.
	var output map[string]interface{}
	if err := json.Unmarshal(encoded, &output); err != nil {
		t.Fatalf("encoded result is not valid JSON: %v", err)
	}

	if perm, ok := output["permissionDecision"].(string); !ok || perm != PermissionAllow {
		t.Errorf("permissionDecision not set correctly in output")
	}
}

// TestClaudeCodeAdapterPostCall validates post-call decoding and encoding.
func TestClaudeCodeAdapterPostCall(t *testing.T) {
	adapter := NewClaudeCodeAdapter()

	// Load fixture.
	fixture, err := os.ReadFile(filepath.Join("fixtures", "claudecode_postcall.json"))
	if err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}

	// Decode.
	event, err := adapter.DecodeHookEvent(fixture)
	if err != nil {
		t.Fatalf("DecodeHookEvent failed: %v", err)
	}

	// Validate.
	if event.EventKind != EventPostToolUse {
		t.Errorf("EventKind: expected %s, got %s", EventPostToolUse, event.EventKind)
	}
	if output, ok := event.ToolOutput.(string); !ok || len(output) == 0 {
		t.Errorf("ToolOutput is empty or not a string")
	}

	// Encode a compressed result.
	result := &HookResult{
		PermissionDecision:  PermissionAllow,
		UpdatedToolOutput:   "[slash: 456 → 32 tokens · retrieve(h_abc123def)]",
		CompressionMeta:     "[slash: 456 → 32 tokens]",
	}
	encoded, err := adapter.EncodeHookResult(result)
	if err != nil {
		t.Fatalf("EncodeHookResult failed: %v", err)
	}

	// Verify structure.
	var output claudeCodeHookResult
	if err := json.Unmarshal(encoded, &output); err != nil {
		t.Fatalf("encoded result is not valid JSON: %v", err)
	}

	if output.PermissionDecision != PermissionAllow {
		t.Errorf("permissionDecision not set")
	}

	if output.HookSpecificOutput == nil {
		t.Errorf("HookSpecificOutput is nil")
	} else if compressed, ok := output.HookSpecificOutput.UpdatedToolOutput.(string); !ok || len(compressed) == 0 {
		t.Errorf("UpdatedToolOutput not set in HookSpecificOutput")
	}
}

// TestSchemaConsistency ensures that HookEvent and HookResult are well-formed.
func TestSchemaConsistency(t *testing.T) {
	// Just verify the structs can be marshaled/unmarshaled.
	event := &HookEvent{
		HostType:    HostClaudeCode,
		EventID:     "test_id",
		SessionID:   "sess_id",
		EventKind:   EventPreToolUse,
		Tool:        "read",
		ToolInput:   map[string]interface{}{"path": "/test"},
		Workspace:   "/workspace",
		MachineID:   "mach_id",
		HostSpecific: map[string]interface{}{},
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal HookEvent: %v", err)
	}

	var decoded HookEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal HookEvent: %v", err)
	}

	if decoded.HostType != event.HostType {
		t.Errorf("HostType mismatch after round-trip")
	}
}
