#!/bin/bash

TOR_BINARY="/usr/bin/tor"
BRIDGE_TORRC="torrc-bridge"
CLIENT_TORRC="torrc-client"
BRIDGE_DATA_DIR="./tor-bridge-data"
FINGERPRINT_FILE="${BRIDGE_DATA_DIR}/fingerprint"

BRIDGE_PT_PORT="24433"
BRIDGE_PT_IP="127.0.0.1"

cleanup() {
  echo -e "\n[*] Shutting down Tor processes..."
  if [[ -n "$CLIENT_PID" ]]; then
    kill "$CLIENT_PID" 2>/dev/null
  fi
  if [[ -n "$BRIDGE_PID" ]]; then
    kill "$BRIDGE_PID" 2>/dev/null
  fi

  echo "[*] Removing temporary data directories..."
  rm -rf ./tor-client-data
  rm -rf ./tor-bridge-data
  rm tor-bridge-log.txt
  rm tor-client-log.txt
  rm torrc-client.bak

  echo "[*] Cleanup complete."
  exit 0
}

trap cleanup SIGINT SIGTERM EXIT

echo "[*] Starting Tor bridge in the background..."
$TOR_BINARY -f "$BRIDGE_TORRC" &
BRIDGE_PID=$!
sleep 2

echo "[*] Waiting for the bridge to generate its fingerprint..."
while [ ! -f "$FINGERPRINT_FILE" ]; do
  if ! kill -0 "$BRIDGE_PID" 2>/dev/null; then
    echo "[!] ERROR: Bridge process died unexpectedly. Check logs."
    exit 1
  fi
  sleep 1
done
echo "[+] Fingerprint file found!"

FINGERPRINT=$(awk '{print $2}' "$FINGERPRINT_FILE")
if [ -z "$FINGERPRINT" ]; then
  echo "[!] ERROR: Could not extract fingerprint."
  exit 1
fi
echo "[+] Extracted Fingerprint: $FINGERPRINT"

sleep 10
echo "[*] Updating client configuration file: ${CLIENT_TORRC}"
sed -i.bak "s/^Bridge netshaper .*/Bridge netshaper ${BRIDGE_PT_IP}:${BRIDGE_PT_PORT} ${FINGERPRINT}/" "$CLIENT_TORRC"

echo "[*] Starting Tor client..."
echo "[*] Press Ctrl+C to stop both client and bridge."
$TOR_BINARY -f "$CLIENT_TORRC" &
CLIENT_PID=$!

wait $CLIENT_PID
