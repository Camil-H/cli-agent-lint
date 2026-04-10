# cli-agent-lint — Agent Skill File

## What this tool does

cli-agent-lint audits CLI tools for AI agent-readiness. It checks across five categories: flow safety, token efficiency, self-describing, automation safety, and predictability.

## Usage

```bash
# Audit a CLI tool
cli-agent-lint check <target-cli>

# JSON output (machine-readable)
cli-agent-lint check <target-cli> --output json

# Passive checks only (no subprocess execution)
cli-agent-lint check <target-cli> --no-probe

# List all checks
cli-agent-lint checks

# Describe a specific check
cli-agent-lint checks <check-id>
```

## Exit codes

- 0: all checks passed
- 1: one or more fail-severity checks did not pass
- 2: cli-agent-lint itself errored

## Key flags

- `--output json` — structured JSON report to stdout
- `--no-color` — disable ANSI colors
- `--quiet` — suppress diagnostics, only essential output
- `--no-probe` — passive checks only, no subprocess execution
- `--severity <level>` — minimum severity: info, warn, fail
- `--category <name>` — filter by check category
- `--skip <id>` — skip specific checks (repeatable)
- `--timeout <duration>` — probe timeout (default 5s)

## Interpreting results

The JSON report contains a `score` object with `earned`, `total`, `percentage`, and `grade` (A-F). Each check in the `checks` array has `id`, `status` (pass/fail/warn/skip), `severity`, and `recommendation`.

## Conventions

- Errors and diagnostics go to stderr
- Data output goes to stdout
- No interactive prompts — all input via flags/args
- Respects NO_COLOR env var
