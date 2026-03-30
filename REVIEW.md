# REVIEW.md — cli-agent-lint

*Generated 2026-03-15 by four specialist reviewers. Triaged 2026-03-30.*

---

## Done

| ID | Finding | Resolution |
|----|---------|------------|
| GO-1 | Prober is a concrete struct | Extracted `probe.Runner` interface + `checks.Prober` type alias |
| GO-2 | GetIndex data race on lazy init | Removed lazy path; runner pre-computes the index |
| GO-3 | Type assertions instead of `errors.As` | All three locations now use `errors.As` |
| GO-4 | Direct deps marked `// indirect` | go.mod is clean (cobra/isatty direct, pflag/mousetrap/sys indirect) |
| GO-7 | `detectOutputFormat` reads `os.Args` directly | Extracted `detectOutputFormatFromArgs(opts, []string)` |
| GO-10/READ-1 | Inconsistent `StatusFail` naming | All constants now prefixed: `StatusPass`, `StatusFail`, `StatusSkip`, `StatusError` |
| SEC-2/SEC-8 | Env var filter gaps (`_SECRET=` dead code, missing `_API_KEY`) | Substrings fixed to `_SECRET`, `_API_KEY`, `_TOKEN` (no trailing `=`) |
| SEC-3 | `Opts.Env` bypass reintroduces secrets | `buildEnv()` now filters extras through `isSensitiveEnvVar` |
| SEC-4 | Unsanitized target output in text reports | `stripANSI(res.Detail)` applied in text formatter |
| SEC-7 | IV-1 uses real sensitive path (`/etc/passwd`) | Changed to `../../tmp/.cli-agent-lint-test-NONEXISTENT` |
| READ-2/PERF-8 | Duplicate `hasFileFlag`/`fileArgFlag` | Merged into single `fileArgFlag` returning flag name or empty string |
| READ-3 | Missing nil guards in TH-2, TH-3, TH-4, IV-3 | Nil guards added; active checks use `skipIfNoProber` helper |
| READ-4/GO-6/PERF-5 | Hardcoded SO-3 dependency in runner | Removed; passive→active ordering ensures dependencies are satisfied |
| READ-6 | `hasAuthTokenFlag` appears unused | Does not exist in codebase (stale reference) |
| READ-7 | `json.Marshal` error ignored in output.go | Error is handled with fallback to plain text |
| READ-8/PERF-6 | Unbuffered text formatter / ignored write errors | Uses `bufio.Writer` with `Flush()` error check |
| READ-9 | Multiple `GetIndex()` calls without caching | All checks cache in local variable |
| READ-10 | Magic concurrency limit 8 | Extracted to `defaultDiscoveryConcurrency` constant |
| READ-11 | `Summary.Error` excluded from JSON | `jsonSummary` now includes `Error` field |
| PERF-7 | `truncate()` always converts to `[]rune` | Added `len(s) <= maxLen` fast path before rune conversion |
| GO-13 | `go.mod` uses patch version | go.mod now has `go 1.25.0` (valid since Go 1.21+) |
| GO-8 | No concurrent test for ResultSet | `TestResultSet_ConcurrentAccess` in `checks/check_test.go` |
| PERF-3 | Active check concurrency defaults to `runtime.NumCPU()` | Changed to `max(4, runtime.NumCPU())` in `runner/runner.go` |
| SEC-1 | TH-4 executes mutating commands with real side effects | TH-4 now prefers non-mutating (list-like) commands; never runs mutating commands bare |
| SEC-6 | Unbounded subcommand discovery | Added `MaxCommands` (default 1000) to `DiscoverOpts` with `atomic.Int32` counter |
| PERF-1/PERF-2 | Pre-compute file/string-input commands in index | Added `FileAccepting()`, `StringInput()`, `FileArgFlag()`, `StringInputFlag()` to `CommandIndex`; removed redundant tree walks from IV-1/IV-2 |
| PERF-4 | O(flags × names) per-command flag lookup | Added `cmdFlags` map and `CmdHasFlag()` to `CommandIndex`; replaced `cmd.HasFlag()` in TH-5, IV-3, OR-3 |
| SEC-11 | Dependency freshness | Updated cobra v1.8.1→v1.10.2, pflag v1.0.5→v1.0.10 |

---

## Not To Do / Not Relevant

| ID | Finding | Reason |
|----|---------|--------|
| SEC-5 | `exec.LookPath` PATH-based binary injection | Standard Go behavior. The user explicitly provides the target path. Requiring absolute paths hurts usability. |
| SEC-9 | Process kill race (double-kill) | The explicit `syscall.Kill(-pgid)` handles process groups that Go's `CommandContext` does not kill. Intentional and safe. |
| SEC-10 | 10 MiB output limit is generous | Theoretical 160 MiB peak is unlikely in practice. Reducing to 1 MiB risks truncating legitimate output from large CLIs. |
| GO-5 | Passive checks run with unbounded concurrency | Passive checks only parse in-memory help text (no I/O). Bounding them adds complexity without benefit. |
| GO-9 | Formatter types lack shared interface | Only two formatters, selected by a simple conditional. An interface adds indirection without value at this scale. |
| GO-11 | `Category` string type has no validation | Categories are package-level constants used internally. Runtime validation adds ceremony without catching real bugs. |
| GO-12 | Missing capacity hints on index slices | Negligible for typical CLI sizes (tens to low hundreds of commands). |
| PERF-9 | AU-1/AU-2 both scan for auth env vars | Different logic paths; caching would couple the checks. The regex runs on in-memory strings — fast enough. |
| READ-5 | `discovery.go` is 592 lines mixing concerns | Cohesive around help-text parsing and discovery. Splitting into 3 files for ~500 lines adds navigation overhead. |

---

## To Do

### Low

**PERF-10: Discovery concurrency not configurable**
`discovery/discovery.go` — The semaphore is a named constant (8) but not exposed via `DiscoverOpts`. For I/O-bound work on fast machines, 8 may be conservative.
