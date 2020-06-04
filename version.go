package main

import (
	"fmt"
	"os"
	"runtime"
)

// version of zagane
const version = "v0.4.0"

func printVersion() bool {
	if len(os.Args) != 2 {
		return false
	}

	switch os.Args[1] {
	case "-v", "--version":
		fmt.Printf("zagane %s (%s)\n", version, runtime.Version())
		return true
	}

	return false
}
