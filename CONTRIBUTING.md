# Contributing to Slash

Slash is open-source (Apache 2.0) and welcomes community contributions. The most valuable way to extend the project is to add support for a new host (e.g., Cursor, Windsurf, Aider, or a future tool).

## Architecture for Contributors

All hosts converge on the same internal schema (`internal/adapters/schema.go`):
- **HookEvent**: the normalized representation of a pre-call or post-call hook.
- **HookResult**: the normalized response the compressor returns.
- **HostAdapter**: the interface each host implements.

Adding a new host = one adapter file + fixtures of real hook JSON + contract tests. The compressor core never changes.

## Adding a New Host Adapter

### Step 1: Create the Adapter File

Create `internal/adapters/newtool.go` following this template:

```go
package adapters

import (
	"encoding/json"
	"fmt"
	"time"
)

// MyToolAdapter translates MyTool's hook JSON to/from HookEvent and HookResult.
// Reference: [link to MyTool's hook documentation or spec]
//
// Hook format:
//   Pre-call: { "hookType": "preToolUse", "toolName": "read", ... }
//   Post-call: { "hookType": "postToolUse", "toolName": "read", "result": ... }

type MyToolAdapter struct{}

// myToolHookEvent is the unmarshaled native JSON structure.
type myToolHookEvent struct {
	HookType     string                 `json:"hookType"`
	ToolName     string                 `json:"toolName"`
	ToolArgs     map[string]interface{} `json:"toolArgs"`
	ToolResult   interface{}            `json:"toolResult,omitempty"`
	WorkspaceDir string                 `json:"workspace"`
	SessionID    string                 `json:"sessionId,omitempty"`
	MachineID    string                 `json:"machineId,omitempty"`
}

// myToolHookResult is the native response format.
type myToolHookResult struct {
	Decision string                 `json:"decision"`
	Modified map[string]interface{} `json:"modifiedToolArgs,omitempty"`
	NewResult interface{}           `json:"newResult,omitempty"`
}

func (a *MyToolAdapter) DecodeHookEvent(raw []byte) (*HookEvent, error) {
	var hook myToolHookEvent
	if err := json.Unmarshal(raw, &hook); err != nil {
		return nil, fmt.Errorf("mytool: failed to unmarshal JSON: %w", err)
	}

	// Map MyTool's field names to the canonical schema.
	eventKind := "preToolUse"
	if hook.HookType == "postToolUse" {
		eventKind = "postToolUse"
	}

	event := &HookEvent{
		HostType:    "mytool",  // Use a unique, lowercased identifier.
		EventID:     hook.SessionID, // Or generate a unique ID if not provided.
		SessionID:   hook.SessionID,
		EventKind:   eventKind,
		Tool:        hook.ToolName,
		ToolInput:   hook.ToolArgs,
		ToolOutput:  hook.ToolResult,
		Workspace:   hook.WorkspaceDir,
		MachineID:   hook.MachineID,
		Timestamp:   time.Now().UTC(),
		HostSpecific: map[string]interface{}{
			"raw": string(raw),
		},
	}

	return event, nil
}

func (a *MyToolAdapter) EncodeHookResult(result *HookResult) ([]byte, error) {
	// Map the canonical schema back to MyTool's format.
	resp := myToolHookResult{
		Decision:      result.PermissionDecision,
		Modified:      result.UpdatedInput,
		NewResult:     result.UpdatedToolOutput,
	}

	return json.Marshal(resp)
}

func (a *MyToolAdapter) DecodePostCallOutput(raw []byte) (interface{}, error) {
	// Adapt the raw tool output if MyTool wraps it in metadata.
	// For most tools, just return it as a string.
	return string(raw), nil
}

func (a *MyToolAdapter) EncodePostCallOutput(compressed interface{}) ([]byte, error) {
	// Inverse of DecodePostCallOutput: wrap the compressed result
	// in MyTool's expected format if needed.
	if s, ok := compressed.(string); ok {
		return []byte(s), nil
	}
	return json.Marshal(compressed)
}

func NewMyToolAdapter() *MyToolAdapter {
	return &MyToolAdapter{}
}
```

### Step 2: Add Fixtures

Create fixture files in `internal/adapters/fixtures/` with real hook JSON from your host:

**fixtures/mytool_precall.json** — a real pre-call hook event
**fixtures/mytool_postcall.json** — a real post-call hook event with output

The fixtures serve as:
1. Documentation of the hook format.
2. Contract tests to ensure your adapter round-trips correctly.

### Step 3: Contract Test

Add test cases to `internal/adapters/schema_test.go`:

```go
func TestMyToolAdapterPreCall(t *testing.T) {
	adapter := NewMyToolAdapter()

	fixture, err := os.ReadFile(filepath.Join("fixtures", "mytool_precall.json"))
	if err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}

	event, err := adapter.DecodeHookEvent(fixture)
	if err != nil {
		t.Fatalf("DecodeHookEvent failed: %v", err)
	}

	// Validate key fields match your fixture.
	if event.HostType != "mytool" {
		t.Errorf("HostType: expected mytool, got %s", event.HostType)
	}
	if event.Tool != "read" {
		t.Errorf("Tool: expected read, got %s", event.Tool)
	}

	// Encode and validate round-trip.
	result := &HookResult{
		PermissionDecision: PermissionAllow,
	}
	encoded, err := adapter.EncodeHookResult(result)
	if err != nil {
		t.Fatalf("EncodeHookResult failed: %v", err)
	}

	var output myToolHookResult
	if err := json.Unmarshal(encoded, &output); err != nil {
		t.Fatalf("encoded result is not valid JSON: %v", err)
	}
}
```

### Step 4: Update the Registry

Edit `internal/adapters/registry.go` to add your host:

```go
func GetAdapter(hostType string) (HostAdapter, error) {
	switch hostType {
	case "mytool":
		return NewMyToolAdapter(), nil
	// ... existing cases ...
	}
}
```

And add a constant to `internal/adapters/schema.go`:

```go
const (
	// ... existing hosts ...
	HostMyTool = "mytool"
)
```

### Step 5: Pull Request Checklist

Before submitting, verify:
- [ ] Adapter file is complete and compiles.
- [ ] Fixtures load and contain real hook JSON (not mock data).
- [ ] Contract tests pass: `go test ./internal/adapters/...`
- [ ] The adapter round-trips: decode(fixture) → encode() → decode() are consistent.
- [ ] CONTRIBUTING.md is referenced in your PR description.
- [ ] You've tested with a real hook from your host if possible (simulator or live session).

### Step 6: Documentation

Add a brief note to the host list in README.md:
```
| **MyTool** | Hook field names | Post-call rewrite | MCP | Installation |
|---|---|---|---|---|
| MyTool | `hookType`, `toolName` | yes | yes | plugin bundle, marketplace |
```

## Testing Your Adapter

**Minimal test:**
```bash
cd internal/adapters
go test -v -run TestMyToolAdapter
```

**With fixtures:**
Ensure your fixtures directory contains both `mytool_precall.json` and `mytool_postcall.json`. The tests will load them and validate round-trip fidelity.

**Integration test (post-v1):**
Once the daemon + hook client are live, you can route real hooks from your host to the adapter and measure compression end-to-end. For now, the contract tests are sufficient.

## Questions?

Open an issue or discussion on GitHub. The schema is the contract; if your hook format doesn't fit, we may expand the schema (which affects all adapters equally, so it's worth discussing first).

Thank you for contributing!
