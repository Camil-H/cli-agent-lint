#!/bin/bash
# interactive-cli.sh — prompts on stdin unconditionally

CMD="${1:-help}"

case "$CMD" in
  --help|-h|help)
    cat <<'HELP'
interactive-cli - A CLI that always prompts

Usage:
  interactive-cli [command]

Available Commands:
  create   Create an item
  setup    Initial setup

Flags:
  -h, --help  Show help
HELP
    exit 0
    ;;

  --version)
    echo "1.0.0"
    exit 0
    ;;

  create)
    echo -n "Enter item name: "
    read -r name
    echo "Created: $name"
    exit 0
    ;;

  setup)
    echo -n "Enter your API key: "
    read -r key
    echo "Configured with key $key"
    exit 0
    ;;

  *)
    echo "Error: unknown command \"$CMD\"" >&2
    exit 1
    ;;
esac
