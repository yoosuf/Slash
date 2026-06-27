package adapters

import (
	"encoding/json"
	"fmt"
	"time"
)

// ClaudeCodeAdapter translates Claude Code's hook JSON to/from HookEvent and HookResult.
// Claude Code emits pre-call and post-call hooks with these shapes:
//
// Pre-call (PreToolUse):
//   {
//     "eventId": "uuid",
//     "eventKind": "preToolUse",
//     "tool": "read",
//     "toolInput": {"path": "/file.txt"},
//     "workspaceDir": "/path/to/workspace"
//   }
//
// Post-call (PostToolUse):
//   {
//     "eventId": "uuid",
//     "eventKind": "postToolUse",
//     "tool": "read",
//     "toolInput": {"path": "/file.txt"},
//     "toolOutput": "file contents...",
//     "workspaceDir": "/path/to/workspace"
//   }
//
// The adapter returns HookResult with fields:
//   "permissionDecision": "allow" | "deny"
//   "hookSpecificOutput": {
//     "updatedInput": {...},  // if pre-call and modified
//     "updatedToolOutput": ..., // if post-call and modified
//   }

type ClaudeCodeAdapter struct{}

// clueCodePreCallHook is the unmarshaled Claude Code pre-call event.
type claudeCodePreCallHook struct {
	EventID      string                 `json:"eventId"`
	EventKind    string                 `json:"eventKind"`
	Tool         string                 `json:"tool"`
	ToolInput    map[string]interface{} `json:"toolInput"`
	WorkspaceDir string                 `json:"workspaceDir"`
	SessionID    string                 `json:"sessionId,omitempty"`
	MachineID    string                 `json:"machineId,omitempty"`
}

// claudeCodePostCallHook is the unmarshaled Claude Code post-call event.
type claudeCodePostCallHook struct {
	EventID      string                 `json:"eventId"`
	EventKind    string                 `json:"eventKind"`
	Tool         string                 `json:"tool"`
	ToolInput    map[string]interface{} `json:"toolInput"`
	ToolOutput   interface{}            `json:"toolOutput"`
	WorkspaceDir string                 `json:"workspaceDir"`
	SessionID    string                 `json:"sessionId,omitempty"`
	MachineID    string                 `json:"machineId,omitempty"`
}

// claudeCodeHookResult is what Claude Code expects in the response JSON.
type claudeCodeHookResult struct {
	PermissionDecision string      `json:"permissionDecision"`
	HookSpecificOutput *struct {
		UpdatedInput      map[string]interface{} `json:"updatedInput,omitempty"`
		UpdatedToolOutput interface{}            `json:"updatedToolOutput,omitempty"`
	} `json:"hookSpecificOutput,omitempty"`
}

func (a *ClaudeCodeAdapter) DecodeHookEvent(raw []byte) (*HookEvent, error) {
	// Try to unmarshal into the pre-call or post-call shape.
	// Both have eventKind, so we can check that to decide.
	var generic map[string]interface{}
	if err := json.Unmarshal(raw, &generic); err != nil {
		return nil, fmt.Errorf("claudecode: failed to unmarshal JSON: %w", err)
	}

	eventKind, ok := generic["eventKind"].(string)
	if !ok {
		return nil, fmt.Errorf("claudecode: missing or non-string eventKind")
	}

	_, ok = generic["eventId"].(string)
	if !ok {
		return nil, fmt.Errorf("claudecode: missing or non-string eventId")
	}

	var event *HookEvent

	if eventKind == EventPostToolUse {
		// Post-call event includes toolOutput.
		var hook claudeCodePostCallHook
		if err := json.Unmarshal(raw, &hook); err != nil {
			return nil, fmt.Errorf("claudecode: failed to unmarshal post-call event: %w", err)
		}
		event = &HookEvent{
			HostType:    HostClaudeCode,
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
		// Pre-call event; no toolOutput.
		var hook claudeCodePreCallHook
		if err := json.Unmarshal(raw, &hook); err != nil {
			return nil, fmt.Errorf("claudecode: failed to unmarshal pre-call event: %w", err)
		}
		event = &HookEvent{
			HostType:    HostClaudeCode,
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
		return nil, fmt.Errorf("claudecode: unrecognized eventKind: %s", eventKind)
	}

	return event, nil
}

func (a *ClaudeCodeAdapter) EncodeHookResult(result *HookResult) ([]byte, error) {
	// Claude Code expects the response shape:
	//   {
	//     "permissionDecision": "allow" | "deny",
	//     "hookSpecificOutput": {
	//       "updatedInput": {...},          // if modified
	//       "updatedToolOutput": ...,       // if modified
	//     }
	//   }

	resp := claudeCodeHookResult{
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

func (a *ClaudeCodeAdapter) DecodePostCallOutput(raw []byte) (interface{}, error) {
	// Claude Code returns the raw tool output as-is (string, JSON, or binary).
	// For now, we just return it as a string.
	return string(raw), nil
}

func (a *ClaudeCodeAdapter) EncodePostCallOutput(compressed interface{}) ([]byte, error) {
	// Convert back to bytes.
	if s, ok := compressed.(string); ok {
		return []byte(s), nil
	}
	// If it's already bytes, return as-is.
	if b, ok := compressed.([]byte); ok {
		return b, nil
	}
	// Otherwise JSON-encode it.
	return json.Marshal(compressed)
}

// NewClaudeCodeAdapter creates a new Claude Code adapter.
func NewClaudeCodeAdapter() *ClaudeCodeAdapter {
	return &ClaudeCodeAdapter{}
}
