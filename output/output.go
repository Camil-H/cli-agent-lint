package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-isatty"
)

type Config struct {
	Format  string // "text" or "json"
	NoColor bool
	Quiet   bool
	Stdout  io.Writer
	Stderr  io.Writer
}

type Output struct {
	format  string
	noColor bool
	quiet   bool
	stdout  io.Writer
	stderr  io.Writer
	isTTY   bool
}

func New(cfg Config) *Output {
	stdout := cfg.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := cfg.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	isTTY := false
	if f, ok := stdout.(*os.File); ok {
		isTTY = isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
	}

	noColor := cfg.NoColor
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		noColor = true
	}

	format := cfg.Format
	if format == "" {
		format = "text"
	}

	return &Output{
		format:  format,
		noColor: noColor || !isTTY,
		quiet:   cfg.Quiet,
		stdout:  stdout,
		stderr:  stderr,
		isTTY:   isTTY,
	}
}

func (o *Output) IsJSON() bool          { return o.format == "json" }
func (o *Output) NoColor() bool         { return o.noColor }
func (o *Output) IsQuiet() bool         { return o.quiet }
func (o *Output) DataWriter() io.Writer { return o.stdout }

func (o *Output) Diag(format string, args ...any) {
	if o.quiet {
		return
	}
	fmt.Fprintf(o.stderr, format+"\n", args...)
}

func (o *Output) Error(err error) error {
	if err == nil {
		return nil
	}
	if o.format == "json" {
		obj := jsonError{Error: err.Error()}
		var coded *CodedError
		var cliErr *CLIError
		if errors.As(err, &coded) {
			obj.Code = coded.Code
			obj.Message = coded.Message
		} else if errors.As(err, &cliErr) {
			obj.Message = cliErr.Message
		}
		data, jsonErr := json.Marshal(obj)
		if jsonErr != nil {
			fmt.Fprintf(o.stderr, "Error: %s\n", err.Error())
			return err
		}
		fmt.Fprintln(o.stderr, string(data))
	} else {
		fmt.Fprintf(o.stderr, "Error: %s\n", err.Error())
	}
	return err
}

type jsonError struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

type CLIError struct {
	Err     error
	Message string
}

func (e *CLIError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *CLIError) Unwrap() error { return e.Err }


type CodedError struct {
	Err     error
	Code    string
	Message string
}

func (e *CodedError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *CodedError) Unwrap() error { return e.Err }


type CheckFailedError struct{}

func (e *CheckFailedError) Error() string {
	return "one or more checks with severity=fail did not pass"
}
