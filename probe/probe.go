package probe

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

var bufPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

// 10 MiB cap prevents unbounded memory growth from misbehaving target CLIs.
const maxOutputBytes = 10 * 1024 * 1024

var sensitiveEnvPrefixes = []string{
	"AWS_SECRET",
	"AWS_SESSION_TOKEN",
	"GITHUB_TOKEN",
	"GH_TOKEN",
	"GITLAB_TOKEN",
	"DOCKER_PASSWORD",
	"DOCKER_AUTH",
	"NPM_TOKEN",
	"NUGET_API_KEY",
	"PYPI_TOKEN",
	"RUBYGEMS_API_KEY",
	"CODECOV_TOKEN",
	"SONAR_TOKEN",
	"SNYK_TOKEN",
	"SENTRY_AUTH_TOKEN",
	"SLACK_TOKEN",
	"SLACK_WEBHOOK",
	"TWILIO_AUTH_TOKEN",
	"SENDGRID_API_KEY",
	"STRIPE_SECRET",
	"DATABASE_PASSWORD",
	"DB_PASSWORD",
	"REDIS_PASSWORD",
}

var sensitiveEnvSubstrings = []string{
	"_SECRET",
	"_PASSWORD",
	"_PRIVATE_KEY",
	"_CREDENTIALS",
	"_API_KEY",
	"_TOKEN",
}

func isSensitiveEnvVar(name string) bool {
	upper := strings.ToUpper(name)
	for _, prefix := range sensitiveEnvPrefixes {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}
	for _, substr := range sensitiveEnvSubstrings {
		if strings.Contains(upper, substr) {
			return true
		}
	}
	return false
}

type Runner interface {
	Run(ctx context.Context, opts Opts) (*Result, error)
	RunHelp(ctx context.Context, args ...string) (*Result, error)
}

var _ Runner = (*Prober)(nil)

type Prober struct {
	targetPath     string
	defaultTimeout time.Duration
	devnull        *os.File
	baseEnv        []string // sanitized base environment, cached at construction
}

type Opts struct {
	Args        []string
	Env         []string // additional env vars (KEY=VALUE); merged with defaults
	Timeout     time.Duration
	Stdin       *os.File
	ForceNonTTY bool // reserved for future use
}

type Result struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
	TimedOut bool
	Duration time.Duration
}

func (r *Result) StdoutStr() string { return strings.TrimSpace(string(r.Stdout)) }
func (r *Result) StderrStr() string { return strings.TrimSpace(string(r.Stderr)) }

func New(targetPath string, defaultTimeout time.Duration) (*Prober, error) {
	resolved, err := exec.LookPath(targetPath)
	if err != nil {
		return nil, fmt.Errorf("target not found or not executable: %s: %w", targetPath, err)
	}
	if defaultTimeout <= 0 {
		defaultTimeout = 5 * time.Second
	}

	devnull, err := os.Open(os.DevNull)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", os.DevNull, err)
	}

	baseEnv := buildBaseEnv()

	return &Prober{
		targetPath:     resolved,
		defaultTimeout: defaultTimeout,
		devnull:        devnull,
		baseEnv:        baseEnv,
	}, nil
}

func (p *Prober) Close() error {
	if p.devnull != nil {
		return p.devnull.Close()
	}
	return nil
}

func (p *Prober) TargetPath() string { return p.targetPath }

func (p *Prober) Run(ctx context.Context, opts Opts) (*Result, error) {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = p.defaultTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, p.targetPath, opts.Args...)

	// Set process group so we can kill the whole group on timeout.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if opts.Stdin != nil {
		cmd.Stdin = opts.Stdin
	} else {
		cmd.Stdin = p.devnull
	}

	stdoutBuf := bufPool.Get().(*bytes.Buffer)
	stderrBuf := bufPool.Get().(*bytes.Buffer)
	stdoutBuf.Reset()
	stderrBuf.Reset()

	cmd.Stdout = &limitedWriter{buf: stdoutBuf, remaining: maxOutputBytes}
	cmd.Stderr = &limitedWriter{buf: stderrBuf, remaining: maxOutputBytes}
	cmd.Env = p.buildEnv(opts.Env)

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	// Copy buffer contents before returning them to the pool.
	result := &Result{
		Stdout:   bytes.Clone(stdoutBuf.Bytes()),
		Stderr:   bytes.Clone(stderrBuf.Bytes()),
		Duration: duration,
	}
	bufPool.Put(stdoutBuf)
	bufPool.Put(stderrBuf)

	if ctx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		if cmd.Process != nil {
			if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil && err != syscall.ESRCH {
				return result, fmt.Errorf("kill process group %d: %w", cmd.Process.Pid, err)
			}
		}
		return result, nil
	}

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			result.ExitCode = exitErr.ExitCode()
			return result, nil
		}
		return result, fmt.Errorf("exec %s: %w", p.targetPath, err)
	}

	return result, nil
}

func (p *Prober) RunHelp(ctx context.Context, args ...string) (*Result, error) {
	fullArgs := make([]string, 0, len(args)+1)
	fullArgs = append(fullArgs, args...)
	fullArgs = append(fullArgs, "--help")
	return p.Run(ctx, Opts{Args: fullArgs})
}

var envDefaults = map[string]string{
	"PAGER":     "cat",
	"GIT_PAGER": "cat",
	"TERM":      "dumb",
}

func buildBaseEnv() []string {
	base := os.Environ()
	result := make([]string, 0, len(base)+len(envDefaults))
	seen := make(map[string]bool, len(base))

	for _, e := range base {
		k, _, _ := strings.Cut(e, "=")
		if isSensitiveEnvVar(k) {
			continue
		}
		if v, ok := envDefaults[k]; ok {
			result = append(result, k+"="+v)
		} else {
			result = append(result, e)
		}
		seen[k] = true
	}

	for k, v := range envDefaults {
		if !seen[k] {
			result = append(result, k+"="+v)
		}
	}

	return result
}

func (p *Prober) buildEnv(extra []string) []string {
	if len(extra) == 0 {
		return p.baseEnv
	}

	overrides := make(map[string]string, len(extra))
	for _, e := range extra {
		if k, v, ok := strings.Cut(e, "="); ok {
			if isSensitiveEnvVar(k) {
				continue
			}
			overrides[k] = v
		}
	}

	result := make([]string, 0, len(p.baseEnv)+len(extra))
	for _, e := range p.baseEnv {
		k, _, _ := strings.Cut(e, "=")
		if v, ok := overrides[k]; ok {
			result = append(result, k+"="+v)
			delete(overrides, k)
		} else {
			result = append(result, e)
		}
	}

	for k, v := range overrides {
		result = append(result, k+"="+v)
	}

	return result
}

type limitedWriter struct {
	buf       *bytes.Buffer
	remaining int
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if w.remaining <= 0 {
		return len(p), nil // discard but report success to avoid breaking the subprocess
	}
	n := len(p)
	if n > w.remaining {
		n = w.remaining
	}
	written, err := w.buf.Write(p[:n])
	w.remaining -= written
	if err != nil {
		return written, err
	}
	return len(p), nil // report full write to subprocess even if truncated
}

var _ io.Writer = (*limitedWriter)(nil)
