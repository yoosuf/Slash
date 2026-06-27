package adapters

import (
	"encoding/json"
	"fmt"
	"time"
)

type OpenCodeAdapter struct{}

type openCodePreCallHook struct {
	EventID      string                 `json:"eventId"`
	EventKind    string                 `json:"eventKind"`
	Tool         string                 `json:"tool"`
	ToolInput    map[string]interface{} `json:"toolInput"`
	WorkspaceDir string                 `json:"workspaceDir"`
	SessionID    string                 `json:"sessionId,omitempty"`
	MachineID    string                 `json:"machineId,omitempty"`
}

type openCodePostCallHook struct {
	EventID      string                 `json:"eventId"`
	EventKind    string                 `json:"eventKind"`
	Tool         string                 `json:"tool"`
	ToolInput    map[string]interface{} `json:"toolInput"`
	ToolOutput   interface{}            `json:"toolOutput"`
	WorkspaceDir string                 `json:"workspaceDir"`
	SessionID    string                 `json:"sessionId,omitempty"`
	MachineID    string                 `json:"machineId,omitempty"`
}

type openCodeHookResult struct {
	PermissionDecision string      `json:"permissionDecision"`
	HookSpecificOutput *struct {
		UpdatedInput      map[string]interface{} `json:"updatedInput,omitempty"`
		UpdatedToolOutput interface{}            `json:"updatedToolOutput,omitempty"`
	} `json:"hookSpecificOutput,omitempty"`
}

func (a *OpenCodeAdapter) DecodeHookEvent(raw []byte) (*HookEvent, error) {
	var generic map[string]interface{}
	if err := json.Unmarshal(raw, &generic); err != nil {
		return nil, fmt.Errorf("opencode: failed to unmarshal JSON: %w", err)
	}

	eventKind, ok := generic["eventKind"].(string)
	if !ok {
		return nil, fmt.Errorf("opencode: missing or non-string eventKind")
	}

	_, ok = generic["eventId"].(string)
	if !ok {
		return nil, fmt.Errorf("opencode: missing or non-string eventId")
	}

	var event *HookEvent

	if eventKind == EventPostToolUse {
		var hook openCodePostCallHook
		if err := json.Unmarshal(raw, &hook); err != nil {
			return nil, fmt.Errorf("opencode: failed to unmarshal post-call event: %w", err)
		}
		event = &HookEvent{
			HostType:   HostOpenCode,
			EventID:    hook.EventID,
			SessionID:  hook.SessionID,
			EventKind:  hook.EventKind,
			Tool:       hook.Tool,
			ToolInput:  hook.ToolInput,
			ToolOutput: hook.ToolOutput,
			Workspace:  hook.WorkspaceDir,
			MachineID:  hook.MachineID,
			Timestamp:  time.Now().UTC(),
			HostSpecific: map[string]interface{}{
				"raw": string(raw),
			},
		}
	} else if eventKind == EventPreToolUse {
		var hook openCodePreCallHook
		if err := json.Unmarshal(raw, &hook); err != nil {
			return nil, fmt.Errorf("opencode: failed to unmarshal pre-call event: %w", err)
		}
		event = &HookEvent{
			HostType:   HostOpenCode,
			EventID:    hook.EventID,
			SessionID:  hook.SessionID,
			EventKind:  hook.EventKind,
			Tool:       hook.Tool,
			ToolInput:  hook.ToolInput,
			Workspace:  hook.WorkspaceDir,
			MachineID:  hook.MachineID,
			Timestamp:  time.Now().UTC(),
			HostSpecific: map[string]interface{}{
				"raw": string(raw),
			},
		}
	} else {
		return nil, fmt.Errorf("opencode: unrecognized eventKind: %s", eventKind)
	}

	return event, nil
}

func (a *OpenCodeAdapter) EncodeHookResult(result *HookResult) ([]byte, error) {
	resp := openCodeHookResult{
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

func (a *OpenCodeAdapter) DecodePostCallOutput(raw []byte) (interface{}, error) {
	return string(raw), nil
}

func (a *OpenCodeAdapter) EncodePostCallOutput(compressed interface{}) ([]byte, error) {
	if s, ok := compressed.(string); ok {
		return []byte(s), nil
	}
	if b, ok := compressed.([]byte); ok {
		return b, nil
	}
	return json.Marshal(compressed)
}

func NewOpenCodeAdapter() *OpenCodeAdapter {
	return &OpenCodeAdapter{}
}
