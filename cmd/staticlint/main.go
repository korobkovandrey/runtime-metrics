// Package staticlint provides a multichecker that combines multiple static analysis tools
// for Go code to ensure high code quality by detecting potential bugs, performance issues,
// and style inconsistencies.
//
// Launching the multichecker:
// To run the multichecker, execute the following command:
//
//	go run cmd/staticlint/main.go <path to package or file>
//
// Example:
//
//	go run cmd/staticlint/main.go ./...
//
// This multichecker includes the following groups of analyzers:
//  1. All standard static analyzers from the package golang.org/x/tools/go/analysis/passes
//  2. All SA-class analyzers from the package honnef.co/go/tools/staticcheck
//  3. One analyzer each from the S, ST, and QF classes from honnef.co/go/tools/simple,
//     stylecheck, and quickfix packages
//  4. Two public analyzers: errcheck and ineffassign
//  5. A custom analyzer: osexit
//
// Standard Analyzers (golang.org/x/tools/go/analysis/passes):
// - appends: Checks for correct usage of the append function
// - asmdecl: Verifies consistency between Go and assembly declarations
// - assign: Detects useless assignments
// - atomic: Ensures correct usage of the sync/atomic package
// - bools: Identifies common errors in boolean expressions
// - buildtag: Validates build tags
// - cgocall: Checks calls to C code via cgo
// - composite: Detects uninitialized composite literals
// - copylock: Checks for copying locks by value
// - deepequalerrors: Ensures reflect.DeepEqual is not used with errors
// - errorsas: Verifies correct usage of errors.As
// - fieldalignment: Identifies struct fields that can be reordered for memory optimization
// - findcall: Detects calls to a specified function
// - framepointer: Checks frame pointer usage
// - httpresponse: Identifies common errors in HTTP responses
// - ifaceassert: Detects redundant interface assertions
// - loopclosure: Checks for issues with closures in loops
// - lostcancel: Detects unclosed contexts
// - nilfunc: Checks for comparisons of functions with nil
// - nilness: Analyzes nil values in expressions
// - pkgfact: Verifies package facts
// - printf: Ensures correct format specifiers in printf-like functions
// - reflectvaluecompare: Checks comparisons of reflect.Value
// - shadow: Detects variable shadowing
// - shift: Identifies invalid shifts
// - sigchanyzer: Checks for incorrect signal handling
// - sortslice: Verifies slice sorting
// - stdmethods: Ensures standard method signatures
// - stringintconv: Checks string-to-integer conversions
// - structtag: Validates struct tags
// - testinggoroutine: Detects goroutine calls in tests
// - tests: Identifies common errors in tests
// - timeformat: Verifies time format strings
// - unmarshal: Ensures correct usage of unmarshal functions
// - unreachable: Detects unreachable code
// - unsafeptr: Checks for incorrect usage of unsafe.Pointer
// - unusedresult: Detects unused function results
// - unusedwrite: Identifies unused writes
// - usesgenerics: Verifies usage of generic types
//
// Staticcheck Analyzers (honnef.co/go/tools/staticcheck):
//   - SAxxxx: SA-class analyzers detect potential bugs and performance issues, such as
//     incorrect function usage, pointer-related errors, and concurrency issues.
//
// Additional Staticcheck Analyzers:
// - S1006: Checks for usage of for range instead of indexed for loops in simple cases
// - ST1015: Verifies that exported functions have documentation
// - QF1001: Suggests converting loops to more efficient queries
//
// Public Analyzers:
// - errcheck: Detects unhandled errors in the code
// - ineffassign: Identifies ineffective or useless assignments
//
// Custom Analyzer:
// - mainosexit: Prohibits direct calls to os.Exit in the main function of the main package
package main

import (
	"github.com/korobkovandrey/runtime-metrics/pkg/mainosexit"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/quickfix/qf1001"
	"honnef.co/go/tools/simple/s1006"
	"honnef.co/go/tools/stylecheck/st1015"

	"honnef.co/go/tools/staticcheck"

	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/kisielk/errcheck/errcheck"
)

func main() {
	var analyzers []*analysis.Analyzer

	analyzers = append(analyzers,
		appends.Analyzer,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		deepequalerrors.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		findcall.Analyzer,
		framepointer.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		pkgfact.Analyzer,
		printf.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
		usesgenerics.Analyzer,
	)

	for _, a := range staticcheck.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}
	analyzers = append(analyzers,
		s1006.Analyzer, st1015.Analyzer, qf1001.Analyzer,
		errcheck.Analyzer, ineffassign.Analyzer,
		mainosexit.Analyzer,
	)
	multichecker.Main(analyzers...)
}
