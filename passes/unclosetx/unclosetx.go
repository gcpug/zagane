package unclosetx

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
		instrs := detectUnclosedTx(f, txTyp, methods)
		for _, instr := range instrs {
			pos := instr.Pos()
			if pos == token.NoPos {
				continue
			}
			line := pass.Fset.File(pos).Line(pos)
			if !cmaps.IgnoreLine(pass.Fset, line, "zagane") &&
				!cmaps.IgnoreLine(pass.Fset, line, "unclosetx") &&
				!isSingle(instr, single) {
				pass.Reportf(pos, "transaction must be closed")
			}
		}
	}

	return nil, nil
}

func detectUnclosedTx(f *ssa.Function, txTyp types.Type, methods []*types.Func) []ssa.Instruction {
	instrs := analysisutil.NotCalledIn(f, txTyp, methods...)
	argInstrs := findArgPassedUnclosedTx(f, txTyp, methods)

	seen := map[ssa.Instruction]bool{}
	for _, instr := range instrs {
		seen[instr] = true
	}
	for _, instr := range argInstrs {
		if !seen[instr] {
			instrs = append(instrs, instr)
			seen[instr] = true
		}
	}
	return instrs
}

// findArgPassedUnclosedTx detects cases where a *ReadOnlyTransaction value is passed as an argument to another function but Close() is never called.
// analysisutil.NotCalledIn skips these cases due to its internal isArg check, so this function complements it.
func findArgPassedUnclosedTx(f *ssa.Function, txTyp types.Type, methods []*types.Func) []ssa.Instruction {
	var result []ssa.Instruction
	for _, b := range f.Blocks {
		for _, instr := range b.Instrs {
			v, ok := instr.(ssa.Value)
			if !ok || v == nil {
				continue
			}
			if !types.Identical(v.Type(), txTyp) {
				continue
			}
			refs := v.Referrers()
			if refs == nil {
				continue
			}
			hasClose := false
			isPassedAsArg := false
			for _, ref := range *refs {
				for _, m := range methods {
					if analysisutil.Called(ref, v, m) {
						hasClose = true
					}
				}
				if isArgOf(ref, v) {
					isPassedAsArg = true
				}
			}
			if isPassedAsArg && !hasClose {
				result = append(result, instr)
			}
		}
	}
	return result
}

// isArgOf reports whether instr is a call instruction and v is passed as one of its arguments.
func isArgOf(instr ssa.Instruction, v ssa.Value) bool {
	call, ok := instr.(ssa.CallInstruction)
	if !ok {
		return false
	}
	common := call.Common()
	if common == nil {
		return false
	}
	args := common.Args
	if common.Signature().Recv() != nil && len(args) > 0 {
		args = args[1:]
	}
	for _, arg := range args {
		if arg == v {
			return true
		}
	}
	return false
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
