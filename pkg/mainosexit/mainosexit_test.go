package mainosexit

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestNoOsExitAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer,
		"main", "other", "mainfunc", "notmain", "without",
	)
}
