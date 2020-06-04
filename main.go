// +build !go1.12

package main

import (
	"github.com/gcpug/zagane/passes/unstopiter"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	if printVersion() {
		return
	}
	singlechecker.Main(unstopiter.Analyzer)
}
