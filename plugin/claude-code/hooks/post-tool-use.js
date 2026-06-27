// Slash post-tool-use hook for Claude Code
// This hook is executed after a tool call and can modify/compress the output.

const fs = require('fs');
const path = require('path');
const net = require('net');

const SOCKET_PATH = process.env.SLASH_SOCKET || process.env.SLASH_DAEMON_SOCKET ||
  path.join(process.env.HOME || '/tmp', '.slash', 'daemon.sock');
const TIMEOUT_MS = 2000;

module.exports = async function handlePostToolUse(event) {
  try {
    const result = await callDaemon(event);
    return result;
  } catch (error) {
    console.error('[slash] post-tool-use hook error:', error.message);
    // Fail open
    return {
      permissionDecision: 'allow',
      hookSpecificOutput: undefined
    };
  }
};

async function callDaemon(event) {
  return new Promise((resolve, reject) => {
    const socket = net.createConnection(SOCKET_PATH, () => {
      const payload = {
        ...event,
        host_type: 'claudecode'
      };
      socket.write(JSON.stringify(payload) + '\n');
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
