package main

import (
	"os"

	"github.com/mona-actions/gh-bbs-analyzer/cmd"
)

func main() {
	rootCmd := cmd.NewRootCommand()
	
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
