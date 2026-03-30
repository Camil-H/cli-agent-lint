package probe

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// isSensitiveEnvVar
// ---------------------------------------------------------------------------

func TestIsSensitiveEnvVar_Prefixes(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{"AWS_SECRET exact", "AWS_SECRET", true},
		{"AWS_SECRET_ACCESS_KEY", "AWS_SECRET_ACCESS_KEY", true},
		{"AWS_SESSION_TOKEN", "AWS_SESSION_TOKEN", true},
		{"GITHUB_TOKEN", "GITHUB_TOKEN", true},
		{"GH_TOKEN", "GH_TOKEN", true},
		{"GITLAB_TOKEN", "GITLAB_TOKEN", true},
		{"DOCKER_PASSWORD", "DOCKER_PASSWORD", true},
		{"DOCKER_AUTH_CONFIG", "DOCKER_AUTH_CONFIG", true},
		{"NPM_TOKEN", "NPM_TOKEN", true},
		{"NUGET_API_KEY", "NUGET_API_KEY", true},
		{"PYPI_TOKEN", "PYPI_TOKEN", true},
		{"RUBYGEMS_API_KEY", "RUBYGEMS_API_KEY", true},
		{"CODECOV_TOKEN", "CODECOV_TOKEN", true},
		{"SONAR_TOKEN", "SONAR_TOKEN", true},
		{"SNYK_TOKEN", "SNYK_TOKEN", true},
		{"SENTRY_AUTH_TOKEN", "SENTRY_AUTH_TOKEN", true},
		{"SLACK_TOKEN", "SLACK_TOKEN", true},
		{"SLACK_WEBHOOK", "SLACK_WEBHOOK", true},
		{"SLACK_WEBHOOK_URL", "SLACK_WEBHOOK_URL", true},
		{"TWILIO_AUTH_TOKEN", "TWILIO_AUTH_TOKEN", true},
		{"SENDGRID_API_KEY", "SENDGRID_API_KEY", true},
		{"STRIPE_SECRET", "STRIPE_SECRET", true},
		{"STRIPE_SECRET_KEY", "STRIPE_SECRET_KEY", true},
		{"DATABASE_PASSWORD", "DATABASE_PASSWORD", true},
		{"DB_PASSWORD", "DB_PASSWORD", true},
		{"REDIS_PASSWORD", "REDIS_PASSWORD", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSensitiveEnvVar(tt.env); got != tt.want {
				t.Errorf("isSensitiveEnvVar(%q) = %v, want %v", tt.env, got, tt.want)
			}
		})
	}
}

func TestIsSensitiveEnvVar_Substrings(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{"contains _SECRET_", "MY_SECRET_VAR", true},
		{"contains _PASSWORD", "MYSQL_PASSWORD", true},
		{"contains _PRIVATE_KEY", "SSH_PRIVATE_KEY", true},
		{"contains _CREDENTIALS", "GCP_CREDENTIALS", true},
		{"contains _SECRET", "MY_SECRET", true},
		{"contains _API_KEY", "OPENAI_API_KEY", true},
		{"contains _API_KEY anthropic", "ANTHROPIC_API_KEY", true},
		{"contains _TOKEN", "CUSTOM_TOKEN", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSensitiveEnvVar(tt.env); got != tt.want {
				t.Errorf("isSensitiveEnvVar(%q) = %v, want %v", tt.env, got, tt.want)
			}
		})
	}
}

func TestIsSensitiveEnvVar_CaseInsensitive(t *testing.T) {
	// The function uppercases the name before matching.
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{"lowercase github_token", "github_token", true},
		{"mixed case Aws_Secret_Access_Key", "Aws_Secret_Access_Key", true},
		{"lowercase my_password", "my_password", true},
		{"lowercase ssh_private_key", "ssh_private_key", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSensitiveEnvVar(tt.env); got != tt.want {
				t.Errorf("isSensitiveEnvVar(%q) = %v, want %v", tt.env, got, tt.want)
			}
		})
	}
}

func TestIsSensitiveEnvVar_SafeVars(t *testing.T) {
	safeVars := []string{
		"PATH",
		"HOME",
		"PAGER",
		"TERM",
		"SHELL",
		"USER",
		"LANG",
		"LC_ALL",
		"EDITOR",
		"GOPATH",
		"GOROOT",
		"XDG_CONFIG_HOME",
	}
	for _, name := range safeVars {
		t.Run(name, func(t *testing.T) {
			if isSensitiveEnvVar(name) {
				t.Errorf("isSensitiveEnvVar(%q) = true, want false", name)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// buildBaseEnv
// ---------------------------------------------------------------------------

func TestBuildBaseEnv_FiltersSensitiveVars(t *testing.T) {
	// Save and restore the environment around the test.
	origEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, e := range origEnv {
			k, v, _ := strings.Cut(e, "=")
			os.Setenv(k, v)
		}
	}()

	os.Clearenv()
	os.Setenv("PATH", "/usr/bin")
	os.Setenv("HOME", "/home/test")
	os.Setenv("GITHUB_TOKEN", "ghp_supersecret")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "wJalrXUtnFEMI")
	os.Setenv("MY_PASSWORD", "hunter2")

	env := buildBaseEnv()
	envMap := envSliceToMap(env)

	// Safe vars should be present.
	if v, ok := envMap["PATH"]; !ok || v != "/usr/bin" {
		t.Errorf("PATH missing or wrong: got %q, ok=%v", v, ok)
	}
	if v, ok := envMap["HOME"]; !ok || v != "/home/test" {
		t.Errorf("HOME missing or wrong: got %q, ok=%v", v, ok)
	}

	// Sensitive vars should be filtered.
	for _, sensitive := range []string{"GITHUB_TOKEN", "AWS_SECRET_ACCESS_KEY", "MY_PASSWORD"} {
		if _, ok := envMap[sensitive]; ok {
			t.Errorf("sensitive var %q should have been filtered", sensitive)
		}
	}
}

func TestBuildBaseEnv_AppliesDefaults(t *testing.T) {
	origEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, e := range origEnv {
			k, v, _ := strings.Cut(e, "=")
			os.Setenv(k, v)
		}
	}()

	os.Clearenv()
	// Set PAGER to something non-default to verify it gets overridden.
	os.Setenv("PAGER", "less")
	os.Setenv("HOME", "/home/test")

	env := buildBaseEnv()
	envMap := envSliceToMap(env)

	// PAGER should be overridden to "cat".
	if v := envMap["PAGER"]; v != "cat" {
		t.Errorf("PAGER = %q, want %q", v, "cat")
	}

	// GIT_PAGER and TERM should be added even if not in original env.
	if v := envMap["GIT_PAGER"]; v != "cat" {
		t.Errorf("GIT_PAGER = %q, want %q", v, "cat")
	}
	if v := envMap["TERM"]; v != "dumb" {
		t.Errorf("TERM = %q, want %q", v, "dumb")
	}

	// HOME should pass through unchanged.
	if v := envMap["HOME"]; v != "/home/test" {
		t.Errorf("HOME = %q, want %q", v, "/home/test")
	}
}

func TestBuildBaseEnv_AddsDefaultsNotInBase(t *testing.T) {
	origEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, e := range origEnv {
			k, v, _ := strings.Cut(e, "=")
			os.Setenv(k, v)
		}
	}()

	// Start with a completely clean environment — no PAGER, GIT_PAGER, TERM.
	os.Clearenv()
	os.Setenv("PATH", "/usr/bin")

	env := buildBaseEnv()
	envMap := envSliceToMap(env)

	for k, want := range envDefaults {
		got, ok := envMap[k]
		if !ok {
			t.Errorf("default %q missing from env", k)
			continue
		}
		if got != want {
			t.Errorf("default %q = %q, want %q", k, got, want)
		}
	}
}

// ---------------------------------------------------------------------------
// Prober.buildEnv with extras
// ---------------------------------------------------------------------------

func TestProberBuildEnv_NoExtras(t *testing.T) {
	base := []string{"PATH=/usr/bin", "HOME=/home/test", "PAGER=cat"}
	p := &Prober{baseEnv: base}

	got := p.buildEnv(nil)
	// With no extras, buildEnv returns the cached baseEnv slice directly.
	if &got[0] != &base[0] {
		t.Error("expected buildEnv(nil) to return the same base slice, got a copy")
	}
}

func TestProberBuildEnv_EmptyExtras(t *testing.T) {
	base := []string{"PATH=/usr/bin", "HOME=/home/test"}
	p := &Prober{baseEnv: base}

	got := p.buildEnv([]string{})
	// Empty slice also returns base directly.
	if &got[0] != &base[0] {
		t.Error("expected buildEnv(empty) to return the same base slice, got a copy")
	}
}

func TestProberBuildEnv_OverridesExisting(t *testing.T) {
	base := []string{"PATH=/usr/bin", "HOME=/home/test", "PAGER=cat"}
	p := &Prober{baseEnv: base}

	got := p.buildEnv([]string{"PAGER=less", "HOME=/tmp"})
	envMap := envSliceToMap(got)

	if envMap["PAGER"] != "less" {
		t.Errorf("PAGER = %q, want %q", envMap["PAGER"], "less")
	}
	if envMap["HOME"] != "/tmp" {
		t.Errorf("HOME = %q, want %q", envMap["HOME"], "/tmp")
	}
	// PATH should be unchanged.
	if envMap["PATH"] != "/usr/bin" {
		t.Errorf("PATH = %q, want %q", envMap["PATH"], "/usr/bin")
	}
}

func TestProberBuildEnv_AddsNewVars(t *testing.T) {
	base := []string{"PATH=/usr/bin"}
	p := &Prober{baseEnv: base}

	got := p.buildEnv([]string{"MY_VAR=hello", "OTHER=world"})
	envMap := envSliceToMap(got)

	if envMap["PATH"] != "/usr/bin" {
		t.Errorf("PATH = %q, want %q", envMap["PATH"], "/usr/bin")
	}
	if envMap["MY_VAR"] != "hello" {
		t.Errorf("MY_VAR = %q, want %q", envMap["MY_VAR"], "hello")
	}
	if envMap["OTHER"] != "world" {
		t.Errorf("OTHER = %q, want %q", envMap["OTHER"], "world")
	}
}

func TestProberBuildEnv_OverrideAndAdd(t *testing.T) {
	base := []string{"PATH=/usr/bin", "TERM=dumb"}
	p := &Prober{baseEnv: base}

	got := p.buildEnv([]string{"TERM=xterm", "CUSTOM=val"})
	envMap := envSliceToMap(got)

	if envMap["TERM"] != "xterm" {
		t.Errorf("TERM = %q, want %q", envMap["TERM"], "xterm")
	}
	if envMap["CUSTOM"] != "val" {
		t.Errorf("CUSTOM = %q, want %q", envMap["CUSTOM"], "val")
	}
	if envMap["PATH"] != "/usr/bin" {
		t.Errorf("PATH = %q, want %q", envMap["PATH"], "/usr/bin")
	}
}

func TestProberBuildEnv_FiltersSensitiveExtras(t *testing.T) {
	base := []string{"PATH=/usr/bin"}
	p := &Prober{baseEnv: base}

	got := p.buildEnv([]string{"GITHUB_TOKEN=ghp_secret", "SAFE_VAR=ok"})
	envMap := envSliceToMap(got)

	if _, ok := envMap["GITHUB_TOKEN"]; ok {
		t.Error("sensitive var GITHUB_TOKEN in extras should have been filtered")
	}
	if envMap["SAFE_VAR"] != "ok" {
		t.Errorf("SAFE_VAR = %q, want %q", envMap["SAFE_VAR"], "ok")
	}
}

func TestProberBuildEnv_DoesNotMutateBase(t *testing.T) {
	base := []string{"PATH=/usr/bin", "HOME=/home/test"}
	p := &Prober{baseEnv: base}

	_ = p.buildEnv([]string{"PATH=/override"})

	// Verify base was not mutated.
	if base[0] != "PATH=/usr/bin" {
		t.Errorf("base[0] mutated to %q", base[0])
	}
	if base[1] != "HOME=/home/test" {
		t.Errorf("base[1] mutated to %q", base[1])
	}
}

// ---------------------------------------------------------------------------
// limitedWriter
// ---------------------------------------------------------------------------

func TestLimitedWriter_WritesUpToLimit(t *testing.T) {
	var buf bytes.Buffer
	lw := &limitedWriter{buf: &buf, remaining: 10}

	data := []byte("hello, world!") // 13 bytes
	n, err := lw.Write(data)

	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	// Should report full length to caller.
	if n != 13 {
		t.Errorf("Write returned n=%d, want 13", n)
	}
	// But buffer should only have 10 bytes.
	if buf.Len() != 10 {
		t.Errorf("buffer length = %d, want 10", buf.Len())
	}
	if buf.String() != "hello, wor" {
		t.Errorf("buffer content = %q, want %q", buf.String(), "hello, wor")
	}
	// remaining should be 0.
	if lw.remaining != 0 {
		t.Errorf("remaining = %d, want 0", lw.remaining)
	}
}

func TestLimitedWriter_DiscardsWhenZeroRemaining(t *testing.T) {
	var buf bytes.Buffer
	lw := &limitedWriter{buf: &buf, remaining: 0}

	data := []byte("all discarded")
	n, err := lw.Write(data)

	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if n != len(data) {
		t.Errorf("Write returned n=%d, want %d", n, len(data))
	}
	if buf.Len() != 0 {
		t.Errorf("buffer length = %d, want 0", buf.Len())
	}
}

func TestLimitedWriter_ExactFit(t *testing.T) {
	var buf bytes.Buffer
	lw := &limitedWriter{buf: &buf, remaining: 5}

	data := []byte("hello") // exactly 5 bytes
	n, err := lw.Write(data)

	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if n != 5 {
		t.Errorf("Write returned n=%d, want 5", n)
	}
	if buf.String() != "hello" {
		t.Errorf("buffer content = %q, want %q", buf.String(), "hello")
	}
	if lw.remaining != 0 {
		t.Errorf("remaining = %d, want 0", lw.remaining)
	}
}

func TestLimitedWriter_MultipleWrites(t *testing.T) {
	var buf bytes.Buffer
	lw := &limitedWriter{buf: &buf, remaining: 10}

	// First write: 6 bytes fit.
	n1, err := lw.Write([]byte("abcdef"))
	if err != nil {
		t.Fatalf("first Write error: %v", err)
	}
	if n1 != 6 {
		t.Errorf("first Write n=%d, want 6", n1)
	}

	// Second write: only 4 of 8 bytes fit.
	n2, err := lw.Write([]byte("ghijklmn"))
	if err != nil {
		t.Fatalf("second Write error: %v", err)
	}
	if n2 != 8 {
		t.Errorf("second Write n=%d, want 8 (full reported length)", n2)
	}

	// Third write: nothing fits, remaining is 0.
	n3, err := lw.Write([]byte("opqrst"))
	if err != nil {
		t.Fatalf("third Write error: %v", err)
	}
	if n3 != 6 {
		t.Errorf("third Write n=%d, want 6", n3)
	}

	if buf.Len() != 10 {
		t.Errorf("buffer length = %d, want 10", buf.Len())
	}
	if buf.String() != "abcdefghij" {
		t.Errorf("buffer content = %q, want %q", buf.String(), "abcdefghij")
	}
}

func TestLimitedWriter_SmallWrites(t *testing.T) {
	var buf bytes.Buffer
	lw := &limitedWriter{buf: &buf, remaining: 3}

	// Write one byte at a time.
	for i, b := range []byte("abcde") {
		n, err := lw.Write([]byte{b})
		if err != nil {
			t.Fatalf("Write %d error: %v", i, err)
		}
		if n != 1 {
			t.Errorf("Write %d returned n=%d, want 1", i, n)
		}
	}

	if buf.String() != "abc" {
		t.Errorf("buffer content = %q, want %q", buf.String(), "abc")
	}
}

// ---------------------------------------------------------------------------
// Result.StdoutStr / StderrStr
// ---------------------------------------------------------------------------

func TestResultStdoutStr(t *testing.T) {
	tests := []struct {
		name   string
		stdout []byte
		want   string
	}{
		{"simple", []byte("hello"), "hello"},
		{"leading space", []byte("  hello"), "hello"},
		{"trailing newline", []byte("hello\n"), "hello"},
		{"both", []byte("  hello world\n\n"), "hello world"},
		{"empty", []byte(""), ""},
		{"only whitespace", []byte("  \n\t "), ""},
		{"tabs and newlines", []byte("\t\nhello\n\t"), "hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Result{Stdout: tt.stdout}
			if got := r.StdoutStr(); got != tt.want {
				t.Errorf("StdoutStr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResultStderrStr(t *testing.T) {
	tests := []struct {
		name   string
		stderr []byte
		want   string
	}{
		{"simple", []byte("error msg"), "error msg"},
		{"trailing newline", []byte("error\n"), "error"},
		{"empty", []byte{}, ""},
		{"surrounding whitespace", []byte("\n  warn  \n"), "warn"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Result{Stderr: tt.stderr}
			if got := r.StderrStr(); got != tt.want {
				t.Errorf("StderrStr() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// envSliceToMap converts a []string of "KEY=VALUE" into a map.
// If a key appears multiple times, the last value wins.
func envSliceToMap(env []string) map[string]string {
	m := make(map[string]string, len(env))
	for _, e := range env {
		k, v, _ := strings.Cut(e, "=")
		m[k] = v
	}
	return m
}
