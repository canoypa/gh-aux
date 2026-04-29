package main

import (
	"os"

	"github.com/canoypa/gh-aux/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
