package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

// ---------- New() defaults ----------

func TestNew_DefaultFormat(t *testing.T) {
	o := New(Config{})
	if o.Format() != "text" {
		t.Errorf("expected default format \"text\", got %q", o.Format())
	}
}

func TestNew_ExplicitFormat(t *testing.T) {
	o := New(Config{Format: "json"})
	if o.Format() != "json" {
		t.Errorf("expected format \"json\", got %q", o.Format())
	}
}

func TestNew_NilWriters_DefaultToOsStdoutStderr(t *testing.T) {
	o := New(Config{})
	if o.DataWriter() != os.Stdout {
		t.Error("expected DataWriter to be os.Stdout when Config.Stdout is nil")
	}
	if o.DiagWriter() != os.Stderr {
		t.Error("expected DiagWriter to be os.Stderr when Config.Stderr is nil")
	}
}

func TestNew_CustomWriters(t *testing.T) {
	var stdout, stderr bytes.Buffer
	o := New(Config{Stdout: &stdout, Stderr: &stderr})
	if o.DataWriter() != &stdout {
		t.Error("expected DataWriter to be the injected stdout buffer")
	}
	if o.DiagWriter() != &stderr {
		t.Error("expected DiagWriter to be the injected stderr buffer")
	}
}

func TestNew_IsTTY_FalseForBuffer(t *testing.T) {
	var buf bytes.Buffer
	o := New(Config{Stdout: &buf})
	if o.IsTTY() {
		t.Error("expected IsTTY() to be false for a bytes.Buffer (non-terminal)")
	}
}

func TestNew_NoColor_TrueForNonTTY(t *testing.T) {
	var buf bytes.Buffer
	o := New(Config{Stdout: &buf})
	// A non-TTY stdout should force noColor to true regardless of Config.NoColor.
	if !o.NoColor() {
		t.Error("expected NoColor() to be true when stdout is not a TTY")
	}
}

func TestNew_NoColor_ExplicitTrue(t *testing.T) {
	var buf bytes.Buffer
	o := New(Config{Stdout: &buf, NoColor: true})
	if !o.NoColor() {
		t.Error("expected NoColor() to be true when explicitly set")
	}
}

func TestNew_NoColor_EnvironmentVariable(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	var buf bytes.Buffer
	o := New(Config{Stdout: &buf})
	if !o.NoColor() {
		t.Error("expected NoColor() to be true when NO_COLOR env var is set")
	}
}

func TestNew_Quiet(t *testing.T) {
	o := New(Config{Quiet: true})
	if !o.IsQuiet() {
		t.Error("expected IsQuiet() to be true when Config.Quiet is true")
	}
}

func TestNew_QuietDefault(t *testing.T) {
	o := New(Config{})
	if o.IsQuiet() {
		t.Error("expected IsQuiet() to be false by default")
	}
}

// ---------- Accessor methods ----------

func TestIsJSON(t *testing.T) {
	tests := []struct {
		format string
		want   bool
	}{
		{"json", true},
		{"text", false},
		{"", false}, // will be defaulted to "text" by New
	}
	for _, tt := range tests {
		o := New(Config{Format: tt.format})
		if got := o.IsJSON(); got != tt.want {
			t.Errorf("IsJSON() with format=%q: got %v, want %v", tt.format, got, tt.want)
		}
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"json", "json"},
		{"text", "text"},
		{"", "text"},
	}
	for _, tt := range tests {
		o := New(Config{Format: tt.input})
		if got := o.Format(); got != tt.want {
			t.Errorf("Format() with input=%q: got %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDataWriter_ReturnsStdout(t *testing.T) {
	var buf bytes.Buffer
	o := New(Config{Stdout: &buf})
	if o.DataWriter() != &buf {
		t.Error("DataWriter() should return the configured stdout writer")
	}
}

func TestDiagWriter_ReturnsStderr(t *testing.T) {
	var buf bytes.Buffer
	o := New(Config{Stderr: &buf})
	if o.DiagWriter() != &buf {
		t.Error("DiagWriter() should return the configured stderr writer")
	}
}

// ---------- Diag() ----------

func TestDiag_WritesToStderr(t *testing.T) {
	var stderr bytes.Buffer
	o := New(Config{Stderr: &stderr})
	o.Diag("hello %s", "world")
	got := stderr.String()
	if got != "hello world\n" {
		t.Errorf("expected \"hello world\\n\", got %q", got)
	}
}

func TestDiag_SuppressedWhenQuiet(t *testing.T) {
	var stderr bytes.Buffer
	o := New(Config{Stderr: &stderr, Quiet: true})
	o.Diag("should not appear")
	if stderr.Len() != 0 {
		t.Errorf("expected no output when quiet, got %q", stderr.String())
	}
}

func TestDiag_MultipleMessages(t *testing.T) {
	var stderr bytes.Buffer
	o := New(Config{Stderr: &stderr})
	o.Diag("line %d", 1)
	o.Diag("line %d", 2)
	got := stderr.String()
	if got != "line 1\nline 2\n" {
		t.Errorf("expected two lines, got %q", got)
	}
}

// ---------- Error() ----------

func TestError_NilReturnsNil(t *testing.T) {
	var stderr bytes.Buffer
	o := New(Config{Stderr: &stderr})
	result := o.Error(nil)
	if result != nil {
		t.Errorf("expected nil return for nil error, got %v", result)
	}
	if stderr.Len() != 0 {
		t.Errorf("expected no output for nil error, got %q", stderr.String())
	}
}

func TestError_TextFormat(t *testing.T) {
	var stderr bytes.Buffer
	o := New(Config{Format: "text", Stderr: &stderr})
	err := errors.New("something broke")
	returned := o.Error(err)
	if returned != err {
		t.Error("Error() should return the original error")
	}
	got := stderr.String()
	if got != "Error: something broke\n" {
		t.Errorf("expected \"Error: something broke\\n\", got %q", got)
	}
}

func TestError_JSONFormat_PlainError(t *testing.T) {
	var stderr bytes.Buffer
	o := New(Config{Format: "json", Stderr: &stderr})
	err := errors.New("plain error")
	returned := o.Error(err)
	if returned != err {
		t.Error("Error() should return the original error")
	}

	var obj jsonError
	if jsonErr := json.Unmarshal(stderr.Bytes(), &obj); jsonErr != nil {
		t.Fatalf("failed to parse JSON output: %v", jsonErr)
	}
	if obj.Error != "plain error" {
		t.Errorf("expected error=\"plain error\", got %q", obj.Error)
	}
	if obj.Code != "" {
		t.Errorf("expected empty code for plain error, got %q", obj.Code)
	}
	if obj.Message != "" {
		t.Errorf("expected empty message for plain error, got %q", obj.Message)
	}
}

func TestError_JSONFormat_CodedError(t *testing.T) {
	var stderr bytes.Buffer
	o := New(Config{Format: "json", Stderr: &stderr})
	ce := NewCodedError("ERR_TEST", "test message %d", 42)
	o.Error(ce)

	var obj jsonError
	if err := json.Unmarshal(stderr.Bytes(), &obj); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}
	if obj.Code != "ERR_TEST" {
		t.Errorf("expected code=\"ERR_TEST\", got %q", obj.Code)
	}
	if obj.Message != "test message 42" {
		t.Errorf("expected message=\"test message 42\", got %q", obj.Message)
	}
	if obj.Error != ce.Error() {
		t.Errorf("expected error=%q, got %q", ce.Error(), obj.Error)
	}
}

func TestError_JSONFormat_CLIError(t *testing.T) {
	var stderr bytes.Buffer
	o := New(Config{Format: "json", Stderr: &stderr})
	inner := errors.New("inner cause")
	cliErr := NewCLIError(inner, "user-facing message")
	o.Error(cliErr)

	var obj jsonError
	if err := json.Unmarshal(stderr.Bytes(), &obj); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}
	if obj.Error != "inner cause" {
		t.Errorf("expected error=\"inner cause\", got %q", obj.Error)
	}
	if obj.Message != "user-facing message" {
		t.Errorf("expected message=\"user-facing message\", got %q", obj.Message)
	}
	if obj.Code != "" {
		t.Errorf("expected empty code for CLIError, got %q", obj.Code)
	}
}

func TestError_ReturnsSameError(t *testing.T) {
	var stderr bytes.Buffer
	o := New(Config{Format: "text", Stderr: &stderr})
	original := errors.New("test error")
	returned := o.Error(original)
	if returned != original {
		t.Error("Error() should return the exact same error passed in")
	}
}

// ---------- NewCLIError ----------

func TestNewCLIError_ErrorReturnsWrapped(t *testing.T) {
	inner := errors.New("root cause")
	cliErr := NewCLIError(inner, "display: %s", "details")
	if cliErr.Error() != "root cause" {
		t.Errorf("expected Error() to return inner error text, got %q", cliErr.Error())
	}
}

func TestNewCLIError_ErrorFallsBackToMessage(t *testing.T) {
	cliErr := NewCLIError(nil, "message only")
	if cliErr.Error() != "message only" {
		t.Errorf("expected Error() to return message when Err is nil, got %q", cliErr.Error())
	}
}

func TestNewCLIError_Unwrap(t *testing.T) {
	inner := errors.New("root cause")
	cliErr := NewCLIError(inner, "display message")
	if cliErr.Unwrap() != inner {
		t.Error("Unwrap() should return the inner error")
	}
}

func TestNewCLIError_UnwrapNil(t *testing.T) {
	cliErr := NewCLIError(nil, "message only")
	if cliErr.Unwrap() != nil {
		t.Error("Unwrap() should return nil when inner error is nil")
	}
}

func TestNewCLIError_MessageFormatting(t *testing.T) {
	inner := errors.New("x")
	cliErr := NewCLIError(inner, "count: %d, name: %s", 3, "alpha")
	if cliErr.Message != "count: 3, name: alpha" {
		t.Errorf("expected formatted message, got %q", cliErr.Message)
	}
}

func TestNewCLIError_ErrorsIs(t *testing.T) {
	inner := errors.New("sentinel")
	cliErr := NewCLIError(inner, "wrapped")
	if !errors.Is(cliErr, inner) {
		t.Error("errors.Is should find the inner error through Unwrap")
	}
}

// ---------- NewCodedError ----------

func TestNewCodedError_Fields(t *testing.T) {
	ce := NewCodedError("E_PARSE", "parse failed at %s", "line 5")
	if ce.Code != "E_PARSE" {
		t.Errorf("expected Code=\"E_PARSE\", got %q", ce.Code)
	}
	if ce.Message != "parse failed at line 5" {
		t.Errorf("expected Message=\"parse failed at line 5\", got %q", ce.Message)
	}
	if ce.Err == nil {
		t.Fatal("expected Err to be set")
	}
	if !strings.Contains(ce.Err.Error(), "E_PARSE") {
		t.Errorf("expected Err to contain code, got %q", ce.Err.Error())
	}
	if !strings.Contains(ce.Err.Error(), "parse failed at line 5") {
		t.Errorf("expected Err to contain message, got %q", ce.Err.Error())
	}
}

func TestNewCodedError_ErrorReturnsWrapped(t *testing.T) {
	ce := NewCodedError("E_NET", "connection timeout")
	// Error() should delegate to Err.Error()
	if ce.Error() != ce.Err.Error() {
		t.Errorf("expected Error() = Err.Error(), got %q vs %q", ce.Error(), ce.Err.Error())
	}
}

func TestNewCodedError_ErrorFallsBackToMessage(t *testing.T) {
	ce := &CodedError{Err: nil, Code: "E_X", Message: "fallback"}
	if ce.Error() != "fallback" {
		t.Errorf("expected Error() to return Message when Err is nil, got %q", ce.Error())
	}
}

func TestNewCodedError_Unwrap(t *testing.T) {
	ce := NewCodedError("E_IO", "disk full")
	unwrapped := ce.Unwrap()
	if unwrapped != ce.Err {
		t.Error("Unwrap() should return the inner Err")
	}
}

// ---------- CheckFailedError ----------

func TestCheckFailedError_Message(t *testing.T) {
	e := &CheckFailedError{}
	want := "one or more checks with severity=fail did not pass"
	if e.Error() != want {
		t.Errorf("expected %q, got %q", want, e.Error())
	}
}

func TestCheckFailedError_ImplementsErrorInterface(t *testing.T) {
	var _ error = &CheckFailedError{}
}

// ---------- Edge cases ----------

func TestError_TextFormat_CodedError(t *testing.T) {
	var stderr bytes.Buffer
	o := New(Config{Format: "text", Stderr: &stderr})
	ce := NewCodedError("E_AUTH", "not authorized")
	o.Error(ce)
	got := stderr.String()
	expected := "Error: " + ce.Error() + "\n"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestError_TextFormat_CLIError(t *testing.T) {
	var stderr bytes.Buffer
	o := New(Config{Format: "text", Stderr: &stderr})
	inner := errors.New("underlying problem")
	cliErr := NewCLIError(inner, "user message")
	o.Error(cliErr)
	got := stderr.String()
	if got != "Error: underlying problem\n" {
		t.Errorf("expected \"Error: underlying problem\\n\", got %q", got)
	}
}

func TestDiag_NoArgs(t *testing.T) {
	var stderr bytes.Buffer
	o := New(Config{Stderr: &stderr})
	o.Diag("plain message with no formatting")
	got := stderr.String()
	if got != "plain message with no formatting\n" {
		t.Errorf("expected plain message, got %q", got)
	}
}

func TestNew_FormatPreservedAsIs(t *testing.T) {
	// Arbitrary format string is preserved (no validation).
	o := New(Config{Format: "yaml"})
	if o.Format() != "yaml" {
		t.Errorf("expected format \"yaml\", got %q", o.Format())
	}
	if o.IsJSON() {
		t.Error("IsJSON() should be false for non-json format")
	}
}
