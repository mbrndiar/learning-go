package tasks

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// tasksImportPrefix is the import-path root shared by the starter and
// solution implementations of this applied project.
const tasksImportPrefix = "github.com/mbrndiar/learning-go/projects/tasks"

// implementationRoots enumerates the parallel trees whose production code
// must respect the architecture boundaries checked by this file.
var implementationRoots = []string{"starter", "solution"}

// boundaryViolation records exactly which production file broke which
// architecture rule by importing which package, so failures are actionable
// without re-deriving the offending edge from a generic diff.
type boundaryViolation struct {
	file   string
	imp    string
	rule   string
	detail string
}

func (v boundaryViolation) String() string {
	return fmt.Sprintf("%s: import %q violates rule %q (%s)", v.file, v.imp, v.rule, v.detail)
}

// TestImplementationImportBoundaries walks the production (non-test) Go
// files under projects/tasks/starter and projects/tasks/solution and checks
// every import of a projects/tasks/... package against the architecture
// rules for the importing package's zone.
func TestImplementationImportBoundaries(t *testing.T) {
	var violations []boundaryViolation
	for _, root := range implementationRoots {
		found, err := checkRootBoundaries(root)
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
		violations = append(violations, found...)
	}
	if len(violations) == 0 {
		return
	}
	sort.Slice(violations, func(i, j int) bool {
		if violations[i].file != violations[j].file {
			return violations[i].file < violations[j].file
		}
		return violations[i].imp < violations[j].imp
	})
	var messages []string
	for _, v := range violations {
		messages = append(messages, v.String())
	}
	t.Fatalf("architecture boundary violations:\n%s", strings.Join(messages, "\n"))
}

// checkRootBoundaries walks a single implementation root ("starter" or
// "solution") under projects/tasks and returns every boundary violation
// found in its production files.
func checkRootBoundaries(root string) ([]boundaryViolation, error) {
	rootDir := filepath.Join(root)
	fileset := token.NewFileSet()
	var violations []boundaryViolation

	walkErr := filepath.WalkDir(rootDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		relDir, relErr := filepath.Rel(rootDir, filepath.Dir(path))
		if relErr != nil {
			return relErr
		}
		relDir = filepath.ToSlash(relDir)
		if relDir == "." {
			relDir = ""
		}

		zone, zoneErr := classifyZone(relDir)
		if zoneErr != nil {
			violations = append(violations, boundaryViolation{
				file:   path,
				imp:    relDir,
				rule:   "unexpected-top-level-package",
				detail: zoneErr.Error(),
			})
			return nil
		}

		imports, parseErr := productionImports(fileset, path)
		if parseErr != nil {
			return parseErr
		}
		for _, imp := range imports {
			if rule, detail, bad := checkImport(root, zone, imp); bad {
				violations = append(violations, boundaryViolation{
					file:   path,
					imp:    imp,
					rule:   rule,
					detail: detail,
				})
			}
		}
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}
	return violations, nil
}

// productionImports parses only the import block of a Go source file (no
// declarations, no bodies) and returns the unquoted import paths.
func productionImports(fileset *token.FileSet, path string) ([]string, error) {
	file, err := parser.ParseFile(fileset, path, nil, parser.ImportsOnly)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	imports := make([]string, 0, len(file.Imports))
	for _, spec := range file.Imports {
		value, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, fmt.Errorf("unquote import in %s: %w", path, err)
		}
		imports = append(imports, value)
	}
	return imports, nil
}

// zone identifies which architecture layer a package belongs to within one
// implementation root, driving which projects/tasks imports it may use.
type zone string

const (
	zoneTask        zone = "task"
	zoneClient      zone = "client"
	zoneServer      zone = "server"
	zoneCmdTasks    zone = "cmd/tasks"
	zoneCmdTasksAPI zone = "cmd/tasks-api"
)

// classifyZone maps a package directory, relative to an implementation
// root, to its architecture zone. Any top-level implementation directory
// other than task, client, server, and cmd (with cmd restricted to the
// explicit tasks and tasks-api entry points) is rejected so the classifier
// cannot silently accept an unexpected or misspelled prefix.
func classifyZone(relDir string) (zone, error) {
	switch {
	case relDir == "task" || hasPathPrefix(relDir, "task"):
		return zoneTask, nil
	case relDir == "client" || hasPathPrefix(relDir, "client"):
		return zoneClient, nil
	case relDir == "server" || hasPathPrefix(relDir, "server"):
		return zoneServer, nil
	case relDir == "cmd/tasks":
		return zoneCmdTasks, nil
	case relDir == "cmd/tasks-api":
		return zoneCmdTasksAPI, nil
	case relDir == "cmd" || hasPathPrefix(relDir, "cmd"):
		return "", fmt.Errorf("package %q is under cmd but is neither cmd/tasks nor cmd/tasks-api", relDir)
	default:
		return "", fmt.Errorf("package %q is not under a recognized top-level implementation directory (task, client, server, cmd)", relDir)
	}
}

// hasPathPrefix reports whether relDir is prefix or a strict subdirectory of
// prefix, using path-segment boundaries so "clientx" never matches "client".
func hasPathPrefix(relDir, prefix string) bool {
	return relDir == prefix || strings.HasPrefix(relDir, prefix+"/")
}

// checkImport validates one projects/tasks/... import against the rules for
// the given root ("starter" or "solution") and importing zone. It returns
// the violated rule name and detail, plus whether the import is disallowed.
// Imports outside the projects/tasks tree (standard library, third-party
// modules, or other capstones) are always allowed.
func checkImport(root string, z zone, imp string) (rule string, detail string, bad bool) {
	if !strings.HasPrefix(imp, tasksImportPrefix+"/") {
		return "", "", false
	}
	rest := strings.TrimPrefix(imp, tasksImportPrefix+"/")

	importRoot, importRel, found := splitFirstSegment(rest)
	if !found {
		// The import is directly under projects/tasks (e.g. the tasks
		// package itself, or capstones/testsupport-style helpers); it is
		// not one of the starter/solution implementation trees.
		return "", "", false
	}

	if importRoot != root {
		if importRoot == "starter" || importRoot == "solution" {
			return "cross-root-import", fmt.Sprintf("%s production code must not import the %s tree", root, importRoot), true
		}
		return "", "", false
	}

	switch z {
	case zoneTask:
		return "task-no-implementation-imports", fmt.Sprintf("task/** must not import any %s implementation package, got %q", root, importRel), true

	case zoneClient:
		if hasPathPrefix(importRel, "server") {
			return "client-no-server-import", "client/** must not import server/**", true
		}
		if hasPathPrefix(importRel, "client") || hasPathPrefix(importRel, "task") {
			return "", "", false
		}
		return "client-scope", "client/** may only depend on its own client subtree and task among implementation packages", true

	case zoneServer:
		if hasPathPrefix(importRel, "client") {
			return "server-no-client-import", "server/** must not import client/**", true
		}
		if hasPathPrefix(importRel, "server") || hasPathPrefix(importRel, "task") {
			return "", "", false
		}
		return "server-scope", "server/** may only depend on its own server subtree and task among implementation packages", true

	case zoneCmdTasks:
		if hasPathPrefix(importRel, "client") || hasPathPrefix(importRel, "task") {
			return "", "", false
		}
		return "cmd-tasks-scope", "cmd/tasks/** may only import the client subtree and task among implementation packages", true

	case zoneCmdTasksAPI:
		if hasPathPrefix(importRel, "server") || hasPathPrefix(importRel, "task") {
			return "", "", false
		}
		return "cmd-tasks-api-scope", "cmd/tasks-api/** may only import the server subtree and task among implementation packages", true

	default:
		return "unknown-zone", fmt.Sprintf("unrecognized zone %q", z), true
	}
}

// splitFirstSegment splits a slash-separated relative import path into its
// first path segment and the remainder. found is false if rest is empty.
func splitFirstSegment(rest string) (first, remainder string, found bool) {
	if rest == "" {
		return "", "", false
	}
	if idx := strings.Index(rest, "/"); idx >= 0 {
		return rest[:idx], rest[idx+1:], true
	}
	return rest, "", true
}
