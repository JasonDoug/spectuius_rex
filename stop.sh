#!/bin/bash

# stop.sh - Stops any running Next.js dev server instances

echo "Stopping Next.js dev server..."

# Find and kill processes matching 'next dev'
PIDS=$(pgrep -f "next dev")

if [ -z "$PIDS" ]; then
  echo "No running dev server found."
else
  echo "Killing PIDs: $PIDS"
  pkill -f "next dev"
  echo "Dev server stopped."
fi
