package main

import (
	"os"

	"github.com/cli-agent-lint/cli-agent-lint/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
