package adapters

import (
	"encoding/json"
	"time"
)

// HookEvent is the shared internal representation of a pre-call hook event
// across all Family-A hosts (Claude Code, Codex, Antigravity, Copilot CLI).
// Each host adapter translates its native JSON to/from this canonical form.
type HookEvent struct {
	// HostType identifies which host emitted this event.
	// One of: "claudecode", "codex", "antigravity", "copilot"
	HostType string `json:"host_type"`

	// EventID is a unique identifier for this hook invocation, used for
	// logging and session tracking.
	EventID string `json:"event_id"`

	// SessionID groups all hook calls within one editing session.
	SessionID string `json:"session_id"`

	// EventKind is the phase of the tool call.
	// One of: "preToolUse", "postToolUse", "sessionStart", "sessionEnd"
	EventKind string `json:"event_kind"`

	// Tool is the name of the tool being called (e.g. "read", "bash", "apply_patch").
	Tool string `json:"tool"`

	// ToolInput is the request before compression. For reads, it includes the path.
	// For bash, the command. Structure varies by tool; the router detects type.
	ToolInput map[string]interface{} `json:"tool_input"`

	// ToolOutput is the result (post-call only). For reads, the file body.
	// For bash, stdout+stderr. May be very large; compression runs here.
	ToolOutput interface{} `json:"tool_output"`

	// Metadata about the host and session context.
	Workspace string `json:"workspace"` // abs path to the workspace root
	MachineID string `json:"machine_id"` // stable across sessions for CCR scoping

	// Timestamp when the hook fired (UTC).
	Timestamp time.Time `json:"timestamp"`

	// HostSpecific holds host-specific fields that don't fit the common schema.
	// Adapters may inject fields here for logging, but the core compressor
	// ignores this and processes the normalized fields above.
	HostSpecific map[string]interface{} `json:"host_specific,omitempty"`
}

// HookResult is the shared internal representation of the hook's response,
// which the adapter translates back to the host's native JSON.
type HookResult struct {
	// PermissionDecision is "allow" or "deny". If "deny", the tool is blocked.
	PermissionDecision string `json:"permission_decision"` // "allow" | "deny"

	// UpdatedInput, if non-nil, replaces ToolEvent.ToolInput before execution.
	// Pre-call only. May be the same as input (no change) or optimized
	// (e.g., a diff-only read range).
	UpdatedInput map[string]interface{} `json:"updated_input,omitempty"`

	// UpdatedToolOutput, if non-nil, replaces the tool result in context.
	// Post-call only. Typically compressed + a retrieve handle.
	UpdatedToolOutput interface{} `json:"updated_tool_output,omitempty"`

	// CompressionMeta is a human-readable hint about what was done
	// (e.g., "[slash: 45.2k → 3.1k tokens · retrieve(h_abc123)]").
	// Shown in session logs and context to the model as a side-channel hint.
	CompressionMeta string `json:"compression_meta,omitempty"`

	// HostSpecific is the inverse of HookEvent.HostSpecific: fields the
	// adapter adds for translation back to the host's native schema.
	HostSpecific map[string]interface{} `json:"host_specific,omitempty"`
}

// HostAdapter is the interface each host implements to translate its native
// hook JSON to/from the canonical HookEvent and HookResult.
type HostAdapter interface {
	// DecodeHookEvent translates the host's native pre-call JSON (bytes)
	// into a HookEvent. Returns an error if the JSON is malformed or
	// missing required fields.
	DecodeHookEvent(raw []byte) (*HookEvent, error)

	// EncodeHookResult translates a HookResult back into the host's native
	// JSON format (bytes). The adapter injects host-specific field names
	// like "updatedInput" (Claude Code) vs. "updated_input" (Antigravity).
	EncodeHookResult(result *HookResult) ([]byte, error)

	// DecodePostCallOutput translates the host's output format (raw bytes
	// from stdout, a streaming JSON response, etc.) into the tool's actual
	// result (a string for bash, a file body for read, etc.). This is
	// host-specific because some hosts wrap tool results in metadata.
	DecodePostCallOutput(raw []byte) (interface{}, error)

	// EncodePostCallOutput is the inverse: translates a compressed tool
	// result back into the host's output format (e.g., wrapped in metadata).
	EncodePostCallOutput(compressed interface{}) ([]byte, error)
}

// RawHostEvent is the minimum contract a host must provide: one JSON object
// on stdin for pre-call and post-call hooks. The exact field names differ
// per host (camelCase vs. snake_case, "updatedInput" vs. "updated_input");
// adapters normalize these.
type RawHostEvent struct {
	// The raw JSON as received. Adapters parse this into the canonical HookEvent.
	Raw json.RawMessage
}

// HostType enum.
const (
	HostClaudeCode  = "claudecode"
	HostCodex       = "codex"
	HostAntigravity = "antigravity"
	HostCopilot     = "copilot"
	HostCursor      = "cursor"
	HostWindsurf    = "windsurf"
	HostAider       = "aider"
	HostOpenCode    = "opencode"
)

// EventKind enum.
const (
	EventPreToolUse   = "preToolUse"
	EventPostToolUse  = "postToolUse"
	EventSessionStart = "sessionStart"
	EventSessionEnd   = "sessionEnd"
)

// PermissionDecision enum.
const (
	PermissionAllow = "allow"
	PermissionDeny  = "deny"
)
