package unstopiter

import (
	"go/ast"
	"go/types"
	"strings"

	"github.com/gcpug/zagane/zaganeutils"
	"github.com/gostaticanalysis/analysisutil"
	"github.com/gostaticanalysis/comment"
	"github.com/gostaticanalysis/comment/passes/commentmap"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
)

var stopMethods = "Stop,Do"

var Analyzer = &analysis.Analyzer{
	Name: "unstopiter",
	Doc:  Doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		buildssa.Analyzer,
		commentmap.Analyzer,
	},
}

const Doc = "unstopiter finds iterators which did not stop"

func init() {
	Analyzer.Flags.StringVar(&stopMethods, "methods", stopMethods, "stop methods")
}

func run(pass *analysis.Pass) (interface{}, error) {
	funcs := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA).SrcFuncs
	cmaps := pass.ResultOf[commentmap.Analyzer].(comment.Maps)

	iterTyp := zaganeutils.TypeOf(pass, "*RowIterator")
	if iterTyp == nil {
		// skip checking
		return nil, nil
	}

	var methods []*types.Func
	for _, s := range strings.Split(stopMethods, ",") {
		if m := analysisutil.MethodOf(iterTyp, s); m != nil {
			methods = append(methods, m)
		}
	}

	skipFile := map[*ast.File]bool{}
	for _, f := range funcs {
		if zaganeutils.Unimported(pass, f, skipFile) {
			// skip this
			continue
		}

		for _, b := range f.Blocks {
			for i := range b.Instrs {
				pos := b.Instrs[i].Pos()

				// skip
				if cmaps.IgnorePos(pos, "zagane") ||
					cmaps.IgnorePos(pos, "unstopiter") {
					continue
				}

				called, ok := analysisutil.CalledFrom(b, i, iterTyp, methods...)
				if ok && !called {
					pass.Reportf(pos, "iterator must be stopped")
				}
			}
		}
	}

	return nil, nil
}
