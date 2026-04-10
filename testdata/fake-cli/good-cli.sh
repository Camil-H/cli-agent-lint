#!/bin/bash
# good-cli.sh — passes all checks
# Implements every feature cli-agent-lint looks for.

set -euo pipefail

OUTPUT_FORMAT="text"
NO_COLOR=""
QUIET=""

# Parse global flags first
ARGS=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --output=*) OUTPUT_FORMAT="${1#*=}"; shift ;;
    --output|-o) OUTPUT_FORMAT="$2"; shift 2 ;;
    --no-color) NO_COLOR=1; shift ;;
    --quiet|-q) QUIET=1; shift ;;
    --version|-v)
      if [[ "${OUTPUT_FORMAT}" == "json" ]]; then
        echo '{"version":"1.2.3"}' >&1
      else
        echo "1.2.3"
      fi
      exit 0
      ;;
    --timeout) shift 2 ;; # accept but ignore
    --help|-h)
      ARGS=("help")
      shift
      ;;
    --) shift; ARGS+=("$@"); break ;;
    *) ARGS+=("$1"); shift ;;
  esac
done

CMD="${ARGS[0]:-help}"

emit_error() {
  local msg="$1"
  if [[ "$OUTPUT_FORMAT" == "json" ]]; then
    echo "{\"error\":\"$msg\",\"message\":\"$msg\"}" >&2
  else
    echo "Error: $msg" >&2
  fi
}

case "$CMD" in
  help|--help|-h)
    cat <<'HELP'
good-cli - A well-behaved CLI tool

Usage:
  good-cli [command] [flags]

Available Commands:
  list        List all items
  get         Get a specific item
  create      Create a new item
  delete      Delete an item
  update      Update an existing item
  completion  Generate shell completions
  checks      List available checks
  login       Log in to the service
  version     Print version

Flags:
  -o, --output string    Output format {text,json,brief} (default "text")
      --no-color          Disable colored output
  -q, --quiet            Suppress informational output
      --brief             Show concise output
      --timeout duration  Request timeout (default 30s)
      --version          Print version
      --retry int        Number of retries (default 0)
  -h, --help             Show help

Environment Variables:
  GOOD_CLI_TOKEN      Authentication token
  GOOD_CLI_API_KEY    API key for authentication

Examples:
  good-cli list --output json
  good-cli create --name my-item --yes
  echo '{"name":"x"}' | good-cli create --from-file -

Exit Codes:
  0  Success
  1  General error
  2  Usage error

Use "good-cli [command] --help" for more information about a command.
HELP
    exit 0
    ;;

  list)
    cat <<'HELP'
List all items

Usage:
  good-cli list [flags]

Flags:
      --limit int      Maximum items to return (default 100)
      --offset int     Offset for pagination
      --page-size int  Items per page
      --page-all       Fetch all pages
      --cursor string  Pagination cursor
      --fields string  Comma-separated list of fields to include
      --filter string  Filter expression
      --jq string      JQ expression for output filtering
  -h, --help           Show help
HELP
    exit 0
    ;;

  get)
    cat <<'HELP'
Get a specific item

Usage:
  good-cli get <id> [flags]

Flags:
      --fields string  Comma-separated list of fields to include
  -h, --help           Show help
HELP
    exit 0
    ;;

  create)
    cat <<'HELP'
Create a new item

Usage:
  good-cli create [flags]

Reads from stdin when --from-file - is specified.

Flags:
      --name string       Item name
      --dry-run           Preview the operation without making changes
      --file string       Path to input file
      --from-file string  Read input from file (use - for stdin)
  -y, --yes               Skip confirmation prompt
  -h, --help              Show help
HELP
    exit 0
    ;;

  delete)
    cat <<'HELP'
Delete an item

Usage:
  good-cli delete <id> [flags]

Flags:
      --dry-run    Preview the operation without making changes
      --force      Skip confirmation
  -h, --help       Show help
HELP
    exit 0
    ;;

  update)
    cat <<'HELP'
Update an existing item

Usage:
  good-cli update <id> [flags]

Flags:
      --name string  New name
      --dry-run      Preview the operation without making changes
  -y, --yes          Skip confirmation prompt
  -h, --help         Show help
HELP
    exit 0
    ;;

  completion)
    echo "# shell completion output"
    exit 0
    ;;

  checks)
    echo "TE-1  JSON output support"
    exit 0
    ;;

  login)
    cat <<'HELP'
Log in to the service

Usage:
  good-cli login [flags]

Flags:
      --token string          Use a token instead of interactive login
      --with-token            Read token from stdin
      --service-account file  Path to service account key file
  -h, --help                  Show help

Environment Variables:
  GOOD_CLI_TOKEN    Authentication token
HELP
    exit 0
    ;;

  version)
    echo "1.2.3"
    exit 0
    ;;

  *)
    emit_error "unknown command \"$CMD\". Run 'good-cli --help' for available commands."
    exit 1
    ;;
esac
