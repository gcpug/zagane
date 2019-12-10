package wraperr

import (
	"go/ast"
	"go/token"
	"go/types"

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
	Run:  new(runner).run,
	Requires: []*analysis.Analyzer{
		buildssa.Analyzer,
		commentmap.Analyzer,
	},
}

const Doc = "wraperr finds ReadWriteTransaction calls which returns wrapped errors"

type runner struct {
	pass                *analysis.Pass
	spannerError        types.Type
	grpcStatusInterface *types.Interface
}

func (r *runner) run(pass *analysis.Pass) (interface{}, error) {
	r.pass = pass
	r.grpcStatusInterface = newGRPCStatusInterface(pass)
	r.spannerError = zaganeutils.TypeOf(pass, "*Error")
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

				if pos := r.wrapped(instr); pos != token.NoPos {
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

func (r *runner) wrapped(instr ssa.Instruction) token.Pos {
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
		return r.returnedWrappedErr(fnc.Fn)
	}

	return token.NoPos
}

func (r *runner) returnedWrappedErr(v ssa.Value) token.Pos {
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

		if r.implementsGRPCStatus(v) ||
			r.isSpannerError(v) {
			continue
		}

		if !zaganeutils.FromSpanner(v) {
			switch v := v.(type) {
			case *ssa.MakeInterface:
				return v.X.Pos()
			case *ssa.Call:
				if r.returnedWrappedErr(v.Common().Value) != token.NoPos {
					return v.Pos()
				}
			}
		}
	}
	return token.NoPos
}

func (r *runner) implementsGRPCStatus(v ssa.Value) bool {
	if r.grpcStatusInterface == nil {
		return false
	}
	switch v := v.(type) {
	case *ssa.MakeInterface:
		return types.Implements(v.X.Type(), r.grpcStatusInterface)
	}
	return types.Implements(v.Type(), r.grpcStatusInterface)
}

func (r *runner) isSpannerError(v ssa.Value) bool {
	if r.spannerError == nil {
		return false
	}
	switch v := v.(type) {
	case *ssa.MakeInterface:
		return types.Identical(v.X.Type(), r.spannerError)
	}
	return types.Identical(v.Type(), r.spannerError)
}

func newGRPCStatusInterface(pass *analysis.Pass) *types.Interface {
	typStatus := analysisutil.TypeOf(pass, "google.golang.org/grpc/status", "*Status")
	if typStatus == nil {
		return nil
	}

	ret := types.NewTuple(types.NewParam(token.NoPos, pass.Pkg, "", typStatus))
	sig := types.NewSignature(nil, types.NewTuple(), ret, false)
	grpcStatusFunc := types.NewFunc(token.NoPos, pass.Pkg, "GRPCStatus", sig)
	return types.NewInterfaceType([]*types.Func{grpcStatusFunc}, nil).Complete()
}
