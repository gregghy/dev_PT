#!/bin/bash

# --- Binary and Config Paths ---
TOR_BINARY="/usr/bin/tor"
PEER_1_BINARY="./peer_1" # Path to your peer1 executable
PEER_2_BINARY="./peer_2" # Path to your peer2 executable
BRIDGE_TORRC="torrc-bridge"
CLIENT_TORRC="torrc-client"
BRIDGE_DATA_DIR="./tor-bridge-data"
FINGERPRINT_FILE="${BRIDGE_DATA_DIR}/fingerprint"

# --- Peer and Bridge Network Config ---
PEER_1_IP="127.0.0.1"
PEER_1_PORT="9999" # From peer_1_config.json

cleanup() {
  echo -e "\n[*] Shutting down all processes..."
  # Kill processes in reverse order of startup
  if [[ -n "$CLIENT_PID" ]]; then kill "$CLIENT_PID" 2>/dev/null; fi
  if [[ -n "$PEER_1_PID" ]]; then kill "$PEER_1_PID" 2>/dev/null; fi
  if [[ -n "$PEER_2_PID" ]]; then kill "$PEER_2_PID" 2>/dev/null; fi
  if [[ -n "$BRIDGE_PID" ]]; then kill "$BRIDGE_PID" 2>/dev/null; fi

  echo "[*] Removing temporary data directories and logs..."
  rm -rf ./tor-client-data ./tor-bridge-data
  rm -f tor-bridge-log.txt tor-client-log.txt torrc-client.bak

  echo "[*] Cleanup complete."
  exit 0
}

trap cleanup SIGINT SIGTERM EXIT

# 1. Start Tor Bridge
echo "[*] Starting Tor bridge in the background..."
$TOR_BINARY -f "$BRIDGE_TORRC" &
BRIDGE_PID=$!
sleep 2 # Give it a moment to start

# 2. Start Peer Processes
echo "[*] Starting peer2..."
$PEER_2_BINARY peer_2_config.json &
PEER_2_PID=$!
sleep 5

echo "[*] Starting peer1..."
$PEER_1_BINARY peer_1_config.json &
PEER_1_PID=$!
sleep 5 # Give peers a moment to initialize

# 3. Get Bridge Fingerprint
echo "[*] Waiting for the bridge to generate its fingerprint..."
while [ ! -f "$FINGERPRINT_FILE" ]; do
  if ! kill -0 "$BRIDGE_PID" 2>/dev/null; then
    echo "[!] ERROR: Bridge process died unexpectedly. Check logs." >&2
    exit 1
  fi
  sleep 1
done
FINGERPRINT=$(awk '{print $2}' "$FINGERPRINT_FILE")
if [ -z "$FINGERPRINT" ]; then
  echo "[!] ERROR: Could not extract fingerprint." >&2
  exit 1
fi
echo "[+] Extracted Bridge Fingerprint: $FINGERPRINT"

# 4. Update Client Config to Point to Peer1
echo "[*] Updating client torrc to use peer1 as the entry point..."
sed -i.bak "s/^Bridge netshaper .*/Bridge netshaper ${PEER_1_IP}:${PEER_1_PORT} ${FINGERPRINT}/" "$CLIENT_TORRC"

# 5. Start Tor Client
echo "[*] Starting Tor client..."
echo "[*] Setup complete. Press Ctrl+C to stop all processes."
$TOR_BINARY -f "$CLIENT_TORRC" &
CLIENT_PID=$!

wait $CLIENT_PID
