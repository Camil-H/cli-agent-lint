#!/bin/bash
# bad-cli.sh — fails all checks
# Does everything wrong: ANSI in pipes, errors on stdout, exit 0 on error,
# prompts, no JSON, no --no-color, etc.

CMD="${1:-help}"

case "$CMD" in
  --help|-h|help)
    # ANSI escape codes in output even when piped
    echo -e "\x1b[1m\x1b[32mbad-cli\x1b[0m - A poorly behaved CLI tool"
    echo ""
    echo "Usage:"
    echo "  bad-cli [command]"
    echo ""
    echo "Available Commands:"
    echo "  list        List items"
    echo "  create      Create an item"
    echo "  delete      Delete an item"
    echo "  login       Log in"
    echo ""
    echo "Flags:"
    echo "  -h, --help  Show help"
    exit 0
    ;;

  --version)
    echo "bad-cli version 2.0.0 (built on 2025-01-01, commit abc123, go1.21)"
    exit 0
    ;;

  list)
    echo "Usage: bad-cli list"
    exit 0
    ;;

  create)
    # Prompts for input even when stdin is not a TTY
    echo -n "Enter item name: "
    read -r name
    echo "Created: $name"
    exit 0
    ;;

  delete)
    echo "Usage: bad-cli delete <id>"
    exit 0
    ;;

  login)
    echo -n "Enter username: "
    read -r user
    echo -n "Enter password: "
    read -rs pass
    echo ""
    echo "Logged in as $user"
    exit 0
    ;;

  *)
    # Error on stdout, exit 0
    echo "Error: unknown command \"$CMD\""
    exit 0
    ;;
esac
