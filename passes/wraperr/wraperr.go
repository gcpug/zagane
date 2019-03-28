package wraperr

import (
	"go/ast"
	"go/token"

	"github.com/gcpug/zagane/zaganeutils"
	"github.com/gostaticanalysis/analysisutil"
	"github.com/gostaticanalysis/comment"
	"github.com/gostaticanalysis/comment/passes/commentmap"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

var Analyzer = &analysis.Analyzer{
	Name: "wraperr",
	Doc:  Doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		buildssa.Analyzer,
		commentmap.Analyzer,
	},
}

const Doc = "wraperr finds ReadWriteTransaction calls which returns wrapped errors"

func run(pass *analysis.Pass) (interface{}, error) {
	cmaps := pass.ResultOf[commentmap.Analyzer].(comment.Maps)
	funcs := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA).SrcFuncs

	cliTyp := zaganeutils.TypeOf(pass, "*Client")
	rwtx := analysisutil.MethodOf(cliTyp, "ReadWriteTransaction")
	if rwtx == nil {
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
			for _, instr := range b.Instrs {

				if !analysisutil.Called(instr, nil, rwtx) {
					continue
				}

				if pos := wrapped(instr); pos != token.NoPos {
					if !cmaps.IgnorePos(pos, "zagane") &&
						!cmaps.IgnorePos(pos, "wraperr") {
						pass.Reportf(pos, "must not be wrapped")
					}
				}
			}
		}
	}

	return nil, nil
}

func wrapped(instr ssa.Instruction) token.Pos {
	call, ok := instr.(ssa.CallInstruction)
	if !ok {
		return token.NoPos
	}

	common := call.Common()
	if common == nil {
		return token.NoPos
	}

	if len(common.Args) != 3 {
		return token.NoPos
	}

	switch fnc := common.Args[2].(type) {
	case *ssa.MakeClosure:
		return returnedWrappedErr(fnc.Fn)
	}

	return token.NoPos
}

func returnedWrappedErr(v ssa.Value) token.Pos {
	for _, ret := range analysisutil.Returns(v) {
		if len(ret.Results) == 0 {
			continue
		}
		v := ret.Results[len(ret.Results)-1]
		if !analysisutil.ImplementsError(v.Type()) {
			continue
		}

		if c, isConst := v.(*ssa.Const); isConst && c.IsNil() {
			continue
		}

		if !zaganeutils.FromSpanner(v) {
			switch v := v.(type) {
			case *ssa.MakeInterface:
				return v.X.Pos()
			default:
				return v.Pos()
			}
		}
	}
	return token.NoPos
}
