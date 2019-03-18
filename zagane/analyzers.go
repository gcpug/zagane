package zagane

import (
	"github.com/gcpug/zagane/passes/unstopiter"
	"golang.org/x/tools/go/analysis"
)

// Analyzers returns analyzers of zagane.
func Analyzers() []*analysis.Analyzer {
	return []*Analyzer{
		unstopiter.Analyzer,
	}
}
