# Slash Daemon & MCP API Documentation

This document describes the API and communication protocols for the **Slash** daemon. 

Slash operates using a **client-daemon-MCP architecture**:
1. **CLI Hook Clients** intercept tool executions in your editor/agent host.
2. The clients normalize the events and send them to the background **Daemon** via a Unix domain socket.
3. The daemon compresses the tool input/output, saves the original content in the local **CCR (Content-Addressable Retrieval) Cache**, and returns the compressed payload.
4. If the agent needs the compressed details, it uses the **MCP Server** (running on localhost) to retrieve the original content.

---

## 1. IPC Daemon API (Unix Domain Socket)

By default, the daemon listens on a Unix domain socket located at:
- **Linux/macOS:** `$HOME/.slash/daemon.sock`
- **Windows:** Named pipe (configured via config or environment variables)

Every hook event sent to the socket is a single JSON payload. The socket connection is transactional: the client opens the connection, sends one JSON event, receives one JSON result, and the connection is closed.

### Event Schema (`HookEvent`)

Clients send a JSON object representing a normalized `HookEvent`:

```json
{
  "host_type": "claudecode",
  "event_id": "evt_12345678",
  "session_id": "sess_87654321",
  "event_kind": "preToolUse",
  "tool": "read_file",
  "tool_input": {
    "path": "/absolute/path/to/file.go"
  },
  "tool_output": null,
  "workspace": "/absolute/path/to/workspace",
  "machine_id": "mac_abc123",
  "timestamp": "2026-06-27T08:30:00Z",
  "host_specific": {}
}
```

| Field Name | Type | Description |
| :--- | :--- | :--- |
| `host_type` | `string` | The ID of the editor/CLI host emitting this event. One of: `"claudecode"`, `"cursor"`, `"windsurf"`, `"codex"`, `"antigravity"`, `"copilot"`, `"aider"`, `"zed"`. |
| `event_id` | `string` | Unique identifier for this hook invocation. |
| `session_id` | `string` | Groups all hook calls within a single workspace/session. |
| `event_kind` | `string` | The phase of the tool execution: `"preToolUse"`, `"postToolUse"`, `"sessionStart"`, `"sessionEnd"`. |
| `tool` | `string` | The name of the tool being executed (e.g. `"read"`, `"bash"`, `"apply_patch"`). |
| `tool_input` | `object` | The input arguments passed to the tool prior to execution. |
| `tool_output` | `any` | The raw result returned by the tool (present only in `"postToolUse"` phase). |
| `workspace` | `string` | Absolute path to the workspace root directory. |
| `machine_id` | `string` | Stable machine-specific identifier for cache scoping. |
| `timestamp` | `string` | ISO 8601 UTC timestamp of the hook event creation. |
| `host_specific`| `object` | Optional host-specific properties parsed by adapters. |

### Result Schema (`HookResult`)

The daemon processes the `HookEvent` and returns a `HookResult` JSON object:

```json
{
  "permission_decision": "allow",
  "updated_input": {
    "path": "/absolute/path/to/file.go",
    "range": [1, 50]
  },
  "updated_tool_output": "[slash: Code AST outline. retrieve(h_abc123) for full content]",
  "compression_meta": "[slash: 1.2k -> 0.2k tokens]",
  "host_specific": {}
}
```

| Field Name | Type | Description |
| :--- | :--- | :--- |
| `permission_decision` | `string` | `"allow"` or `"deny"`. If `"deny"`, the host will abort tool execution. |
| `updated_input` | `object` | Replaces the tool's input arguments in `"preToolUse"` (e.g., specifying a line range for read optimizations). |
| `updated_tool_output` | `any` | Replaces the tool output in `"postToolUse"` with compressed representations + cache retrieval handle notes. |
| `compression_meta` | `string` | Human-readable compression details logged by the host and sent as side-channel hints to the LLM. |
| `host_specific` | `object` | Optional fields added back by adapters for native host encoding. |

---

## 2. Model Context Protocol (MCP) API

The daemon runs an integrated HTTP server to expose Slash features as standard MCP tools. By default, it runs on port `8080` (or another port specified in `config.json`).

All tools are called via HTTP `POST` requests to JSON endpoints.

### Health Check

- **Endpoint:** `/health`
- **Method:** `GET`
- **Response:**
  ```json
  {
    "status": "healthy",
    "service": "slash-mcp"
  }
  ```

---

### Tool 1: `retrieve` (Retrieve Cached Content)

Fetches the original, uncompressed tool input or output using its content-addressable handle.

- **Endpoint:** `/mcp/retrieve`
- **Method:** `POST`
- **Request Body:**
  ```json
  {
    "handle": "h_abc123xyz",
    "start": 0,
    "end": 1000
  }
  ```
  - `handle` (required): The unique content handle (e.g. `h_abc123xyz`).
  - `start` (optional): Byte offset to begin retrieving.
  - `end` (optional): Byte offset to stop retrieving.

- **Response (200 OK):**
  ```json
  {
    "handle": "h_abc123xyz",
    "content": "package main\n\nimport \"fmt\"...",
    "size": 2561
  }
  ```

---

### Tool 2: `repomap` (Repository Structure Mapping)

Builds a symbol map of the repository, highlighting key functions, interfaces, imports, and variables.

- **Endpoint:** `/mcp/repomap`
- **Method:** `POST`
- **Request Body:**
  ```json
  {
    "path": "/absolute/path/to/subfolder"
  }
  ```
  - `path` (optional): Limits the symbol map to files inside a specific directory.

- **Response (200 OK):**
  ```json
  {
    "path": "/absolute/path/to/subfolder",
    "modules": [
      "main.go: main()",
      "server.go: MCPServer struct, Start(), handleRetrieve()"
    ],
    "note": "Use retrieve(handle) on individual file skeletons to read implementation bodies."
  }
  ```

---

### Tool 3: `stats` (Compression Statistics)

Queries the current session's token savings, active cache size, and elapsed uptime.

- **Endpoint:** `/mcp/stats`
- **Method:** `POST`
- **Request Body:** `{}` (empty object)

- **Response (200 OK):**
  ```json
  {
    "cache_size_mb": 14.5,
    "cache_entries": 128,
    "timestamp": "2026-06-27T08:31:02Z"
  }
  ```
