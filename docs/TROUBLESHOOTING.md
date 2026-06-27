# Slash Troubleshooting Guide

This guide describes how to diagnose and resolve common issues encountered while running the **Slash** daemon or its editor hook integrations.

---

## 1. IPC Connection Issues

### Symptom: `dial unix ~/.slash/daemon.sock: connect: no such file or directory` or `connection refused`

This means the hook adapter client cannot establish a connection with the Slash daemon process.

#### Solutions:
1. **Check if the daemon is running**:
   ```bash
   ps aux | grep "slash daemon"
   ```
2. **Start the daemon manually**:
   If it is not running, launch it in the background:
   ```bash
   slash daemon &
   ```
3. **Verify the socket file exists**:
   ```bash
   ls -la ~/.slash/daemon.sock
   ```
   If the socket file exists but you still receive connection errors, the socket may be orphaned. Terminate any stale processes and remove the socket file:
   ```bash
   pkill -f "slash daemon"
   rm -f ~/.slash/daemon.sock
   slash daemon &
   ```

---

## 2. Port Binding Conflicts

### Symptom: `listen tcp :8080: bind: address already in use`

This happens when the built-in MCP HTTP server tries to bind to port 8080, but another service on your machine is already using it.

#### Solutions:
1. Identify the process using port 8080:
   - **macOS**: `lsof -i :8080`
   - **Linux**: `ss -lptn 'sport = :8080'`
2. Change the MCP port inside your `~/.slash/config.json`:
   ```json
   {
     "daemon": {
       "port": 9090
     }
   }
   ```
3. Restart the daemon to apply changes.

---

## 3. Caching and Stale Data Issues

### Symptom: Stale files returned or tool executions retrieve incorrect contents

This occurs if the CCR (Content-Addressable Retrieval) cache database gets out of sync, or file updates fail to invalidate cache entries correctly.

#### Solutions:
1. **Force purge the cache**:
   ```bash
   slash purge
   ```
2. **Verify permissions on the cache directory**:
   Ensure the current user has read/write permissions to the caching location:
   - macOS/Linux: `~/.cache/slash` (or `$XDG_CACHE_HOME/slash`)
   - Windows: `%LOCALAPPDATA%\slash`
   Run:
   ```bash
   chmod -R 700 ~/.cache/slash
   ```

---

## 4. High Resource Consumption

### Symptom: Daemon CPU or memory spikes during large repository indexes

Slash parses code files to create AST skeletons. Scanning a directory containing millions of dependencies (e.g., `node_modules` or `.venv`) can cause spikes.

#### Solutions:
1. **Exclude folders using `.gitignore`**:
   Slash respects your workspace `.gitignore` file. Ensure large dependency folders are properly ignored.
2. **Create a `.slashignore` file**:
   If you want to exclude additional paths from compression or scanning without modifying git behavior, add a `.slashignore` file in the root of your workspace.
3. **Disable repository mapping**:
   If the repomap builder is consuming too many resources, turn it off in `~/.slash/config.json`:
   ```json
   {
     "compression": {
       "repo_map_inject": false
     }
   }
   ```
