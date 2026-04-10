package main

import (
	"os"

	"github.com/Camil-H/cli-agent-lint/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
