import * as vscode from "vscode";
import { defaultSocketPath, requestStats, DaemonStats } from "./daemon";
import { spawn, ChildProcess } from "child_process";
import * as path from "path";
import * as fs from "fs";

let statusBarItem: vscode.StatusBarItem | undefined;
let pollingTimer: NodeJS.Timeout | undefined;

export function activate(context: vscode.ExtensionContext) {
  statusBarItem = vscode.window.createStatusBarItem(
    vscode.StatusBarAlignment.Right,
    100,
  );
  statusBarItem.command = "slash.stats";
  statusBarItem.tooltip = "Slash Compressor";
  statusBarItem.show();

  context.subscriptions.push(
    vscode.commands.registerCommand("slash.toggle", toggleCompression),
    vscode.commands.registerCommand("slash.stats", showStats),
    vscode.commands.registerCommand("slash.restartDaemon", restartDaemon),
    statusBarItem,
  );

  pollDaemon();
}

function deactivate() {
  if (pollingTimer) clearTimeout(pollingTimer);
  if (statusBarItem) statusBarItem.dispose();
}

function getSocketPath(): string {
  const config = vscode.workspace.getConfiguration("slash");
  return config.get<string>("socketPath", "") || defaultSocketPath();
}

function findSlashBinary(): string | undefined {
  const candidates = [
    path.join(require("os").homedir(), ".local", "bin", "slash"),
    "/opt/homebrew/bin/slash",
    "/usr/local/bin/slash",
  ];
  for (const p of candidates) {
    try {
      fs.accessSync(p, fs.constants.X_OK);
      return p;
    } catch {}
  }
  return undefined;
}

function updateStatusBar(
  state: "connected" | "starting" | "stopped",
  stats?: DaemonStats,
) {
  if (!statusBarItem) return;

  if (state === "connected" && stats) {
    const total = stats.total_calls ?? 0;
    const compressed = stats.calls_compressed ?? 0;
    const pct = total > 0 ? Math.round((compressed / total) * 100) : 0;
    statusBarItem.text = `$(rocket) Slash ${pct}%`;
    statusBarItem.backgroundColor = undefined;
    statusBarItem.tooltip = "Slash Compressor — click for stats";
  } else if (state === "starting") {
    statusBarItem.text = "$(loading~spin) Slash";
    statusBarItem.backgroundColor = undefined;
    statusBarItem.tooltip = "Slash daemon starting...";
  } else {
    statusBarItem.text = "$(circle-slash) Slash";
    statusBarItem.backgroundColor = new vscode.ThemeColor(
      "statusBarItem.warningBackground",
    );
    statusBarItem.tooltip = "Slash daemon not running";
  }
}

async function pollDaemon() {
  const socketPath = getSocketPath();
  try {
    const stats = await requestStats(socketPath, 2000);
    updateStatusBar("connected", stats);
  } catch {
    updateStatusBar("stopped");
  }
  pollingTimer = setTimeout(pollDaemon, 5000);
}

async function waitForDaemon(
  socketPath: string,
  maxAttempts = 20,
  intervalMs = 500,
): Promise<DaemonStats> {
  for (let i = 0; i < maxAttempts; i++) {
    try {
      return await requestStats(socketPath, 2000);
    } catch {
      if (i < maxAttempts - 1) {
        await new Promise((r) => setTimeout(r, intervalMs));
      }
    }
  }
  throw new Error("Daemon did not start within timeout");
}

async function showStats() {
  const socketPath = getSocketPath();

  let stats: DaemonStats;
  try {
    stats = await requestStats(socketPath, 2000);
  } catch {
    const action = await vscode.window.showErrorMessage(
      "Slash daemon is not running.",
      "Start Daemon",
      "Install Slash",
    );
    if (action === "Start Daemon") {
      startDaemon();
    } else if (action === "Install Slash") {
      vscode.env.openExternal(
        vscode.Uri.parse(
          "https://github.com/yoosuf/Slash#quick-install",
        ),
      );
    }
    return;
  }

  const panel = vscode.window.createWebviewPanel(
    "slashStats",
    "Slash Compression Stats",
    vscode.ViewColumn.Beside,
    { enableScripts: false },
  );

  const breakdown = stats.method_breakdown
    ? Object.entries(stats.method_breakdown)
        .map(([k, v]) => `<tr><td>${k}</td><td>${v}</td></tr>`)
        .join("")
    : "";

  panel.webview.html = `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><title>Slash Stats</title>
<style>
  body { font-family: -apple-system, sans-serif; padding: 16px; }
  h2 { margin: 0 0 16px; }
  table { border-collapse: collapse; width: 100%; }
  td, th { padding: 6px 12px; text-align: left; border-bottom: 1px solid var(--vscode-editor-lineHighlightBorder); }
  th { font-weight: 600; }
</style>
</head>
<body>
  <h2>Slash Compression Stats</h2>
  <table>
    <tr><th>Metric</th><th>Value</th></tr>
    <tr><td>Total calls</td><td>${stats.total_calls ?? "—"}</td></tr>
    <tr><td>Calls compressed</td><td>${stats.calls_compressed ?? "—"}</td></tr>
    <tr><td>Active sessions</td><td>${stats.active_sessions ?? "—"}</td></tr>
    <tr><td>Latency p50</td><td>${stats.latency_p50_ms ?? "—"} ms</td></tr>
    <tr><td>Latency p95</td><td>${stats.latency_p95_ms ?? "—"} ms</td></tr>
    <tr><td>Cache size</td><td>${stats.cache_size_mb ?? "—"} MB</td></tr>
    <tr><td>Cache entries</td><td>${stats.cache_entries ?? "—"}</td></tr>
    ${breakdown ? `<tr><th colspan="2">Method Breakdown</th></tr>${breakdown}` : ""}
  </table>
</body>
</html>`;
}

function toggleCompression() {
  const config = vscode.workspace.getConfiguration("slash");
  const current = config.get<boolean>("enabled", true);
  config.update("enabled", !current, true);
  vscode.window.showInformationMessage(
    `Slash compression ${current ? "disabled" : "enabled"}`,
  );
  if (statusBarItem) {
    statusBarItem.text = current
      ? "$(circle-slash) Slash (off)"
      : "$(rocket) Slash";
  }
}

async function startDaemon() {
  const slashPath = findSlashBinary();
  if (!slashPath) {
    vscode.window.showErrorMessage(
      "Slash binary not found. Install: brew install yoosuf/tap/slash",
    );
    return;
  }

  updateStatusBar("starting");
  vscode.window.withProgress(
    {
      location: vscode.ProgressLocation.Notification,
      title: "Starting Slash daemon...",
      cancellable: true,
    },
    async (progress, token) => {
      token.onCancellationRequested(() => {});

      const stderrChunks: Buffer[] = [];
      let spawnError: string | undefined;
      let exitCode: number | undefined;

      const proc = spawn(slashPath, ["daemon"], {
        stdio: ["ignore", "ignore", "pipe"],
        detached: true,
      });
      proc.stderr?.on("data", (d: Buffer) => stderrChunks.push(d));
      proc.on("error", (err) => {
        spawnError = err.message;
      });
      proc.on("exit", (code) => {
        exitCode = code ?? undefined;
      });
      proc.unref();

      try {
        const socketPath = getSocketPath();
        await waitForDaemon(socketPath);
        vscode.window.showInformationMessage("Slash daemon ready");
        pollDaemon();
      } catch {
        const parts: string[] = [`Binary: ${slashPath}`];
        if (spawnError) parts.push(`Spawn error: ${spawnError}`);
        if (exitCode !== undefined) parts.push(`Exit code: ${exitCode}`);
        const stderr = Buffer.concat(stderrChunks)
          .toString()
          .trim();
        if (stderr) parts.push(`Stderr: ${stderr.split("\n").pop()}`);
        vscode.window.showErrorMessage(
          `Failed to start Slash daemon. ${parts.join(" | ")}`,
        );
        updateStatusBar("stopped");
      }
    },
  );
}

async function restartDaemon() {
  startDaemon();
}
