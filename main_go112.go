// +build go1.12

package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/gcpug/zagane/zagane"
	"golang.org/x/tools/go/analysis/unitchecker"
)

// version of zagane
const version = "v0.4.0"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "version" {
		fmt.Printf("zagane %s (%s)\n", version, runtime.Version())
		return
	}
	unitchecker.Main(zagane.Analyzers()...)
}
