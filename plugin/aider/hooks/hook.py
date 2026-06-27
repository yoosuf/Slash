#!/usr/bin/env python3
"""
Slash hook for Aider CLI.
Sends hook events to the daemon for compression.
"""

import json
import socket
import os
from pathlib import Path


def get_socket_path():
    """Get the Slash daemon socket path."""
    return os.environ.get(
        'SLASH_SOCKET',
        str(Path.home() / '.slash' / 'daemon.sock')
    )


def call_daemon(event: dict) -> dict:
    """Send an event to the daemon and receive the result."""
    socket_path = get_socket_path()
    timeout_ms = 2000

    try:
        sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        sock.settimeout(timeout_ms / 1000.0)

        # Connect to daemon
        sock.connect(socket_path)

        # Send event as JSON
        event_json = json.dumps(event) + '\n'
        sock.sendall(event_json.encode())

        # Receive result
        data = sock.recv(65536)
        result = json.loads(data.decode())

        sock.close()
        return result

    except Exception as e:
        # Fail open on any error
        print(f"[slash] hook error: {e}", file=__import__('sys').stderr)
        return {
            'permission_decision': 'allow',
            'updated_input': None,
            'updated_tool_output': None
        }


def handle_hook(event: dict) -> dict:
    """Main hook handler."""
    # Ensure host type is set
    if 'host_type' not in event:
        event['host_type'] = 'aider'

    result = call_daemon(event)
    return result


# If invoked directly (for testing)
if __name__ == '__main__':
    import sys

    try:
        event = json.loads(sys.stdin.read())
        result = handle_hook(event)
        print(json.dumps(result))
    except Exception as e:
        print(json.dumps({'error': str(e)}), file=sys.stderr)
        sys.exit(1)
