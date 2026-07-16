# ✅ Course Verification and CI

Run commands from the repository root. The canonical automation is
[`.github/workflows/course.yml`](../.github/workflows/course.yml), which tests
Go 1.25.x and 1.26.x. The `.x` selectors intentionally resolve the latest
security patch in each supported release line.

## Why `go test ./...` fails in a learner checkout

The twelve packages directly under `exercises/` are learner starters. Their
tests describe the work and intentionally fail while the TODO implementations
remain. Therefore a raw:

```bash
go test ./...
```

is useful after completing every exercise, but it is not the clean-checkout
health command. The capstone starters are different: their small harnesses
recognize the explicit `ErrNotImplemented`/HTTP 501 placeholders, while CI also
type-checks every starter package with `go test -run '^$'`.

CI keeps learner intent separate from repository health:

```bash
go test ./lessons/...
while IFS= read -r package; do
  [[ -z "$package" ]] && continue
  go run "$package"
done < <(
  go list -f '{{if eq .Name "main"}}{{.ImportPath}}{{end}}' ./lessons/... |
    sed '/^$/d'
)

go list ./exercises/... |
  grep -v '/solution$' |
  xargs -r go test -run '^$'
go list ./exercises/... |
  grep '/solution$' |
  xargs -r go test

go test -run '^$' \
  ./capstones/comparative/starter/... \
  ./capstones/idiomatic/starter/...
go test \
  ./capstones/comparative/starter/kvstore \
  ./capstones/comparative/solution/... \
  ./capstones/comparative/tests/... \
  ./capstones/idiomatic/starter/monitor \
  ./capstones/idiomatic/solution/... \
  ./capstones/idiomatic/tests/... \
  ./capstones
```

## Module, formatting, vet, and links

```bash
go mod download
go mod tidy
git diff --exit-code -- go.mod go.sum

test -z "$(gofmt -l .)"
go vet ./...
go run ./tools/checklinks
```

The link checker validates repository-local Markdown targets. External URLs and
heading fragments are intentionally outside its offline scope.

## Race detector

The Go 1.26 CI job race-checks the code that owns concurrency, storage, HTTP, or
process boundaries:

```bash
go test -race \
  ./capstones/comparative/tests/... \
  ./capstones/idiomatic/solution/monitor/... \
  ./capstones/idiomatic/tests/...

mkdir -p capstones/comparative/.conformance/race
go build -race \
  -o capstones/comparative/.conformance/race/kvstore \
  ./capstones/comparative/solution/kvstore/cmd/kvstore
COMPARATIVE_KV_PROGRAM="$PWD/capstones/comparative/.conformance/race/kvstore" \
  go test -race -count=2 ./capstones/comparative/solution/kvstore
rm -rf capstones/comparative/.conformance/race

go test -race ./lessons/10_concurrency/...
go test -race ./lessons/11_sql_and_sqlite/...
go test -race ./lessons/12_rest_apis_and_clients/...
go test -race ./exercises/10_concurrency/solution
go test -race ./exercises/11_sql_and_sqlite/solution
go test -race ./exercises/12_rest_apis_and_clients/solution
```

It does not use `go test -race ./...`, because that would run the intentionally
failing exercise starter tests and would obscure the defined race surface.

## Coverage

CI enforces 85% for the idiomatic monitor and 75% for the comparative
command/process implementation:

```bash
go test -coverprofile=idiomatic-coverage.out \
  ./capstones/idiomatic/solution/monitor/...
bash scripts/check-coverage.sh idiomatic-coverage.out 85
```

The comparative capstone is measured through its real executable boundary:

```bash
mkdir -p \
  capstones/comparative/.coverage/bin \
  capstones/comparative/.coverage/data
go build -cover \
  -coverpkg=./capstones/comparative/solution/... \
  -o capstones/comparative/.coverage/bin/kvstore \
  ./capstones/comparative/solution/kvstore/cmd/kvstore
COMPARATIVE_KV_PROGRAM="$PWD/capstones/comparative/.coverage/bin/kvstore" \
GOCOVERDIR="$PWD/capstones/comparative/.coverage/data" \
  go test ./capstones/comparative/solution/kvstore
coverage_inputs=$(
  find capstones/comparative/.coverage/data \
    -mindepth 1 -maxdepth 1 -type d -printf '%p,' |
    sed 's/,$//'
)
go tool covdata textfmt \
  -i="$coverage_inputs" \
  -o=comparative-coverage.out
bash scripts/check-coverage.sh comparative-coverage.out 75
rm -rf capstones/comparative/.coverage
```

The comparative recipe uses GNU `find` as provided by the Ubuntu CI runner.

## Fuzz, staticcheck, and vulnerability gates

CI performs bounded fuzz smoke runs so fuzz targets are continuously executable:

```bash
go test -run '^$' \
  -fuzz '^FuzzParseKeyValue$' \
  -fuzztime=2s \
  ./lessons/08_testing/08_fuzzing
go test -run '^$' \
  -fuzz '^FuzzLoadConfig$' \
  -fuzztime=2s \
  ./capstones/idiomatic/solution/monitor/domain
```

Install the same pinned analyzers as CI, then analyze completed teaching and
reference packages rather than intentional starter implementations:

```bash
go install honnef.co/go/tools/cmd/staticcheck@v0.7.0
{
  go list ./lessons/...
  go list ./exercises/... | grep '/solution$'
  go list ./capstones/comparative/solution/...
  go list ./capstones/comparative/tests/...
  go list ./capstones/idiomatic/solution/...
  go list ./capstones/idiomatic/tests/...
  go list ./capstones/testsupport
  go list ./tools/...
} | xargs "$(go env GOPATH)/bin/staticcheck"

go install golang.org/x/vuln/cmd/govulncheck@v1.6.0
"$(go env GOPATH)/bin/govulncheck" ./...
```

`go vet ./...` and `govulncheck ./...` can inspect the whole buildable module.
Staticcheck deliberately follows the completed-code package list above so
learner placeholders do not become the repository's style baseline.
`govulncheck` also scans the Go standard library, so run it with a current patch
toolchain; an unpatched base release can correctly fail even when dependencies
are clean.
