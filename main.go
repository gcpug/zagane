// +build !go1.12

package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/gcpug/zagane/passes/unstopiter"
	"golang.org/x/tools/go/analysis/singlechecker"
)

// version of zagane
const version = "v0.4.0"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "version" {
		fmt.Printf("zagane %s (%s)\n", version, runtime.Version())
		return
	}
	singlechecker.Main(unstopiter.Analyzer)
}
