package checks

func registerFlowSafetyChecks(r *Registry) {
	r.Register(newCheckFS1()) // Stderr vs stdout discipline
	r.Register(newCheckFS2()) // Non-TTY detection (no ANSI in pipes)
	r.Register(newCheckFS3()) // No interactive prompts in non-TTY
	r.Register(newCheckFS4()) // Env var auth support
	r.Register(newCheckFS5()) // No mandatory interactive auth
	r.Register(newCheckFS6()) // Exit codes
}

func registerTokenEfficiencyChecks(r *Registry) {
	r.Register(newCheckTE1()) // JSON output support
	r.Register(newCheckTE2()) // Stdin/pipe input support
	r.Register(newCheckTE3()) // --no-color flag
	r.Register(newCheckTE4()) // --quiet / --silent flag
	r.Register(newCheckTE5()) // Pagination support
	r.Register(newCheckTE6()) // Field masks / response filtering
	r.Register(newCheckTE7()) // Help output size
	r.Register(newCheckTE8()) // Concise output mode
}

func registerSelfDescribingChecks(r *Registry) {
	r.Register(newCheckSD1()) // Error format is structured
	r.Register(newCheckSD2()) // Version output is parseable
	r.Register(newCheckSD3()) // Shell completions available
	r.Register(newCheckSD4()) // Schema / describe introspection
	r.Register(newCheckSD5()) // Skill / context files
	r.Register(newCheckSD6()) // Help text with usage examples
	r.Register(newCheckSD7()) // Actionable error messages
	r.Register(newCheckSD8()) // Subcommand fan-out
}

func registerAutomationSafetyChecks(r *Registry) {
	r.Register(newCheckSA1()) // Confirmation bypass for destructive commands
	r.Register(newCheckSA2()) // Rejects path traversal
	r.Register(newCheckSA3()) // Rejects control characters
	r.Register(newCheckSA4()) // Dry-run support
	r.Register(newCheckSA5()) // Idempotency indicators
	r.Register(newCheckSA6()) // Read/write command separation
}

func registerPredictabilityChecks(r *Registry) {
	r.Register(newCheckPV1()) // Timeout flag
	r.Register(newCheckPV2()) // Retry / rate-limit hints
	r.Register(newCheckPV3()) // Deterministic output
	r.Register(newCheckPV4()) // Distinct exit codes for error classes
	r.Register(newCheckPV5()) // Reports actual effects
	r.Register(newCheckPV6()) // Long-running operation support
}
