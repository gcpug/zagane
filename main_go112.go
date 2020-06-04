// +build go1.12

package main

import (
	"github.com/gcpug/zagane/zagane"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() {
	if printVersion() {
		return
	}
	unitchecker.Main(zagane.Analyzers()...)
}
