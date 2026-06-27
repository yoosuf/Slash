# Slash Frequently Asked Questions (FAQ)

Here are answers to the most common questions about **Slash**.

---

### Q: What is Slash?
**A:** Slash is a local-first, reversible token compressor for coding agents and IDEs. It runs as a lightweight background daemon that intercepts reads, file outputs, and terminal logs to compress them by 40–60% before they enter your LLM context, saving tokens, money, and time.

---

### Q: Does my code leave my machine?
**A:** **No.** By default, Slash operates entirely offline. Your files are parsed, compressed, and stored in a local SQLite cache under your standard user-level permissions. No code or metadata is sent to any external server. 

An optional, opt-in cloud module (`slash-cloud`) will be available in V2 for teams wishing to share cache buckets, but this will require explicit installation and configuration.

---

### Q: Does compression degrade coding agent performance or accuracy?
**A:** **No, because the compression is fully reversible.** When Slash compresses a file (e.g. collapsing function implementation details into a skeleton outline), it appends a retrieve handle: `__slash_note: [retrieve(h_abc123) for full]`. 

If the agent needs the hidden implementation details, it automatically invokes the `retrieve` tool via the MCP server and receives the original content. This "fail-safe" design ensures accuracy is maintained.

---

### Q: How much token savings should I expect?
**A:** Typical session-wide token savings are **40% to 60%**. The exact savings depend on the tasks:
- **Code reads**: 60–70% savings (by emitting AST skeletons).
- **JSON payloads**: 40–50% savings (by stripping leaf values and retaining structure).
- **Terminal output/logs**: 50–80% savings (by deduplicating repeating lines).
- **Re-reads of modified files**: 80–95% savings (by utilizing diff-only outputs).

---

### Q: What editors/hosts are supported?
**A:** Out of the box, Slash includes integration adapters for:
- **Claude Code**
- **Cursor**
- **Windsurf**
- **Codex**
- **Antigravity (agy)**
- **Copilot CLI**
- **Aider**
- **Zed** (via MCP extension)

---

### Q: How do I disable or uninstall Slash?
**A:** 
- **Temporary Disable**: Set `"enabled": false` under `"compression"` in your `~/.slash/config.json` file.
- **Stop Daemon**: Run `pkill -f "slash daemon"`.
- **Remove Hooks**: Run `slash plugin uninstall <host>` (or delete the hook configuration in your editor settings).
