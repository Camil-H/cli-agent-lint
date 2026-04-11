# cli-agent-lint

Audit CLI tools for AI agent-readiness. Runs 34 checks across 5 categories and produces a letter-grade scorecard telling you how well your CLI plays with autonomous agents.

## Demo

**Google Workspace CLI**

![cli-agent-lint auditing gws](demo/demo-gws.gif)

**GitHub CLI**

![cli-agent-lint auditing gh](demo/demo-gh.gif)

## Install

```bash
go install github.com/Camil-H/cli-agent-lint@latest
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
cli-agent-lint checks TE-1
```

## What it checks

Every check targets a specific property that matters when an AI agent drives a CLI non-interactively. Checks are organized into five categories, each addressing a distinct failure mode agents hit in practice.

| Category | ID | TL;DR | Why it matters |
|---|---|---|---|
| **Flow Safety** | `FS-*` | Will this CLI block my agent? | An agent that gets stuck on an interactive prompt, can't authenticate, or can't tell success from failure is a dead agent. These checks verify your CLI won't block an autonomous loop &mdash; no TTY prompts, non-interactive auth paths, clean stderr/stdout separation, and reliable exit codes. |
| **Token Efficiency** | `TE-*` | Will this CLI blow my context window? | Every byte of CLI output eats into the agent's context window. Verbose output fills that window fast, degrading the agent's ability to follow instructions and reason about results. These checks ensure your CLI supports JSON output, `--quiet`/`--no-color` modes, pagination for list commands, and field filtering &mdash; so agents get signal, not noise. |
| **Self-Describing** | `SD-*` | Can my agent learn this CLI from --help alone? | Agents learn your CLI by reading `--help`. If your help text lacks examples, your errors are cryptic, or there's no way to discover command schemas, the agent resorts to trial-and-error &mdash; wasting tokens and breaking flows. These checks verify your CLI teaches agents how to use it correctly on the first try. |
| **Automation Safety** | `SA-*` | Can my agent use this CLI without causing damage? | Agents retry failed commands, pass untrusted input, and can't read "Are you sure?" prompts. These checks verify your CLI has `--yes`/`--force` flags on destructive commands, rejects path traversal and control characters, and supports `--dry-run` so agents can preview before committing. |
| **Predictability** | `PV-*` | Can my agent trust the output? | An agent that can't trust its tool output will second-guess every step. These checks verify your CLI produces deterministic output, documents distinct exit codes for different error classes, supports `--timeout` for time-bounded execution, and surfaces retry/rate-limit information. |

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
