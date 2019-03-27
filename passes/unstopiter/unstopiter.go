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
	methods   []*types.Func
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
		case "Stop", "Do":
			r.methods = append(r.methods, mthd)
		}
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

				// skip
				if cmaps.IgnorePos(pos, "zagane") &&
					cmaps.IgnorePos(pos, "unstopiter") {
					continue
				}

				called, ok := analysisutil.CalledFrom(b, i, r.iterTyp, r.methods...)
				if ok && !called {
					pass.Reportf(pos, "iterator must be stop")
				}
			}
		}
	}

	return nil, nil
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
