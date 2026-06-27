package adapters

import (
	"encoding/json"
	"fmt"
	"time"
)

// CopilotAdapter translates GitHub Copilot CLI's hook JSON format.
// Copilot CLI exposes a full hook event list (sessionStart, sessionEnd,
// preToolUse, postToolUse, userPromptSubmitted, etc.) with a slightly
// different field schema. The core handles preToolUse and postToolUse.

type CopilotAdapter struct{}

type copilotHookEvent struct {
	EventID      string                 `json:"event_id"`
	EventKind    string                 `json:"event_kind"`
	Tool         string                 `json:"tool"`
	ToolInput    map[string]interface{} `json:"tool_input"`
	ToolOutput   interface{}            `json:"tool_output,omitempty"`
	WorkspaceDir string                 `json:"workspace_dir"`
	SessionID    string                 `json:"session_id,omitempty"`
	MachineID    string                 `json:"machine_id,omitempty"`
}

type copilotHookResult struct {
	PermissionDecision string                 `json:"permission_decision"`
	UpdatedInput       map[string]interface{} `json:"updated_input,omitempty"`
	UpdatedToolOutput  interface{}            `json:"updated_tool_output,omitempty"`
	CompressionMeta    string                 `json:"compression_meta,omitempty"`
}

func (a *CopilotAdapter) DecodeHookEvent(raw []byte) (*HookEvent, error) {
	var hook copilotHookEvent
	if err := json.Unmarshal(raw, &hook); err != nil {
		return nil, fmt.Errorf("copilot: failed to unmarshal JSON: %w", err)
	}

	if hook.EventKind == "" {
		return nil, fmt.Errorf("copilot: missing event_kind")
	}

	event := &HookEvent{
		HostType:    HostCopilot,
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

func (a *CopilotAdapter) EncodeHookResult(result *HookResult) ([]byte, error) {
	resp := copilotHookResult{
		PermissionDecision: result.PermissionDecision,
		UpdatedInput:       result.UpdatedInput,
		UpdatedToolOutput:  result.UpdatedToolOutput,
		CompressionMeta:    result.CompressionMeta,
	}

	return json.Marshal(resp)
}

func (a *CopilotAdapter) DecodePostCallOutput(raw []byte) (interface{}, error) {
	return string(raw), nil
}

func (a *CopilotAdapter) EncodePostCallOutput(compressed interface{}) ([]byte, error) {
	if s, ok := compressed.(string); ok {
		return []byte(s), nil
	}
	if b, ok := compressed.([]byte); ok {
		return b, nil
	}
	return json.Marshal(compressed)
}

func NewCopilotAdapter() *CopilotAdapter {
	return &CopilotAdapter{}
}
