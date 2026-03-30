#!/bin/bash
# no-json-cli.sh — partial compliance, specifically missing --output json
# Has good exit codes, stderr discipline, no ANSI, etc. but no JSON output.

CMD="${1:-help}"

case "$CMD" in
  --help|-h|help)
    cat <<'HELP'
no-json-cli - A CLI without JSON output

Usage:
  no-json-cli [command]

Available Commands:
  list        List items
  create      Create an item
  version     Print version

Flags:
      --no-color  Disable colored output
  -q, --quiet     Suppress informational output
  -h, --help      Show help
HELP
    exit 0
    ;;

  --version)
    echo "0.5.0"
    exit 0
    ;;

  list)
    cat <<'HELP'
List items

Usage:
  no-json-cli list

Flags:
  -h, --help  Show help
HELP
    exit 0
    ;;

  create)
    cat <<'HELP'
Create an item

Usage:
  no-json-cli create --name <name>

Flags:
      --name string  Item name
      --dry-run      Preview changes
  -h, --help         Show help
HELP
    exit 0
    ;;

  *)
    echo "Error: unknown command \"$CMD\"" >&2
    exit 1
    ;;
esac
