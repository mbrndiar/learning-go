// Package testsupport contains reusable capstone harness utilities.
package testsupport

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

// CompareExportedSurface verifies that paired implementation trees expose the
// same declarations while allowing their function bodies to differ.
func CompareExportedSurface(starterRoot, solutionRoot string) error {
	starter, err := exportedSurface(starterRoot)
	if err != nil {
		return fmt.Errorf("read starter surface: %w", err)
	}
	solution, err := exportedSurface(solutionRoot)
	if err != nil {
		return fmt.Errorf("read solution surface: %w", err)
	}
	if reflect.DeepEqual(starter, solution) {
		return nil
	}
	return fmt.Errorf("exported surfaces differ:\nstarter: %v\nsolution: %v", starter, solution)
}

func exportedSurface(root string) (map[string][]string, error) {
	directories := make(map[string]struct{})
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		directories[filepath.Dir(path)] = struct{}{}
		return nil
	})
	if err != nil {
		return nil, err
	}

	surface := make(map[string][]string, len(directories))
	for directory := range directories {
		relative, err := filepath.Rel(root, directory)
		if err != nil {
			return nil, err
		}
		declarations, err := packageSurface(directory)
		if err != nil {
			return nil, err
		}
		surface[filepath.ToSlash(relative)] = declarations
	}
	return surface, nil
}

func packageSurface(directory string) ([]string, error) {
	fileset := token.NewFileSet()
	packages, err := parser.ParseDir(fileset, directory, func(info fs.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	if len(packages) != 1 {
		return nil, fmt.Errorf("%s contains %d packages, want 1", directory, len(packages))
	}

	var declarations []string
	for _, parsed := range packages {
		documented := doc.New(parsed, "", 0)
		for _, value := range documented.Consts {
			declarations = append(declarations, renderNode(fileset, value.Decl))
		}
		for _, value := range documented.Vars {
			declarations = append(declarations, renderNode(fileset, value.Decl))
		}
		for _, function := range documented.Funcs {
			declarations = append(declarations, renderFunction(fileset, function.Decl))
		}
		for _, namedType := range documented.Types {
			declarations = append(declarations, renderNode(fileset, namedType.Decl))
			for _, function := range namedType.Funcs {
				declarations = append(declarations, renderFunction(fileset, function.Decl))
			}
			for _, method := range namedType.Methods {
				declarations = append(declarations, renderFunction(fileset, method.Decl))
			}
		}
	}
	sort.Strings(declarations)
	return declarations, nil
}

func renderFunction(fileset *token.FileSet, declaration *ast.FuncDecl) string {
	signature := *declaration
	signature.Body = nil
	return renderNode(fileset, &signature)
}

func renderNode(fileset *token.FileSet, node ast.Node) string {
	var output strings.Builder
	if err := format.Node(&output, fileset, node); err != nil {
		return fmt.Sprintf("<format error: %v>", err)
	}
	return output.String()
}
