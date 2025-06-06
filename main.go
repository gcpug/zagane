package main

import (
	"golang.org/x/tools/go/analysis/unitchecker"

	"github.com/gcpug/zagane/zagane"
)

func main() {
	if printVersion() {
		return
	}
	unitchecker.Main(zagane.Analyzers()...)
}
