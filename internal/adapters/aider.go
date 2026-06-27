package adapters

import (
	"encoding/json"
	"fmt"
	"time"
)

// AiderAdapter translates Aider CLI's hook JSON to/from HookEvent and HookResult.
// Aider is a CLI-based AI pair programmer that integrates with Claude via subprocess.
// It uses snake_case JSON similar to Codex, but with simpler structure.

type AiderAdapter struct{}

type aiderHookEvent struct {
	EventID      string                 `json:"event_id"`
	EventKind    string                 `json:"event_kind"`
	Tool         string                 `json:"tool"`
	ToolInput    map[string]interface{} `json:"tool_input"`
	ToolOutput   interface{}            `json:"tool_output,omitempty"`
	WorkspaceDir string                 `json:"workspace_dir"`
	SessionID    string                 `json:"session_id,omitempty"`
	MachineID    string                 `json:"machine_id,omitempty"`
}

type aiderHookResult struct {
	PermissionDecision string                 `json:"permission_decision"`
	UpdatedInput       map[string]interface{} `json:"updated_input,omitempty"`
	UpdatedToolOutput  interface{}            `json:"updated_tool_output,omitempty"`
	CompressionMeta    string                 `json:"compression_meta,omitempty"`
}

func (a *AiderAdapter) DecodeHookEvent(raw []byte) (*HookEvent, error) {
	var hook aiderHookEvent
	if err := json.Unmarshal(raw, &hook); err != nil {
		return nil, fmt.Errorf("aider: failed to unmarshal JSON: %w", err)
	}

	if hook.EventKind == "" {
		return nil, fmt.Errorf("aider: missing event_kind")
	}

	event := &HookEvent{
		HostType:    "aider",
		EventID:     hook.EventID,
		SessionID:   hook.SessionID,
		EventKind:   hook.EventKind,
		Tool:        hook.Tool,
		ToolInput:   hook.ToolInput,
		ToolOutput:  hook.ToolOutput,
		Workspace:   hook.WorkspaceDir,
		MachineID:   hook.MachineID,
		Timestamp:   time.Now().UTC(),
		HostSpecific: map[string]interface{}{
			"raw": string(raw),
		},
	}

	return event, nil
}

func (a *AiderAdapter) EncodeHookResult(result *HookResult) ([]byte, error) {
	resp := aiderHookResult{
		PermissionDecision: result.PermissionDecision,
		UpdatedInput:       result.UpdatedInput,
		UpdatedToolOutput:  result.UpdatedToolOutput,
		CompressionMeta:    result.CompressionMeta,
	}

	return json.Marshal(resp)
}

func (a *AiderAdapter) DecodePostCallOutput(raw []byte) (interface{}, error) {
	return string(raw), nil
}

func (a *AiderAdapter) EncodePostCallOutput(compressed interface{}) ([]byte, error) {
	if s, ok := compressed.(string); ok {
		return []byte(s), nil
	}
	if b, ok := compressed.([]byte); ok {
		return b, nil
	}
	return json.Marshal(compressed)
}

func NewAiderAdapter() *AiderAdapter {
	return &AiderAdapter{}
}
