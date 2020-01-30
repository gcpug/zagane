package unclosetx

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
	"golang.org/x/tools/go/ssa"
)

var closeMethods = "Close"

var Analyzer = &analysis.Analyzer{
	Name: "unclosetx",
	Doc:  Doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		buildssa.Analyzer,
		commentmap.Analyzer,
	},
}

const Doc = "unclosetx finds transactions which does not close"

func run(pass *analysis.Pass) (interface{}, error) {
	funcs := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA).SrcFuncs
	cmaps := pass.ResultOf[commentmap.Analyzer].(comment.Maps)

	txTyp := zaganeutils.TypeOf(pass, "*ReadOnlyTransaction")
	if txTyp == nil {
		// skip checking
		return nil, nil
	}

	var methods []*types.Func
	for _, s := range strings.Split(closeMethods, ",") {
		if m := analysisutil.MethodOf(txTyp, s); m != nil {
			methods = append(methods, m)
		}
	}

	cliTyp := zaganeutils.TypeOf(pass, "*Client")
	single := analysisutil.MethodOf(cliTyp, "Single")
	if single == nil {
		// skip checking
		return nil, nil
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
				line := pass.Fset.File(pos).Line(pos)

				// skip
				if cmaps.IgnoreLine(pass.Fset, line, "zagane") ||
					cmaps.IgnoreLine(pass.Fset, line, "unclosetx") ||
					isSingle(b.Instrs[i], single) {
					continue
				}

				called, ok := analysisutil.CalledFrom(b, i, txTyp, methods...)
				if ok && !called {
					pass.Reportf(pos, "transaction must be closed")
				}
			}
		}
	}

	return nil, nil
}

func isSingle(instr ssa.Instruction, single *types.Func) bool {
	call, ok := instr.(ssa.CallInstruction)
	if !ok {
		return false
	}

	common := call.Common()
	if common == nil {
		return false
	}

	callee := common.StaticCallee()
	if callee == nil {
		return false
	}

	fn, ok := callee.Object().(*types.Func)
	if !ok {
		return false
	}

	return fn == single
}
