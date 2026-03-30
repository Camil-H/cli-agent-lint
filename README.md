# cli-agent-lint

Audit CLI tools for AI agent-readiness. Checks 21 properties across 6 categories and produces a letter-grade scorecard.

## Install

```bash
go install github.com/cli-agent-lint/cli-agent-lint@latest
```

## Usage

```bash
# Audit a CLI tool
cli-agent-lint check ./my-tool

# JSON output
cli-agent-lint check --output json ./my-tool

# Passive only (parse help text, no subprocess execution)
cli-agent-lint check --no-probe ./my-tool

# List all checks
cli-agent-lint checks

# Describe a specific check
cli-agent-lint checks SO-1
```

## Categories

| Category | Prefix | Checks | What it covers |
|---|---|---|---|
| Structured Output | SO | 5 | JSON output, stderr discipline, structured errors, version parsing, stdin support |
| Terminal Hygiene | TH | 5 | ANSI detection, --no-color, --quiet, prompt suppression, confirmation bypass |
| Input Validation | IV | 3 | Path traversal, control characters, dry-run support |
| Schema Discovery | SD | 4 | Shell completions, introspection, context files, usage examples |
| Auth | AU | 2 | Env var auth, non-interactive auth alternatives |
| Operational | OR | 7 | Exit codes, timeouts, pagination, retry hints, determinism, field filtering, exit code docs |

## Flags

```
--output, -o   Output format: text, json (default: text)
--no-color     Disable colored output
--quiet, -q    Suppress informational output
--no-probe     Passive checks only
--severity     Minimum severity: info, warn, fail (default: info)
--category     Filter by category
--skip         Skip specific check IDs (repeatable)
--timeout      Probe timeout (default: 5s)
```

## Exit codes

- **0** — all checks passed
- **1** — one or more fail-severity checks did not pass
- **2** — usage or runtime error

## Grading

| Grade | Score | Label |
|---|---|---|
| A | >= 90% | Agent-ready |
| B | >= 70% | Mostly ready, some gaps |
| C | >= 50% | Significant gaps |
| D | >= 30% | Major work needed |
| F | < 30% | Not agent-ready |

## Check methods

- **Passive** — analyzes --help text only, always safe
- **Active** — executes the target CLI with crafted input (disabled with `--no-probe`)

## Building from source

```bash
go build -o cli-agent-lint .
go test ./...
go vet ./...
```

## License

See [LICENSE](LICENSE).
