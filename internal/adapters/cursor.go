package adapters

import (
	"encoding/json"
	"fmt"
	"time"
)

// CursorAdapter translates Cursor's hook JSON to/from HookEvent and HookResult.
// Cursor is built on VSCode and uses a similar hook model to Claude Code.
// Hook format is nearly identical to Claude Code (camelCase JSON).

type CursorAdapter struct{}

type cursorHookEvent struct {
	EventID      string                 `json:"eventId"`
	EventKind    string                 `json:"eventKind"`
	Tool         string                 `json:"tool"`
	ToolInput    map[string]interface{} `json:"toolInput"`
	ToolOutput   interface{}            `json:"toolOutput,omitempty"`
	WorkspaceDir string                 `json:"workspaceDir"`
	SessionID    string                 `json:"sessionId,omitempty"`
	MachineID    string                 `json:"machineId,omitempty"`
}

type cursorHookResult struct {
	PermissionDecision string      `json:"permissionDecision"`
	HookSpecificOutput *struct {
		UpdatedInput      map[string]interface{} `json:"updatedInput,omitempty"`
		UpdatedToolOutput interface{}            `json:"updatedToolOutput,omitempty"`
	} `json:"hookSpecificOutput,omitempty"`
}

func (a *CursorAdapter) DecodeHookEvent(raw []byte) (*HookEvent, error) {
	var generic map[string]interface{}
	if err := json.Unmarshal(raw, &generic); err != nil {
		return nil, fmt.Errorf("cursor: failed to unmarshal JSON: %w", err)
	}

	eventKind, ok := generic["eventKind"].(string)
	if !ok {
		return nil, fmt.Errorf("cursor: missing or non-string eventKind")
	}

	_, ok = generic["eventId"].(string)
	if !ok {
		return nil, fmt.Errorf("cursor: missing or non-string eventId")
	}

	var event *HookEvent

	if eventKind == EventPostToolUse {
		var hook cursorHookEvent
		if err := json.Unmarshal(raw, &hook); err != nil {
			return nil, fmt.Errorf("cursor: failed to unmarshal post-call event: %w", err)
		}
		event = &HookEvent{
			HostType:    "cursor",
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
	} else if eventKind == EventPreToolUse {
		var hook cursorHookEvent
		if err := json.Unmarshal(raw, &hook); err != nil {
			return nil, fmt.Errorf("cursor: failed to unmarshal pre-call event: %w", err)
		}
		event = &HookEvent{
			HostType:    "cursor",
			EventID:     hook.EventID,
			SessionID:   hook.SessionID,
			EventKind:   hook.EventKind,
			Tool:        hook.Tool,
			ToolInput:   hook.ToolInput,
			Workspace:   hook.WorkspaceDir,
			MachineID:   hook.MachineID,
			Timestamp:   time.Now().UTC(),
			HostSpecific: map[string]interface{}{
				"raw": string(raw),
			},
		}
	} else {
		return nil, fmt.Errorf("cursor: unrecognized eventKind: %s", eventKind)
	}

	return event, nil
}

func (a *CursorAdapter) EncodeHookResult(result *HookResult) ([]byte, error) {
	resp := cursorHookResult{
		PermissionDecision: result.PermissionDecision,
	}

	if result.UpdatedInput != nil || result.UpdatedToolOutput != nil {
		resp.HookSpecificOutput = &struct {
			UpdatedInput      map[string]interface{} `json:"updatedInput,omitempty"`
			UpdatedToolOutput interface{}            `json:"updatedToolOutput,omitempty"`
		}{
			UpdatedInput:      result.UpdatedInput,
			UpdatedToolOutput: result.UpdatedToolOutput,
		}
	}

	return json.Marshal(resp)
}

func (a *CursorAdapter) DecodePostCallOutput(raw []byte) (interface{}, error) {
	return string(raw), nil
}

func (a *CursorAdapter) EncodePostCallOutput(compressed interface{}) ([]byte, error) {
	if s, ok := compressed.(string); ok {
		return []byte(s), nil
	}
	if b, ok := compressed.([]byte); ok {
		return b, nil
	}
	return json.Marshal(compressed)
}

func NewCursorAdapter() *CursorAdapter {
	return &CursorAdapter{}
}
