# 📦 Module 7 — Packages, Modules and Generics

This module zooms out from a single file to a whole program made of
packages, then zooms into the two Go features that let one function body
work across many types: generics and range-over-func iterators.

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
