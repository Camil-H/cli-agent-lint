# cli-agent-lint

Audit CLI tools for AI agent-readiness. Runs 26 checks across 6 categories and produces a letter-grade scorecard telling you how well your CLI plays with autonomous agents.

## Demo

**Google Workspace CLI**

![cli-agent-lint auditing gws](demo/demo-gws.gif)

**GitHub CLI**

![cli-agent-lint auditing gh](demo/demo-gh.gif)

## Install

```bash
go install github.com/cli-agent-lint/cli-agent-lint@latest
```

## Quick start

```bash
# Audit a CLI tool
cli-agent-lint check ./my-tool

# JSON output
cli-agent-lint check --output json ./my-tool

# Passive only — parse help text, never execute the target
cli-agent-lint check --no-probe ./my-tool

# List all checks
cli-agent-lint checks

# Describe a specific check
cli-agent-lint checks SO-1
```

## What it checks

Every check targets a specific property that matters when an AI agent drives a CLI non-interactively.

| Category | ID | Checks | What it covers |
|---|---|---|---|
| Structured Output | `SO-*` | 5 | JSON output, stderr discipline, structured errors, version parsing, stdin support |
| Terminal Hygiene | `TH-*` | 5 | ANSI detection, `--no-color`, `--quiet`, prompt suppression, confirmation bypass |
| Input Validation | `IV-*` | 3 | Path traversal, control characters, dry-run support |
| Schema & Discovery | `SD-*` | 4 | Shell completions, introspection, context files, usage examples |
| Auth | `AU-*` | 2 | Env-var auth, non-interactive auth alternatives |
| Operational Robustness | `OR-*` | 7 | Exit codes, timeouts, pagination, retry hints, determinism, field filtering |

### Check methods

- **Passive** — analyzes `--help` text only; always safe, zero side-effects.
- **Active** — executes the target CLI with crafted input. Disable with `--no-probe`.

## Grading

| Grade | Score | Meaning |
|:---:|:---:|---|
| **A** | &ge; 90% | Agent-ready |
| **B** | &ge; 70% | Mostly ready, some gaps |
| **C** | &ge; 50% | Significant gaps |
| **D** | &ge; 30% | Major work needed |
| **F** | &lt; 30% | Not agent-ready |

## Flags

| Flag | Description | Default |
|---|---|---|
| `--output, -o` | Output format: `text`, `json` | `text` |
| `--no-color` | Disable colored output | |
| `--quiet, -q` | Suppress informational output | |
| `--no-probe` | Passive checks only | |
| `--severity` | Minimum severity to report: `info`, `warn`, `fail` | `info` |
| `--category` | Run checks from a single category | |
| `--skip` | Skip specific check IDs (repeatable) | |
| `--timeout` | Probe command timeout | `5s` |

## Exit codes

| Code | Meaning |
|:---:|---|
| `0` | All checks passed |
| `1` | One or more fail-severity checks did not pass |
| `2` | Usage or runtime error |

## Building from source

```bash
go build -o cli-agent-lint .
go test ./...
go vet ./...
```

## License

MIT &mdash; see [LICENSE](LICENSE).
