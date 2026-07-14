# 📦 07 — Packages and Generics

Build small, reusable generic collection helpers and types to practice type
parameters, custom constraints, and designing a package around an exported
API boundary with unexported internal state.

## ▶️ Workflow

```bash
go test ./exercises/07_packages_and_generics
go test ./exercises/07_packages_and_generics/solution
```

The starter package in this folder (`collections.go`) fails until you
implement every function and method. Compare with `solution/collections.go`
only after a genuine attempt.

## 🧩 Tasks

1. Declare the `Number` constraint (an interface listing the permitted
   underlying numeric types with `~`) and implement
   `Sum[T Number](values []T) T`, returning the zero value of `T` for an
   empty or nil slice.
2. Implement `Map[T, U any](values []T, fn func(T) U) []U`, applying `fn` to
   every element and preserving order. Return an empty, non-nil slice for an
   empty input.
3. Implement `Filter[T any](values []T, predicate func(T) bool) []T`,
   keeping only elements for which `predicate` returns true and preserving
   order.
4. Implement `Reduce[T, U any](values []T, initial U, fn func(U, T) U) U`,
   folding `values` into a single result starting from `initial`.
5. Implement the generic `Stack[T any]` type's `Push`, `Pop`, `Peek`, and
   `Len` methods on top of its unexported `items` field. `Pop` and `Peek`
   must return `(zero value, false)` on an empty stack instead of panicking.
6. Implement the generic `Queue[T any]` type's `Enqueue`, `Dequeue`, and
   `Len` methods on top of its unexported `items` field, preserving
   first-in-first-out order. `Dequeue` must return `(zero value, false)` on
   an empty queue instead of panicking.

## 🔍 What this covers

- Type parameters (`[T any]`) on functions and struct types.
- Custom type constraints built from underlying-type unions (`~int |
  ~float64 | ...`).
- Generic collection types (`Stack[T]`, `Queue[T]`) with an exported method
  API guarding unexported internal storage — a package API boundary even
  within a single file.
- Higher-order generic functions (`Map`, `Filter`, `Reduce`) as an
  alternative to hand-written loops for common transformations.

## ⚠️ Common mistakes

- Forgetting the `~` prefix in a constraint union, which then only accepts
  the exact listed types and rejects named types with that underlying type.
- Returning a `nil` slice from `Map`/`Filter` when the input is empty instead
  of an empty, non-nil slice, which can surprise callers relying on
  `reflect.DeepEqual` or JSON marshaling (`nil` encodes as `null`, `[]T{}`
  encodes as `[]`).
- Exposing a collection's backing slice directly instead of unexported
  storage plus exported methods, which lets callers bypass the type's
  invariants.
- Using a pointer receiver on some methods and a value receiver on others
  for the same generic type, which causes inconsistent mutation behavior.
