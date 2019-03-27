package unclosetx_test

import (
	"testing"

	"github.com/gcpug/zagane/passes/unclosetx"
	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, unclosetx.Analyzer, "a")
}