package tasks

import "testing"

// TestClassifyZone table-tests the zone classifier in isolation so a typo'd
// or overly broad prefix check (e.g. "client" matching "clientx") cannot
// silently let the architecture test accept packages it shouldn't.
func TestClassifyZone(t *testing.T) {
	tests := []struct {
		name    string
		relDir  string
		want    zone
		wantErr bool
	}{
		{name: "task root", relDir: "task", want: zoneTask},
		{name: "task nested", relDir: "task/sub", want: zoneTask},
		{name: "client root", relDir: "client", want: zoneClient},
		{name: "client nested", relDir: "client/cli", want: zoneClient},
		{name: "client internal nested", relDir: "client/internal/httpcontract", want: zoneClient},
		{name: "server root", relDir: "server", want: zoneServer},
		{name: "server nested", relDir: "server/api/chi", want: zoneServer},
		{name: "cmd tasks", relDir: "cmd/tasks", want: zoneCmdTasks},
		{name: "cmd tasks-api", relDir: "cmd/tasks-api", want: zoneCmdTasksAPI},
		{name: "cmd bare rejected", relDir: "cmd", wantErr: true},
		{name: "cmd unknown subdir rejected", relDir: "cmd/bogus", wantErr: true},
		{name: "lookalike prefix taskrunner rejected", relDir: "taskrunner", wantErr: true},
		{name: "lookalike prefix clientx rejected", relDir: "clientx", wantErr: true},
		{name: "lookalike prefix serverless rejected", relDir: "serverless", wantErr: true},
		{name: "unrelated top level rejected", relDir: "internal", wantErr: true},
		{name: "package root rejected", relDir: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := classifyZone(tt.relDir)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("classifyZone(%q) = %q, nil; want error", tt.relDir, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("classifyZone(%q) unexpected error: %v", tt.relDir, err)
			}
			if got != tt.want {
				t.Fatalf("classifyZone(%q) = %q, want %q", tt.relDir, got, tt.want)
			}
		})
	}
}

// TestCheckImport table-tests the per-zone import rule enforcement in
// isolation, covering every rule plus the near-miss prefixes (e.g.
// "servertools" or "clientside") that a naive strings.HasPrefix check would
// wrongly accept or reject.
func TestCheckImport(t *testing.T) {
	const otherModuleImport = "github.com/example/other/pkg"

	tests := []struct {
		name     string
		root     string
		zone     zone
		imp      string
		wantBad  bool
		wantRule string
	}{
		{
			name: "non-project import always allowed",
			root: "starter", zone: zoneTask, imp: otherModuleImport,
		},
		{
			name: "task package itself is not an implementation root",
			root: "starter", zone: zoneClient, imp: tasksImportPrefix,
		},
		{
			name: "starter importing solution is cross-root",
			root: "starter", zone: zoneTask, imp: tasksImportPrefix + "/solution/task",
			wantBad: true, wantRule: "cross-root-import",
		},
		{
			name: "solution importing starter is cross-root",
			root: "solution", zone: zoneClient, imp: tasksImportPrefix + "/starter/client",
			wantBad: true, wantRule: "cross-root-import",
		},
		{
			name: "other implementation root families are ignored",
			root: "starter", zone: zoneTask, imp: tasksImportPrefix + "/tests/m1",
		},
		{
			name: "task importing client is forbidden",
			root: "starter", zone: zoneTask, imp: tasksImportPrefix + "/starter/client",
			wantBad: true, wantRule: "task-no-implementation-imports",
		},
		{
			name: "task importing server is forbidden",
			root: "starter", zone: zoneTask, imp: tasksImportPrefix + "/starter/server",
			wantBad: true, wantRule: "task-no-implementation-imports",
		},
		{
			name: "task importing cmd is forbidden",
			root: "starter", zone: zoneTask, imp: tasksImportPrefix + "/starter/cmd/tasks",
			wantBad: true, wantRule: "task-no-implementation-imports",
		},
		{
			name: "client importing own subtree is allowed",
			root: "solution", zone: zoneClient, imp: tasksImportPrefix + "/solution/client/cli",
		},
		{
			name: "client importing task is allowed",
			root: "solution", zone: zoneClient, imp: tasksImportPrefix + "/solution/task",
		},
		{
			name: "client importing server is forbidden",
			root: "solution", zone: zoneClient, imp: tasksImportPrefix + "/solution/server",
			wantBad: true, wantRule: "client-no-server-import",
		},
		{
			name: "client importing server lookalike prefix still forbidden",
			root: "solution", zone: zoneClient, imp: tasksImportPrefix + "/solution/server/api/chi",
			wantBad: true, wantRule: "client-no-server-import",
		},
		{
			name: "client importing cmd is out of scope",
			root: "solution", zone: zoneClient, imp: tasksImportPrefix + "/solution/cmd/tasks",
			wantBad: true, wantRule: "client-scope",
		},
		{
			name: "server importing own subtree is allowed",
			root: "starter", zone: zoneServer, imp: tasksImportPrefix + "/starter/server/storage/sqlite",
		},
		{
			name: "server importing task is allowed",
			root: "starter", zone: zoneServer, imp: tasksImportPrefix + "/starter/task",
		},
		{
			name: "server importing client is forbidden",
			root: "starter", zone: zoneServer, imp: tasksImportPrefix + "/starter/client",
			wantBad: true, wantRule: "server-no-client-import",
		},
		{
			name: "server importing cmd is out of scope",
			root: "starter", zone: zoneServer, imp: tasksImportPrefix + "/starter/cmd/tasks-api",
			wantBad: true, wantRule: "server-scope",
		},
		{
			name: "cmd/tasks importing client is allowed",
			root: "solution", zone: zoneCmdTasks, imp: tasksImportPrefix + "/solution/client/resty",
		},
		{
			name: "cmd/tasks importing task is allowed",
			root: "solution", zone: zoneCmdTasks, imp: tasksImportPrefix + "/solution/task",
		},
		{
			name: "cmd/tasks importing server is forbidden",
			root: "solution", zone: zoneCmdTasks, imp: tasksImportPrefix + "/solution/server",
			wantBad: true, wantRule: "cmd-tasks-scope",
		},
		{
			name: "cmd/tasks importing cmd/tasks-api is forbidden",
			root: "solution", zone: zoneCmdTasks, imp: tasksImportPrefix + "/solution/cmd/tasks-api",
			wantBad: true, wantRule: "cmd-tasks-scope",
		},
		{
			name: "cmd/tasks-api importing server is allowed",
			root: "starter", zone: zoneCmdTasksAPI, imp: tasksImportPrefix + "/starter/server",
		},
		{
			name: "cmd/tasks-api importing task is allowed",
			root: "starter", zone: zoneCmdTasksAPI, imp: tasksImportPrefix + "/starter/task",
		},
		{
			name: "cmd/tasks-api importing client is forbidden",
			root: "starter", zone: zoneCmdTasksAPI, imp: tasksImportPrefix + "/starter/client",
			wantBad: true, wantRule: "cmd-tasks-api-scope",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, detail, bad := checkImport(tt.root, tt.zone, tt.imp)
			if bad != tt.wantBad {
				t.Fatalf("checkImport(%q, %q, %q) bad = %v, want %v (rule=%q detail=%q)", tt.root, tt.zone, tt.imp, bad, tt.wantBad, rule, detail)
			}
			if tt.wantBad && rule != tt.wantRule {
				t.Fatalf("checkImport(%q, %q, %q) rule = %q, want %q", tt.root, tt.zone, tt.imp, rule, tt.wantRule)
			}
		})
	}
}

// TestHasPathPrefix guards the segment-boundary helper directly, since every
// zone and rule check in checkImport depends on it never treating
// "clientx" as being within "client".
func TestHasPathPrefix(t *testing.T) {
	tests := []struct {
		relDir string
		prefix string
		want   bool
	}{
		{relDir: "client", prefix: "client", want: true},
		{relDir: "client/cli", prefix: "client", want: true},
		{relDir: "clientx", prefix: "client", want: false},
		{relDir: "client-extra", prefix: "client", want: false},
		{relDir: "serverless", prefix: "server", want: false},
		{relDir: "server", prefix: "server", want: true},
		{relDir: "task", prefix: "task", want: true},
		{relDir: "taskrunner", prefix: "task", want: false},
	}
	for _, tt := range tests {
		got := hasPathPrefix(tt.relDir, tt.prefix)
		if got != tt.want {
			t.Fatalf("hasPathPrefix(%q, %q) = %v, want %v", tt.relDir, tt.prefix, got, tt.want)
		}
	}
}
