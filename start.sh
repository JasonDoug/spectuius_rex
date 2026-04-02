#!/bin/bash

# start.sh - Starts the Next.js dev server if not already running

# Check if port 3000 is already in use
PORT_IN_USE=$(lsof -i:3000 -t)

# Check if a 'next dev' process is already running
NEXT_RUNNING=$(pgrep -f "next dev")

if [ -n "$NEXT_RUNNING" ] || [ -n "$PORT_IN_USE" ]; then
  echo "Next.js dev server is already running (PID: $NEXT_RUNNING)."
  exit 0
fi

echo "Starting Next.js dev server..."

# Explicitly load .env.local if it exists
if [ -f .env.local ]; then
  echo "Loading environment from .env.local"
  export $(grep -v '^#' .env.local | xargs)
fi

# Run in background and redirect output to a log file
npm run dev > dev-server.log 2>&1 &

# Wait a moment to check if it started successfully
sleep 2
NEW_PID=$(pgrep -f "next dev")

if [ -n "$NEW_PID" ]; then
  echo "Dev server started successfully in background (PID: $NEW_PID)."
  echo "Logs are being written to dev-server.log"
else
  echo "Failed to start dev server. Check dev-server.log for details."
  exit 1
fi
