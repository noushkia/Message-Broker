package main

import (
	"Message-Broker/internal/app/cmd"
	"os"
)

const (
	exitFailure = 1
)

func main() {
	root := cmd.NewRootCommand()

	if root != nil {
		if err := root.Execute(); err != nil {
			os.Exit(exitFailure)
		}
	}
}
