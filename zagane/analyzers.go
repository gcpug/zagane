package zagane

import (
	"github.com/gcpug/zagane/passes/unclosetx"
	"github.com/gcpug/zagane/passes/unstopiter"
	"github.com/gcpug/zagane/passes/wraperr"
	"golang.org/x/tools/go/analysis"
)

// Analyzers returns analyzers of zagane.
func Analyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		unstopiter.Analyzer,
		unclosetx.Analyzer,
		wraperr.Analyzer,
	}
}
