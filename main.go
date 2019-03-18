package main

import (
	"github.com/gcpug/zagane/passes/unstopiter"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(unstopiter.Analyzer) }