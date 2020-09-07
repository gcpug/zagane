// This file can build as a plugin for golangci-lint by below command.
//    go build -buildmode=plugin -o unclosetx.so github.com/gcpug/zagane/passes/unclosetx/plugin
// See: https://golangci-lint.run/contributing/new-linters/#how-to-add-a-private-linter-to-golangci-lint

package main

import (
	"github.com/gcpug/zagane/passes/unclosetx"
	"golang.org/x/tools/go/analysis"
)

// AnalyzerPlugin provides analyzers as a plugin.
// It follows golangci-lint style plugin.
var AnalyzerPlugin analyzerPlugin

type analyzerPlugin struct{}

func (analyzerPlugin) GetAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		unclosetx.Analyzer,
	}
}
