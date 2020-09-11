package main

import (
	"fmt"
	"os"
	"runtime"
)

// version of zagane
const version = "v0.5.2"

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
