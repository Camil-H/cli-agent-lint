#!/bin/bash
# noisy-cli.sh — ANSI output + errors on stdout

CMD="${1:-help}"

case "$CMD" in
  --help|-h|help)
    echo -e "\x1b[1mnoisy-cli\x1b[0m - A noisy CLI"
    echo ""
    echo "Usage:"
    echo "  noisy-cli [command]"
    echo ""
    echo "Available Commands:"
    echo "  list   List items"
    echo ""
    echo "Flags:"
    echo "  -h, --help  Show help"
    exit 0
    ;;

  --version)
    echo -e "\x1b[1mnoisy-cli\x1b[0m version 1.0.0"
    exit 0
    ;;

  list)
    echo "Usage: noisy-cli list"
    exit 0
    ;;

  *)
    # Errors go to stdout
    echo "Error: unknown command \"$CMD\""
    exit 0
    ;;
esac
