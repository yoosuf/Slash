package adapters

import (
	"encoding/json"
	"fmt"
)

// GetAdapter returns the HostAdapter for the given host type.
// This is the factory function that instantiates the right adapter.
func GetAdapter(hostType string) (HostAdapter, error) {
	switch hostType {
	case HostClaudeCode:
		return NewClaudeCodeAdapter(), nil
	case HostCodex:
		return NewCodexAdapter(), nil
	case HostAntigravity:
		return NewAntigravityAdapter(), nil
	case HostCopilot:
		return NewCopilotAdapter(), nil
	case HostCursor:
		return NewCursorAdapter(), nil
	case HostWindsurf:
		return NewWindsurfAdapter(), nil
	case HostAider:
		return NewAiderAdapter(), nil
	case HostOpenCode:
		return NewOpenCodeAdapter(), nil
	default:
		return nil, fmt.Errorf("unknown host type: %s", hostType)
	}
}

// DetectHostType attempts to infer the host type from a raw hook JSON event
// by checking for host-specific field patterns. This is a best-effort heuristic.
// If detection fails, callers should fall back to explicit configuration.
func DetectHostType(raw []byte) (string, error) {
	// Parse as generic JSON to inspect structure.
	var generic map[string]interface{}
	if err := json.Unmarshal(raw, &generic); err != nil {
		return "", fmt.Errorf("failed to parse JSON for host detection: %w", err)
	}

	// Claude Code uses camelCase (eventId, workspaceDir).
	if _, hasEventID := generic["eventId"]; hasEventID {
		if _, hasWorkspaceDir := generic["workspaceDir"]; hasWorkspaceDir {
			return HostClaudeCode, nil
		}
	}

	// Codex, Antigravity, Copilot use snake_case (event_id, workspace_dir).
	if _, hasEventID := generic["event_id"]; hasEventID {
		if _, hasWorkspaceDir := generic["workspace_dir"]; hasWorkspaceDir {
			// Could be any of the three. Look for host-specific hints.
			// For now, default to Codex (the next most common).
			// In practice, the host should be configured explicitly.
			return HostCodex, nil
		}
	}

	return "", fmt.Errorf("unable to detect host type from hook JSON")
}
