package main

import (
	"os"

	"github.com/fuckssh/fuckssh/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
