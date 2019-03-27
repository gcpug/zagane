package zaganeutils

import (
	"go/types"

	"github.com/gostaticanalysis/analysisutil"
	"golang.org/x/tools/go/analysis"
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
