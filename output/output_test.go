package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"testing"
)

// ---------- New() defaults ----------

func TestNew_NilWriters_DefaultToOsStdout(t *testing.T) {
	o := New(Config{})
	if o.DataWriter() != os.Stdout {
		t.Error("expected DataWriter to be os.Stdout when Config.Stdout is nil")
	}
}

func TestNew_CustomWriters(t *testing.T) {
	var stdout bytes.Buffer
	o := New(Config{Stdout: &stdout})
	if o.DataWriter() != &stdout {
		t.Error("expected DataWriter to be the injected stdout buffer")
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

func TestDataWriter_ReturnsStdout(t *testing.T) {
	var buf bytes.Buffer
	o := New(Config{Stdout: &buf})
	if o.DataWriter() != &buf {
		t.Error("DataWriter() should return the configured stdout writer")
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

func TestError_ReturnsSameError(t *testing.T) {
	var stderr bytes.Buffer
	o := New(Config{Format: "text", Stderr: &stderr})
	original := errors.New("test error")
	returned := o.Error(original)
	if returned != original {
		t.Error("Error() should return the exact same error passed in")
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

func TestDiag_NoArgs(t *testing.T) {
	var stderr bytes.Buffer
	o := New(Config{Stderr: &stderr})
	o.Diag("plain message with no formatting")
	got := stderr.String()
	if got != "plain message with no formatting\n" {
		t.Errorf("expected plain message, got %q", got)
	}
}

func TestNew_NonJSONFormat(t *testing.T) {
	o := New(Config{Format: "yaml"})
	if o.IsJSON() {
		t.Error("IsJSON() should be false for non-json format")
	}
}
