package unstopiter

import (
	"go/ast"
	"go/token"
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
		instrs := analysisutil.NotCalledIn(f, iterTyp, methods...)
		for _, instr := range instrs {
			pos := instr.Pos()
			if pos == token.NoPos {
				continue
			}
			line := pass.Fset.File(pos).Line(pos)
			if !cmaps.IgnoreLine(pass.Fset, line, "zagane") &&
				!cmaps.IgnoreLine(pass.Fset, line, "unstopiter") {
				pass.Reportf(pos, "iterator must be stopped")
			}
		}
	}

	return nil, nil
}
