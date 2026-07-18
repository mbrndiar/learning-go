# 📦 Module 7 — Packages, Modules and Generics

This module zooms out from a single file to a whole program made of
packages, then zooms into the two Go features that let one function body
work across many types: generics and range-over-func iterators.

**Prerequisites:** Modules 1–6 (Basics through Errors, Files, Directories, JSON
and Time).

## 🧠 Mental model

A **package** is a directory of source files sharing one namespace; a
**module** is the versioned unit — rooted at `go.mod` — that can contain
many packages and is what `go get` resolves and what `go.sum` checksums.
Only names starting with an uppercase letter are exported, so a package's
real API is its exported identifiers, not everything in the directory;
`internal/` extends that boundary with a compiler-enforced rule instead of
a naming convention.

Generics let a function or type declare a relationship between types at
compile time — "these parameters and this return value are the same type
`T`" — without giving up static type checking or paying reflection's
runtime cost. A **constraint** is the interface that lists which operations
a type parameter must support; `any` guarantees no comparison, arithmetic, or
ordering operations because it also includes non-comparable types such as
slices and maps. Those operations require a narrower constraint such as
`comparable`, `cmp.Ordered`, or a custom one.

Range-over-func (`iter.Seq`) lets a plain function be the thing you `range`
over: it receives a `yield` callback and calls it once per produced value,
stopping as soon as `yield` returns `false`. That single contract is what
lets custom iterators compose with `for range` like any built-in sequence.

Generics and iterators both add a layer of indirection a reader must
mentally unwrap — write the concrete version first, and generalize only
once a second or third real, differently-typed caller justifies the cost.

## 🎯 Learning goals

By the end of this module you will be able to:

- organize code into packages with a clear exported API;
- explain how a module path and a directory path combine into an import
  path;
- use an `internal` directory to enforce a package boundary the compiler
  checks, not just a convention;
- write generic functions with type parameters and constraints, including
  the standard library's `cmp.Ordered`;
- write and consume custom iterator functions with range-over-func, and use
  the standard library's iterator-returning helpers; and
- judge when generalizing code (with generics or otherwise) is worth its
  cost, and when duplication is the better choice.

## 📦 Lessons

1. [`01_package_organization/`](01_package_organization/) — package
   organization, exported names, the module/import path relationship, and an
   `internal/` package the compiler prevents other trees from importing.
2. [`02_generic_helpers/`](02_generic_helpers/) — type parameters,
   constraints (`any`, a custom `Number` constraint, `cmp.Ordered`), and
   generic `Map`/`Filter`/`Reduce`/`Sum`/`Max` helpers.
3. [`03_iterators_range_over_func/`](03_iterators_range_over_func/) — writing
   `iter.Seq`-based iterator functions, ranging over them directly,
   composing iterators, and the standard library's `slices.Values`/
   `slices.Collect`.
4. [`04_avoiding_premature_abstraction/`](04_avoiding_premature_abstraction/)
   — duplicated code versus a generic helper versus an over-generalized
   combinator, and a rule of thumb for choosing between them.

## ▶️ How to run a lesson

From the repository root:

```bash
go run ./lessons/07_packages_and_generics/01_package_organization
go run ./lessons/07_packages_and_generics/02_generic_helpers
go run ./lessons/07_packages_and_generics/03_iterators_range_over_func
go run ./lessons/07_packages_and_generics/04_avoiding_premature_abstraction
```

Lesson 1 contains two supporting packages, `catalog/` and `internal/pricing/`,
that `main.go` imports; read all three files together.

**Experiment:** in `03_iterators_range_over_func/main.go`, add a `break`
partway through a `for range` loop over a custom iterator, and add a print
statement right before the iterator's `yield` call — run it and observe that
`yield` returning `false` stops the iterator from producing further values.

## 🧩 Matching exercises

[`exercises/07_packages_and_generics/`](../../exercises/07_packages_and_generics/README.md)
— generic constraints, `Map`/`Filter`/`Reduce`, and generic `Stack`/`Queue`
types.

## 🧰 Module and dependency workflow

A **package** is one directory of Go source. A **module** is the versioned unit
rooted at a `go.mod` file and can contain many packages. The module path plus a
package's directory determines its import path.

For a new standalone project, the practical lifecycle is:

```bash
go mod init example.com/myapp              # create go.mod once
go get example.com/dependency@v1.2.3       # add or update a dependency
go mod tidy                                # match go.mod/go.sum to imports
go mod download                            # pre-download declared modules
go list -m all                             # inspect selected module versions
```

`go.sum` records cryptographic checksums for downloaded module content; it is
not a lockfile that pins every selected version. Commit both `go.mod` and
`go.sum`, prefer explicit versions in reproducible automation, and review their
diff whenever dependencies change. See the official
[Go module reference](https://go.dev/ref/mod).

## 🚧 Common mistakes

- **Confusing the module path with a package's directory path.** An import
  path is the module path from `go.mod` (`github.com/mbrndiar/learning-go`)
  joined with the package's directory path. Renaming a directory changes
  every import path that points into it.
- **Assuming `internal` is just a naming convention.** It is enforced by the
  compiler: a package under a path segment named `internal` can only be
  imported by code rooted at the parent of that `internal` directory.
  Importing it from anywhere else fails the build.
- **Exporting more than callers need.** Prefer the smallest exported surface
  (as in `catalog.Product`/`catalog.New`) and keep implementation details
  either unexported or behind `internal/`.
- **Reaching for `any` when a real constraint exists.** `Sum` in lesson 2
  needs arithmetic, so it is constrained to `Number` (or the standard
  `cmp.Ordered` for comparisons), not `any`; an unconstrained type parameter
  cannot be added, compared with `<`, or ranged over numerically.
- **Forgetting that `yield` can return false.** A custom iterator must stop
  producing values as soon as `yield` returns false (typically because the
  caller's range loop hit a `break`); ignoring that return value can keep
  computing work nobody will see.
- **Generalizing before a third real caller exists.** The `aggregate`
  function in lesson 4 is not wrong, but it demonstrates the tax an early
  abstraction places on every reader. Two concrete implementations are
  usually cheaper to maintain than one wrong abstraction.

## ❓ Review questions

1. How is an import path constructed from a module's `go.mod` declaration
   and a package's location on disk?
2. What exactly does putting a package under an `internal/` directory
   prevent, and who enforces it?
3. Why does `Sum` need a constraint like `Number` while `Map` and `Filter`
   can stay generic over `any`?
4. What does returning `false` from a `yield` callback mean for an iterator
   function, and when does that happen during a `range` loop?
5. In lesson 4, why is duplicating `sumInts` and `sumFloats` initially
   preferable to writing a generic `Sum` right away, and what changed once a
   third case appeared?
6. What cost does the `aggregate` combinator in lesson 4 impose on callers
   compared with `Sum`, even though both are technically correct?
7. What is the difference between a package and a module, and what roles do
   `go get`, `go mod tidy`, and `go.sum` play when dependencies change?

## 📚 References

- [Go Modules Reference](https://go.dev/ref/mod)
- [Tutorial: Create a Go module](https://go.dev/doc/tutorial/create-module)
- [An Introduction To Generics](https://go.dev/blog/intro-generics)
- [The Go Programming Language Specification: Type parameter declarations](https://go.dev/ref/spec#Type_parameter_declarations)
- [Range over function types (Go 1.23)](https://go.dev/blog/range-functions)
