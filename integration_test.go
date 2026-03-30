package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// binaryPath holds the path to the built cli-agent-lint binary.
var binaryPath string

func TestMain(m *testing.M) {
	// Build the binary once for all integration tests.
	binaryPath = "./cli-agent-lint"
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build binary: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	// Clean up the built binary.
	os.Remove(binaryPath)

	os.Exit(code)
}

// runBinary executes the built binary with the given arguments and returns
// stdout, stderr, and any error (including exit code info).
func runBinary(args ...string) (stdout string, stderr string, err error) {
	cmd := exec.Command(binaryPath, args...)
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

// parseJSON parses a JSON string into a map[string]interface{}.
func parseJSON(t *testing.T, s string) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\nraw output:\n%s", err, s)
	}
	return result
}

// parseJSONArray parses a JSON string into a []interface{}.
func parseJSONArray(t *testing.T, s string) []interface{} {
	t.Helper()
	var result []interface{}
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		t.Fatalf("failed to parse JSON array: %v\nraw output:\n%s", err, s)
	}
	return result
}

// getCheckByID finds a check result by its ID in the checks array of a parsed report.
func getCheckByID(t *testing.T, report map[string]interface{}, checkID string) map[string]interface{} {
	t.Helper()
	checksRaw, ok := report["checks"]
	if !ok {
		t.Fatal("report JSON missing 'checks' field")
	}
	checksArr, ok := checksRaw.([]interface{})
	if !ok {
		t.Fatal("'checks' field is not an array")
	}
	for _, c := range checksArr {
		cm, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		if cm["id"] == checkID {
			return cm
		}
	}
	return nil
}

func TestSelfAudit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	stdout, stderr, err := runBinary("check", binaryPath, "--output", "json")
	if err != nil {
		t.Logf("stderr: %s", stderr)
		// The self-audit might exit 1 if there are critical failures;
		// we still want to parse and check the output.
		if stdout == "" {
			t.Fatalf("self-audit produced no stdout; error: %v", err)
		}
	}

	report := parseJSON(t, stdout)

	// Verify the score section exists.
	scoreRaw, ok := report["score"]
	if !ok {
		t.Fatal("report JSON missing 'score' field")
	}
	score, ok := scoreRaw.(map[string]interface{})
	if !ok {
		t.Fatal("'score' field is not an object")
	}

	grade, ok := score["grade"].(string)
	if !ok {
		t.Fatal("score.grade is not a string")
	}

	if grade != "A" {
		t.Errorf("expected self-audit grade 'A', got %q", grade)
	}

	// Verify no critical failures: no checks with severity=fail and status=fail.
	checksRaw, ok := report["checks"]
	if !ok {
		t.Fatal("report JSON missing 'checks' field")
	}
	checksArr, ok := checksRaw.([]interface{})
	if !ok {
		t.Fatal("'checks' is not an array")
	}
	for _, c := range checksArr {
		cm, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		severity, _ := cm["severity"].(string)
		status, _ := cm["status"].(string)
		if severity == "fail" && status == "fail" {
			t.Errorf("critical failure in self-audit: check %v (%v)", cm["id"], cm["name"])
		}
	}
}

func TestGoodCLI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	stdout, stderr, err := runBinary("check", "./testdata/fake-cli/good-cli.sh", "--output", "json")
	if err != nil {
		t.Logf("stderr: %s", stderr)
		if stdout == "" {
			t.Fatalf("good-cli check produced no stdout; error: %v", err)
		}
	}

	report := parseJSON(t, stdout)

	scoreRaw, ok := report["score"]
	if !ok {
		t.Fatal("report JSON missing 'score' field")
	}
	score, ok := scoreRaw.(map[string]interface{})
	if !ok {
		t.Fatal("'score' field is not an object")
	}

	grade, ok := score["grade"].(string)
	if !ok {
		t.Fatal("score.grade is not a string")
	}

	// Good CLI should be at least a B, ideally an A.
	if grade != "A" && grade != "B" {
		t.Errorf("expected good-cli grade 'A' or 'B', got %q", grade)
	}
}

func TestBadCLI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	stdout, stderr, err := runBinary("check", "./testdata/fake-cli/bad-cli.sh", "--output", "json")

	// bad-cli should cause exit code 1 (critical failures).
	if err == nil {
		t.Error("expected non-zero exit code for bad-cli, got 0")
	} else {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("expected *exec.ExitError, got %T: %v", err, err)
		}
		if exitErr.ExitCode() != 1 {
			t.Errorf("expected exit code 1 for bad-cli, got %d", exitErr.ExitCode())
		}
	}

	if stdout == "" {
		t.Logf("stderr: %s", stderr)
		t.Fatal("bad-cli check produced no stdout")
	}

	report := parseJSON(t, stdout)

	scoreRaw, ok := report["score"]
	if !ok {
		t.Fatal("report JSON missing 'score' field")
	}
	score, ok := scoreRaw.(map[string]interface{})
	if !ok {
		t.Fatal("'score' field is not an object")
	}

	grade, ok := score["grade"].(string)
	if !ok {
		t.Fatal("score.grade is not a string")
	}

	// Bad CLI should get F or D.
	if grade != "F" && grade != "D" {
		t.Errorf("expected bad-cli grade 'F' or 'D', got %q", grade)
	}
}

func TestNoJSONCLI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	stdout, stderr, err := runBinary("check", "./testdata/fake-cli/no-json-cli.sh", "--output", "json")
	if err != nil {
		t.Logf("stderr: %s", stderr)
		if stdout == "" {
			t.Fatalf("no-json-cli check produced no stdout; error: %v", err)
		}
	}

	report := parseJSON(t, stdout)

	// Verify SO-1 (JSON output support) fails.
	so1 := getCheckByID(t, report, "SO-1")
	if so1 == nil {
		t.Fatal("SO-1 check not found in report")
	}

	status, _ := so1["status"].(string)
	if status != "fail" {
		t.Errorf("expected SO-1 status 'fail' for no-json-cli, got %q", status)
	}
}

func TestNoisyCLI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	stdout, stderr, err := runBinary("check", "./testdata/fake-cli/noisy-cli.sh", "--output", "json")
	if err != nil {
		t.Logf("stderr: %s", stderr)
		if stdout == "" {
			t.Fatalf("noisy-cli check produced no stdout; error: %v", err)
		}
	}

	report := parseJSON(t, stdout)

	// Verify TH-1 (no ANSI in piped output) fails.
	th1 := getCheckByID(t, report, "TH-1")
	if th1 == nil {
		t.Fatal("TH-1 check not found in report")
	}

	status, _ := th1["status"].(string)
	if status != "fail" && status != "warn" {
		t.Errorf("expected TH-1 status 'fail' or 'warn' for noisy-cli, got %q", status)
	}
}

func TestVersionOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	stdout, stderr, err := runBinary("--version")
	if err != nil {
		t.Fatalf("--version failed: %v\nstderr: %s", err, stderr)
	}

	version := strings.TrimSpace(stdout)
	if version != "0.1.0" {
		t.Errorf("expected version '0.1.0', got %q", version)
	}

	// Verify it is a clean semver (no extra text).
	if strings.Contains(version, " ") {
		t.Errorf("version output contains spaces (not clean semver): %q", version)
	}
}

func TestJSONVersionOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	stdout, stderr, err := runBinary("--version", "-o", "json")
	if err != nil {
		t.Fatalf("--version -o json failed: %v\nstderr: %s", err, stderr)
	}

	result := parseJSON(t, stdout)

	version, ok := result["version"].(string)
	if !ok {
		t.Fatal("JSON version output missing 'version' field")
	}

	if version != "0.1.0" {
		t.Errorf("expected JSON version '0.1.0', got %q", version)
	}
}

func TestChecksListJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	stdout, stderr, err := runBinary("checks", "--output", "json")
	if err != nil {
		t.Fatalf("checks --output json failed: %v\nstderr: %s", err, stderr)
	}

	items := parseJSONArray(t, stdout)

	if len(items) == 0 {
		t.Fatal("checks list returned empty array")
	}

	// Verify each item has expected fields.
	for i, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			t.Fatalf("checks[%d] is not an object", i)
		}

		for _, field := range []string{"id", "name", "category", "severity", "method"} {
			if _, ok := m[field]; !ok {
				t.Errorf("checks[%d] missing field %q", i, field)
			}
		}
	}
}

func TestNoANSIWhenPiped(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Running via os/exec means stdout is a pipe, not a TTY.
	// The tool should not emit ANSI escape codes when piped.
	stdout, stderr, err := runBinary("check", binaryPath)
	if err != nil {
		t.Logf("stderr: %s", stderr)
		// Non-zero exit is OK here; we only care about ANSI in output.
		if stdout == "" {
			t.Fatalf("check produced no stdout; error: %v", err)
		}
	}

	if strings.Contains(stdout, "\x1b[") {
		t.Error("stdout contains ANSI escape sequences when piped (should not)")
	}
}
