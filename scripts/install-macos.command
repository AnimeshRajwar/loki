#!/usr/bin/env bash
# Double-clickable on macOS (Terminal will open). Make executable in ZIP/workflow.
DIR="$(cd "$(dirname "$0")" && pwd)"
sudo "$DIR/install-loki-system.sh" stable
read -n1 -r -p "Press any key to close..."
