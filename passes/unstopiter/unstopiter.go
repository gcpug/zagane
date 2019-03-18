package unstopiter

import (
	"fmt"
	"go/types"

	"github.com/gostaticanalysis/analysisutil"
	"github.com/gostaticanalysis/comment"
	"github.com/gostaticanalysis/comment/passes/commentmap"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

var Analyzer = &analysis.Analyzer{
	Name: "unstopiter",
	Doc:  Doc,
	Run:  new(runner).run,
	Requires: []*analysis.Analyzer{
		buildssa.Analyzer,
		commentmap.Analyzer,
	},
}

const (
	Doc = "unstopiter finds iterators which did not stop"

	spannerPath = "cloud.google.com/go/spanner"
)

type runner struct {
	iterObj   types.Object
	iterNamed *types.Named
	iterTyp   *types.Pointer
	stopMthd  *types.Func
}

func (r *runner) run(pass *analysis.Pass) (interface{}, error) {
	funcs := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA).SrcFuncs
	cmaps := pass.ResultOf[commentmap.Analyzer].(comment.Maps)

	r.iterObj = analysisutil.LookupFromImports(pass.Pkg.Imports(), spannerPath, "RowIterator")
	if r.iterObj == nil {
		return nil, fmt.Errorf("cannot find spanner.RowIterator")
	}

	iterNamed, ok := r.iterObj.Type().(*types.Named)
	if !ok {
		return nil, fmt.Errorf("cannot find spanner.RowIterator")
	}
	r.iterNamed = iterNamed
	r.iterTyp = types.NewPointer(r.iterNamed)

	for i := 0; i < r.iterNamed.NumMethods(); i++ {
		if mthd := r.iterNamed.Method(i); mthd.Id() == "Stop" {
			r.stopMthd = mthd
		}
	}
	if r.stopMthd == nil {
		return nil, fmt.Errorf("cannot find spanner.RowIterator.Stop")
	}

	for _, f := range funcs {
		for _, b := range f.Blocks {
			for i := range b.Instrs {
				pos := b.Instrs[i].Pos()
				if !cmaps.IgnorePos(pos, "zagane") &&
					!cmaps.IgnorePos(pos, "unstopiter") &&
					r.unstop(b, i) {
					pass.Reportf(pos, "iterator must be stop")
				}
			}
		}
	}

	return nil, nil
}

func (r *runner) unstop(b *ssa.BasicBlock, i int) bool {
	call, ok := b.Instrs[i].(*ssa.Call)
	if !ok {
		return false
	}

	if !types.Identical(call.Type(), r.iterTyp) {
		return false
	}

	if r.callStopIn(b.Instrs[i:], call) {
		return false
	}

	for _, nb := range b.Succs {
		if r.callStopIn(nb.Instrs, call) {
			return false
		}
	}

	return true
}

func (r *runner) callStopIn(instrs []ssa.Instruction, call *ssa.Call) bool {
	for _, instr := range instrs {
		switch instr := instr.(type) {
		case ssa.CallInstruction:
			fn := instr.Common().StaticCallee()
			args := instr.Common().Args
			if fn.RelString(fn.Package().Pkg) == "(*RowIterator).Stop" &&
				types.Identical(fn.Signature, r.stopMthd.Type()) &&
				len(args) != 0 && call == args[0] {
				return true
			}
		}
	}
	return false
}
