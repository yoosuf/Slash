// Slash pre-tool-use hook for Claude Code
// This hook is executed before a tool call and can modify the input.

const fs = require('fs');
const path = require('path');
const net = require('net');

// Configuration from Claude Code environment.
const SOCKET_PATH = process.env.SLASH_SOCKET || process.env.SLASH_DAEMON_SOCKET ||
  path.join(process.env.HOME || '/tmp', '.slash', 'daemon.sock');
const TIMEOUT_MS = 2000; // 2 second timeout; fail open if exceeded

module.exports = async function handlePreToolUse(event) {
  try {
    const result = await callDaemon(event);
    return result;
  } catch (error) {
    console.error('[slash] pre-tool-use hook error:', error.message);
    // Fail open: return original event
    return {
      permissionDecision: 'allow',
      hookSpecificOutput: undefined
    };
  }
};

async function callDaemon(event) {
  return new Promise((resolve, reject) => {
    const socket = net.createConnection(SOCKET_PATH, () => {
      // Connected; send event as JSON.
      socket.write(JSON.stringify(event) + '\n');
      socket.write(JSON.stringify({host_type: 'claudecode'}) + '\n');
    });

    socket.setTimeout(TIMEOUT_MS, () => {
      socket.destroy();
      reject(new Error('daemon timeout'));
    });

    socket.on('data', (data) => {
      try {
        const result = JSON.parse(data.toString());
        socket.end();
        resolve(result);
      } catch (e) {
        reject(new Error('invalid daemon response'));
      }
    });

    socket.on('error', (error) => {
      reject(error);
    });
  });
}
