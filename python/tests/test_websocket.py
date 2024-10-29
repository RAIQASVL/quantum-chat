#!/usr/bin/env python3

import asyncio
import websockets
import json
import logging
import sys

logging.basicConfig(level=logging.DEBUG)


async def test_connection(token):
    uri = "ws://localhost/ws"
    headers = {"Authorization": f"Bearer {token}"}

    try:
        logging.info("Connecting to WebSocket server...")
        websocket = await websockets.connect(
            uri, extra_headers=headers, ping_interval=None, compression=None
        )

        logging.info("Connected successfully!")

        message = {
            "type": "chat",
            "content": json.dumps({"text": "Hello!"}),
            "receiver_id": 1,
        }

        logging.info(f"Sending message: {message}")
        await websocket.send(json.dumps(message))

        response = await websocket.recv()
        logging.info(f"Received response: {response}")

        await websocket.close()
        return True

    except Exception as e:
        logging.error(f"Error: {str(e)}")
        return False


if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python test_websocket.py <token>")
        sys.exit(1)

    token = sys.argv[1]
    result = asyncio.get_event_loop().run_until_complete(test_connection(token))

    if result:
        print("Test completed successfully!")
        sys.exit(0)
    else:
        print("Test failed!")
        sys.exit(1)
