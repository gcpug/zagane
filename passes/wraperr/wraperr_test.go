package wraperr_test

import (
	"testing"

	"github.com/gcpug/zagane/passes/wraperr"
	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, wraperr.Analyzer, "a")
}