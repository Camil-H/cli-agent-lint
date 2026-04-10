package checks

import (
	"testing"

	"github.com/Camil-H/cli-agent-lint/discovery"
)

func makeTree(root *discovery.Command) *discovery.CommandTree {
	return &discovery.CommandTree{Root: root, TargetPath: "/usr/bin/test-cli"}
}

func makeIndex(root *discovery.Command) *discovery.CommandIndex {
	return discovery.NewIndex(root)
}

func makeInput(root *discovery.Command) *Input {
	tree := makeTree(root)
	var idx *discovery.CommandIndex
	if root != nil {
		idx = discovery.NewIndex(root)
	}
	return &Input{Tree: tree, Index: idx}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello..."},
		{"", 5, ""},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc..."},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.maxLen)
		if got != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.expected)
		}
	}
}
