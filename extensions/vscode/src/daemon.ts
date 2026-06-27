import * as net from "net";
import * as os from "os";
import * as path from "path";

export interface DaemonStats {
  total_calls?: number;
  calls_compressed?: number;
  active_sessions?: number;
  latency_p50_ms?: number;
  latency_p95_ms?: number;
  method_breakdown?: Record<string, number>;
  cache_size_mb?: number;
  cache_entries?: number;
}

export function defaultSocketPath(): string {
  const home = os.homedir();
  return path.join(home, ".slash", "daemon.sock");
}

export async function requestStats(
  socketPath: string,
  timeoutMs = 5000,
): Promise<DaemonStats> {
  const body = JSON.stringify({ type: "stats" });
  const data = await send(socketPath, body, timeoutMs);
  if (!data || data.length === 0) {
    throw new Error("empty response from daemon");
  }
  return JSON.parse(data) as DaemonStats;
}

function send(
  socketPath: string,
  body: string,
  timeoutMs: number,
): Promise<string> {
  return new Promise((resolve, reject) => {
    const socket = new net.Socket();
    let settled = false;
    let chunks: Buffer[] = [];

    const timer = setTimeout(() => {
      if (settled) return;
      settled = true;
      socket.destroy();
      reject(new Error(`Socket timeout after ${timeoutMs}ms`));
    }, timeoutMs);

    function done(err?: Error) {
      if (settled) return;
      settled = true;
      clearTimeout(timer);
      if (err) {
        reject(err);
      } else {
        resolve(Buffer.concat(chunks).toString());
      }
    }

    socket.on("error", (err) => {
      done(err);
    });

    socket.on("close", (hadError: boolean) => {
      if (hadError) {
        done(new Error("Socket closed with error"));
      } else if (chunks.length === 0) {
        done(new Error("Daemon closed connection without sending data"));
      } else {
        done();
      }
    });

    socket.on("data", (data: Buffer) => {
      chunks.push(data);
    });

    socket.connect(socketPath, () => {
      socket.write(body);
    });
  });
}
