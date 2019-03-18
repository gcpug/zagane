package unstopiter_test

import (
	"testing"

	"github.com/gcpug/zagane/passes/unstopiter"
	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, unstopiter.Analyzer, "a")
}