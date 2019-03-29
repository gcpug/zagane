package zaganeutils

import (
	"go/ast"
	"go/types"
	"strconv"

	"github.com/gostaticanalysis/analysisutil"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ssa"
)

// ImportPath is import path of spanner package.
const ImportPath = "cloud.google.com/go/spanner"

// ObjectOf returns types.Object by given name in spanner package.
func ObjectOf(pass *analysis.Pass, name string) types.Object {
	return analysisutil.ObjectOf(pass, ImportPath, name)
}

// TypeOf returns types.Type by given name in spanner package.
// TypeOf accepts pointer types such as *Client.
func TypeOf(pass *analysis.Pass, name string) types.Type {
	return analysisutil.TypeOf(pass, ImportPath, name)
}

// Unimported returns whether file which has function f
// does not import spanner package.
func Unimported(pass *analysis.Pass, f *ssa.Function, skipFile map[*ast.File]bool) (ret bool) {
	obj := f.Object()
	if obj == nil {
		return false
	}

	file := analysisutil.File(pass, obj.Pos())
	if file == nil {
		return false
	}

	if skip, has := skipFile[file]; has {
		return skip
	}
	defer func() {
		skipFile[file] = ret
	}()

	for _, impt := range file.Imports {
		path, err := strconv.Unquote(impt.Path.Value)
		if err != nil {
			continue
		}
		path = analysisutil.RemoveVendor(path)
		if path == ImportPath {
			return false
		}
	}

	return true
}

// FromSpanner whether v came from spanner pacakge.
func FromSpanner(v ssa.Value) bool {
	switch v := v.(type) {
	case *ssa.Extract:
		return FromSpanner(v.Tuple)
	case ssa.CallInstruction:
		common := v.Common()
		if common == nil {
			return false
		}
		fn := common.StaticCallee()
		if fn == nil {
			return false
		}

		pkg := fn.Pkg
		if pkg == nil {
			return false
		}

		path := analysisutil.RemoveVendor(pkg.Pkg.Path())
		return path == ImportPath
	}
	return false
}
