#!/usr/bin/env python3

import asyncio
import json
import sys
import websockets
import argparse
import time
from datetime import datetime
import logging

logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s"
)


class Colors:
    GREEN = "\033[32m"
    RED = "\033[31m"
    YELLOW = "\033[33m"
    BLUE = "\033[34m"
    RESET = "\033[0m"


class WebSocketTest:
    def __init__(self, token):
        self.token = token
        self.uri = "ws://localhost/ws"
        self.headers = {
            "Authorization": f"Bearer {token}",
        }
        self.test_results = []

    async def test_connection(self):
        try:
            websocket = await websockets.connect(
                self.uri,
                extra_headers=self.headers,
                ping_interval=None,
                compression=None,
            )
            logging.info("Connection established")

            message = {
                "type": "chat",
                "content": json.dumps({"text": "Connection test"}),
                "receiver_id": 1,
            }
            logging.info(f"Sending test message: {message}")
            await websocket.send(json.dumps(message))

            response = await websocket.recv()
            logging.info(f"Received response: {response}")

            data = json.loads(response)
            assert data["type"] == "ack", "Expected acknowledgment"

            await websocket.close()
            return True
        except Exception as e:
            logging.error(f"Connection test failed: {e}")
            return False

    async def test_chat_message(self):
        try:
            websocket = await websockets.connect(
                self.uri,
                extra_headers=self.headers,
                ping_interval=None,
                compression=None,
            )

            message = {
                "type": "chat",
                "content": json.dumps(
                    {"text": "Hello from test suite!", "timestamp": int(time.time())}
                ),
                "receiver_id": 1,
            }

            logging.info(f"Sending chat message: {message}")
            await websocket.send(json.dumps(message))

            response = await websocket.recv()
            logging.info(f"Received response: {response}")

            data = json.loads(response)
            assert data["type"] == "ack", "Expected acknowledgment"
            assert "message_id" in data, "Expected message ID in response"

            await websocket.close()
            return True
        except Exception as e:
            logging.error(f"Chat test failed: {e}")
            return False

    async def run_test(self, name, test_func):
        print(f"\n{Colors.BLUE}Running test: {name}{Colors.RESET}")
        try:
            result = await test_func()
            if result:
                print(f"{Colors.GREEN}✓ Test passed: {name}{Colors.RESET}")
                self.test_results.append((name, True))
            else:
                print(f"{Colors.RED}✗ Test failed: {name}{Colors.RESET}")
                self.test_results.append((name, False))
        except Exception as e:
            print(f"{Colors.RED}✗ Test failed: {name}")
            print(f"Error: {str(e)}{Colors.RESET}")
            self.test_results.append((name, False))

    async def run_all_tests(self):
        print(
            f"{Colors.YELLOW}Starting WebSocket test suite at {datetime.now()}{Colors.RESET}"
        )

        tests = [
            ("Connection Test", self.test_connection),
            ("Chat Message Test", self.test_chat_message),
        ]

        for name, test_func in tests:
            await self.run_test(name, test_func)

        print(f"\n{Colors.YELLOW}Test Summary:{Colors.RESET}")
        passed = sum(1 for _, result in self.test_results if result)
        total = len(self.test_results)

        print(f"Passed: {Colors.GREEN}{passed}/{total}{Colors.RESET}")
        if passed == total:
            print(f"{Colors.GREEN}All tests passed!{Colors.RESET}")
        else:
            print(f"{Colors.RED}Some tests failed.{Colors.RESET}")
            failed_tests = [name for name, result in self.test_results if not result]
            print("Failed tests:")
            for test in failed_tests:
                print(f"{Colors.RED}- {test}{Colors.RESET}")


def main():
    parser = argparse.ArgumentParser(description="WebSocket Test Suite")
    parser.add_argument("token", help="JWT access token")
    parser.add_argument("--debug", action="store_true", help="Enable debug logging")
    args = parser.parse_args()

    if args.debug:
        logging.getLogger().setLevel(logging.DEBUG)

    try:
        test_suite = WebSocketTest(args.token)
        asyncio.get_event_loop().run_until_complete(test_suite.run_all_tests())
    except Exception as e:
        print(f"{Colors.RED}Test suite failed: {str(e)}{Colors.RESET}")
        sys.exit(1)


if __name__ == "__main__":
    main()
