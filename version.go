package main

import (
	_ "embed"
	"fmt"
	"os"
	"runtime"
	"strings"
)

//go:embed version.txt
var version string

// print version of zagane
func printVersion() bool {
	if len(os.Args) != 2 {
		return false
	}

	switch os.Args[1] {
	case "-v", "--version":
		fmt.Printf("zagane %s (%s)\n", strings.TrimSpace(version), runtime.Version())
		return true
	}

	return false
}
