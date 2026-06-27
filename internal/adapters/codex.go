package adapters

import (
	"encoding/json"
	"fmt"
	"time"
)

// CodexAdapter translates Codex's hook JSON to/from HookEvent and HookResult.
// Codex follows a similar shape to Claude Code but with subtle field-name differences
// (snake_case for some fields, and apply_patch rather than "patch" for edits).
//
// The adapter handles partial output interception (Codex does not fully support
// PostToolUse rewriting of arbitrary results, only for Bash and apply_patch).

type CodexAdapter struct{}

type codexHookEvent struct {
	EventID      string                 `json:"event_id"`
	EventKind    string                 `json:"event_kind"`
	Tool         string                 `json:"tool"`
	ToolInput    map[string]interface{} `json:"tool_input"`
	ToolOutput   interface{}            `json:"tool_output,omitempty"`
	WorkspaceDir string                 `json:"workspace_dir"`
	SessionID    string                 `json:"session_id,omitempty"`
	MachineID    string                 `json:"machine_id,omitempty"`
}

type codexHookResult struct {
	PermissionDecision string                 `json:"permission_decision"`
	UpdatedInput       map[string]interface{} `json:"updated_input,omitempty"`
	UpdatedToolOutput  interface{}            `json:"updated_tool_output,omitempty"`
}

func (a *CodexAdapter) DecodeHookEvent(raw []byte) (*HookEvent, error) {
	var hook codexHookEvent
	if err := json.Unmarshal(raw, &hook); err != nil {
		return nil, fmt.Errorf("codex: failed to unmarshal JSON: %w", err)
	}

	if hook.EventKind == "" {
		return nil, fmt.Errorf("codex: missing event_kind")
	}

	event := &HookEvent{
		HostType:    HostCodex,
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

func (a *CodexAdapter) EncodeHookResult(result *HookResult) ([]byte, error) {
	// Codex expects snake_case field names.
	resp := codexHookResult{
		PermissionDecision: result.PermissionDecision,
		UpdatedInput:       result.UpdatedInput,
		UpdatedToolOutput:  result.UpdatedToolOutput,
	}

	return json.Marshal(resp)
}

func (a *CodexAdapter) DecodePostCallOutput(raw []byte) (interface{}, error) {
	return string(raw), nil
}

func (a *CodexAdapter) EncodePostCallOutput(compressed interface{}) ([]byte, error) {
	if s, ok := compressed.(string); ok {
		return []byte(s), nil
	}
	if b, ok := compressed.([]byte); ok {
		return b, nil
	}
	return json.Marshal(compressed)
}

func NewCodexAdapter() *CodexAdapter {
	return &CodexAdapter{}
}
