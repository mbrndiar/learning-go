package textproc

// This file is test infrastructure, not an exercise file: do not edit it.
//
// It exists because `go test` alone can silently "succeed" even when a
// benchmark or fuzz target was never written: Benchmark* functions only run
// with `-bench`, Fuzz* functions only run with `-fuzz`, and a t.Skip/b.Skip/
// f.Skip call makes a test report as skipped rather than failed. A learner
// could leave every TODO stub untouched and still see `go test ./...` print
// "ok". TestRequiredTestsExist guards against that by statically inspecting
// the package's own test source for the required function names and for
// telltale signs that each was actually completed (no leftover Skip call,
// table-driven use of t.Run, a real b.N loop, and a seeded f.Fuzz call).

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
)

// requiredTest describes one function that TestRequiredTestsExist expects to
// find, fully implemented, somewhere among this package's *_test.go files.
type requiredTest struct {
	name          string // exact function name, e.g. "TestNormalize"
	kind          string // "test", "benchmark", or "fuzz", for error messages
	needsRunLoop  bool   // require a range/for loop that calls t.Run (table-driven)
	needsBN       bool   // require a reference to b.N (real benchmark loop)
	needsFuzz     bool   // require an f.Fuzz(...) call
	needsFuzzSeed bool   // require at least one f.Add(...) seed call
}

var requiredTests = []requiredTest{
	{name: "TestNormalize", kind: "test", needsRunLoop: true},
	{name: "TestWordFrequency", kind: "test", needsRunLoop: true},
	{name: "TestReverse", kind: "test", needsRunLoop: true},
	{name: "TestSafeSlice", kind: "test", needsRunLoop: true},
	{name: "BenchmarkWordFrequency", kind: "benchmark", needsBN: true},
	{name: "FuzzSafeSlice", kind: "fuzz", needsFuzz: true, needsFuzzSeed: true},
}

// TestRequiredTestsExist parses every *_test.go file in this package
// directory and fails with a precise, per-item report if any required test,
// benchmark, or fuzz function is missing, still skipped, or missing the
// structural marker (t.Run/b.N/f.Fuzz/f.Add) that shows real work was done.
func TestRequiredTestsExist(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine this file's location via runtime.Caller")
	}
	dir := filepath.Dir(thisFile)

	files, err := filepath.Glob(filepath.Join(dir, "*_test.go"))
	if err != nil {
		t.Fatalf("globbing test files: %v", err)
	}
	if len(files) == 0 {
		t.Fatalf("no _test.go files found in %s", dir)
	}

	fset := token.NewFileSet()
	found := map[string]*ast.FuncDecl{}
	for _, file := range files {
		f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parsing %s: %v", file, err)
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil {
				continue
			}
			found[fn.Name.Name] = fn
		}
	}

	var problems []string
	for _, rt := range requiredTests {
		fn, ok := found[rt.name]
		if !ok {
			problems = append(problems, fmt.Sprintf("%s %s is missing (see README)", rt.kind, rt.name))
			continue
		}
		if callsSkip(fn) {
			problems = append(problems, fmt.Sprintf(
				"%s %s still calls Skip; replace the TODO body with a real implementation", rt.kind, rt.name))
			continue
		}
		if rt.needsRunLoop && !hasRunInsideLoop(fn) {
			problems = append(problems, fmt.Sprintf(
				"%s %s must be table-driven: call t.Run(...) from inside a for/range loop over a cases table",
				rt.kind, rt.name))
		}
		if rt.needsBN && !referencesBN(fn) {
			problems = append(problems, fmt.Sprintf(
				"benchmark %s must loop using b.N (e.g. \"for i := 0; i < b.N; i++\")", rt.name))
		}
		if rt.needsFuzz && countCalls(fn, "Fuzz") == 0 {
			problems = append(problems, fmt.Sprintf(
				"fuzz target %s must call f.Fuzz with a fuzz function", rt.name))
		}
		if rt.needsFuzzSeed && countCalls(fn, "Add") == 0 {
			problems = append(problems, fmt.Sprintf(
				"fuzz target %s must seed the corpus with at least one f.Add(...) call", rt.name))
		}
	}

	if len(problems) > 0 {
		sort.Strings(problems)
		t.Fatalf("required tests are incomplete:\n  - %s", joinLines(problems))
	}
}

// callsSkip reports whether fn's body contains a call whose selector is
// Skip or Skipf (t.Skip, b.Skip, f.Skip, t.Skipf, ...).
func callsSkip(fn *ast.FuncDecl) bool {
	skipped := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if sel.Sel.Name == "Skip" || sel.Sel.Name == "Skipf" {
			skipped = true
		}
		return true
	})
	return skipped
}

// hasRunInsideLoop reports whether fn's body contains a for or range loop
// whose body (directly or nested) calls x.Run(...), which is the shape of an
// idiomatic table-driven test iterating over a cases slice.
func hasRunInsideLoop(fn *ast.FuncDecl) bool {
	found := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		var loopBody *ast.BlockStmt
		switch loop := n.(type) {
		case *ast.RangeStmt:
			loopBody = loop.Body
		case *ast.ForStmt:
			loopBody = loop.Body
		default:
			return true
		}
		ast.Inspect(loopBody, func(inner ast.Node) bool {
			call, ok := inner.(*ast.CallExpr)
			if !ok {
				return true
			}
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Run" {
				found = true
			}
			return true
		})
		return true
	})
	return found
}

// countCalls counts calls of the form x.<method> anywhere in fn's body.
func countCalls(fn *ast.FuncDecl, method string) int {
	count := 0
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == method {
			count++
		}
		return true
	})
	return count
}

// referencesBN reports whether fn's body refers to a selector named N
// (as in b.N), which is required for a benchmark to actually measure work.
func referencesBN(fn *ast.FuncDecl) bool {
	found := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if ok && sel.Sel.Name == "N" {
			found = true
		}
		return true
	})
	return found
}

func joinLines(lines []string) string {
	out := ""
	for i, line := range lines {
		if i > 0 {
			out += "\n  - "
		}
		out += line
	}
	return out
}
