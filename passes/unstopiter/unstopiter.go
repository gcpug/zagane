package unstopiter

import (
	"fmt"
	"go/ast"
	"go/types"
	"strconv"

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
	pass      *analysis.Pass
	iterObj   types.Object
	iterNamed *types.Named
	iterTyp   *types.Pointer
	stopMthd  *types.Func
	doMthd    *types.Func
	skipFile  map[*ast.File]bool
}

func (r *runner) run(pass *analysis.Pass) (interface{}, error) {
	r.pass = pass
	funcs := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA).SrcFuncs
	cmaps := pass.ResultOf[commentmap.Analyzer].(comment.Maps)

	r.iterObj = analysisutil.LookupFromImports(pass.Pkg.Imports(), spannerPath, "RowIterator")
	if r.iterObj == nil {
		// skip checking
		return nil, nil
	}

	iterNamed, ok := r.iterObj.Type().(*types.Named)
	if !ok {
		return nil, fmt.Errorf("cannot find spanner.RowIterator")
	}
	r.iterNamed = iterNamed
	r.iterTyp = types.NewPointer(r.iterNamed)

	for i := 0; i < r.iterNamed.NumMethods(); i++ {
		mthd := r.iterNamed.Method(i)
		switch mthd.Id() {
		case "Stop":
			r.stopMthd = mthd
		case "Do":
			r.doMthd = mthd
		}
	}
	if r.stopMthd == nil {
		return nil, fmt.Errorf("cannot find spanner.RowIterator.Stop")
	}
	if r.doMthd == nil {
		return nil, fmt.Errorf("cannot find spanner.RowIterator.Do")
	}

	r.skipFile = map[*ast.File]bool{}
	for _, f := range funcs {
		if r.noImportedSpanner(f) {
			// skip this
			continue
		}

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

	if r.callStopInSuccs(b, call, map[*ssa.BasicBlock]bool{}) {
		return false
	}

	return true
}

func (r *runner) callStopIn(instrs []ssa.Instruction, call *ssa.Call) bool {
	for _, instr := range instrs {
		switch instr := instr.(type) {
		case ssa.CallInstruction:
			if analysisutil.Called(instr, call, r.stopMthd) ||
				analysisutil.Called(instr, call, r.doMthd) {
				return true
			}
		}
	}
	return false
}

func (r *runner) callStopInSuccs(b *ssa.BasicBlock, call *ssa.Call, done map[*ssa.BasicBlock]bool) bool {
	if done[b] {
		return false
	}
	done[b] = true

	if len(b.Succs) == 0 {
		return r.isReturnIter(b.Instrs, call)
	}

	for _, s := range b.Succs {
		if !r.callStopIn(s.Instrs, call) &&
			!r.callStopInSuccs(s, call, done) {
			return false
		}
	}

	return true
}

func (r *runner) isReturnIter(instrs []ssa.Instruction, call *ssa.Call) bool {
	if len(instrs) == 0 {
		return false
	}

	ret, isRet := instrs[len(instrs)-1].(*ssa.Return)
	if !isRet {
		return false
	}

	for _, r := range ret.Results {
		if r == call {
			return true
		}
	}

	return false
}

func (r *runner) noImportedSpanner(f *ssa.Function) (ret bool) {
	obj := f.Object()
	if obj == nil {
		return false
	}

	file := analysisutil.File(r.pass, obj.Pos())
	if file == nil {
		return false
	}

	if skip, has := r.skipFile[file]; has {
		return skip
	}
	defer func() {
		r.skipFile[file] = ret
	}()

	for _, impt := range file.Imports {
		path, err := strconv.Unquote(impt.Path.Value)
		if err != nil {
			continue
		}
		path = analysisutil.RemoveVendor(path)
		if path == spannerPath {
			return false
		}
	}

	return true
}
