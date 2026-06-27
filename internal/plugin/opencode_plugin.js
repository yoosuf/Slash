// Slash compression plugin for opencode
// Hooks into tool execution to compress tool outputs via the Slash daemon.

const path = require('path');
const net = require('net');
const os = require('os');
const crypto = require('crypto');

const SOCKET_PATH = process.env.SLASH_SOCKET
  || process.env.SLASH_DAEMON_SOCKET
  || path.join(os.homedir(), '.slash', 'daemon.sock');
const TIMEOUT_MS = 3000;

function generateId() {
  return crypto.randomUUID();
}

function getSessionId() {
  if (!process.env.SLASH_SESSION_ID) {
    process.env.SLASH_SESSION_ID = generateId();
  }
  return process.env.SLASH_SESSION_ID;
}

async function callDaemon(event) {
  return new Promise((resolve, reject) => {
    const socket = net.createConnection(SOCKET_PATH, () => {
      socket.write(JSON.stringify(event) + '\n');
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

module.exports = async function () {
  return {
    'tool.execute.after': async (input, output) => {
      try {
        const toolName = input.tool || input.name || 'unknown';
        const toolOutput = output.result != null ? output.result
          : output.content != null ? output.content
          : '';

        const event = {
          host_type: 'opencode',
          eventId: generateId(),
          eventKind: 'postToolUse',
          sessionId: getSessionId(),
          tool: toolName,
          toolInput: input.args || {},
          toolOutput: toolOutput,
          workspaceDir: process.cwd(),
          machineId: os.hostname(),
        };

        const result = await callDaemon(event);

        if (result && result.hookSpecificOutput && result.hookSpecificOutput.updatedToolOutput != null) {
          const compressed = result.hookSpecificOutput.updatedToolOutput;
          if (output.result != null) {
            output.result = compressed;
          }
          if (output.content != null) {
            output.content = compressed;
          }
        }
      } catch (error) {
        // Fail open
      }
    },
  };
};
