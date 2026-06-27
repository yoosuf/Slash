package adapters

import (
	"encoding/json"
	"fmt"
	"time"
)

// AntigravityAdapter translates Antigravity (agy) CLI's hooks.json format
// to/from HookEvent and HookResult.
//
// Antigravity hooks fire pre-flight (before execution) and post-flight
// (after execution) with structured JSON. The adapter treats these the same
// as Claude Code's preToolUse / postToolUse.

type AntigravityAdapter struct{}

type antigravityHookEvent struct {
	EventID      string                 `json:"event_id"`
	EventKind    string                 `json:"event_kind"`
	Tool         string                 `json:"tool"`
	ToolInput    map[string]interface{} `json:"tool_input"`
	ToolOutput   interface{}            `json:"tool_output,omitempty"`
	WorkspaceDir string                 `json:"workspace_dir"`
	SessionID    string                 `json:"session_id,omitempty"`
	MachineID    string                 `json:"machine_id,omitempty"`
}

type antigravityHookResult struct {
	PermissionDecision string                 `json:"permission_decision"`
	UpdatedInput       map[string]interface{} `json:"updated_input,omitempty"`
	UpdatedToolOutput  interface{}            `json:"updated_tool_output,omitempty"`
	CompressionMeta    string                 `json:"compression_meta,omitempty"`
}

func (a *AntigravityAdapter) DecodeHookEvent(raw []byte) (*HookEvent, error) {
	var hook antigravityHookEvent
	if err := json.Unmarshal(raw, &hook); err != nil {
		return nil, fmt.Errorf("antigravity: failed to unmarshal JSON: %w", err)
	}

	if hook.EventKind == "" {
		return nil, fmt.Errorf("antigravity: missing event_kind")
	}

	event := &HookEvent{
		HostType:    HostAntigravity,
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

func (a *AntigravityAdapter) EncodeHookResult(result *HookResult) ([]byte, error) {
	resp := antigravityHookResult{
		PermissionDecision: result.PermissionDecision,
		UpdatedInput:       result.UpdatedInput,
		UpdatedToolOutput:  result.UpdatedToolOutput,
		CompressionMeta:    result.CompressionMeta,
	}

	return json.Marshal(resp)
}

func (a *AntigravityAdapter) DecodePostCallOutput(raw []byte) (interface{}, error) {
	return string(raw), nil
}

func (a *AntigravityAdapter) EncodePostCallOutput(compressed interface{}) ([]byte, error) {
	if s, ok := compressed.(string); ok {
		return []byte(s), nil
	}
	if b, ok := compressed.([]byte); ok {
		return b, nil
	}
	return json.Marshal(compressed)
}

func NewAntigravityAdapter() *AntigravityAdapter {
	return &AntigravityAdapter{}
}
