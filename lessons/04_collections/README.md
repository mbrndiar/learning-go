# 🗂️ Module 4: Collections

This module covers Go's built-in collection types: arrays, slices, and maps.
The emphasis is on the parts that surprise newcomers — shared backing
arrays, `append`'s growth behavior, and map iteration order — because
misunderstanding them causes real bugs, not just stylistic issues.

## 🎯 Learning objectives

By the end of this module you will be able to:

- explain the difference between an array and a slice, and describe a
  slice's length, capacity, and how `append` grows it;
- explain why two slices can share the same backing array, predict when a
  mutation through one slice is visible through another, and use `copy` to
  make an independent duplicate;
- use maps for membership checks and deletion, and explain why map
  iteration order is randomized and how to get deterministic output anyway;
- sort slices with `slices.Sort`/`slices.SortFunc` and combine the `slices`
  and `maps` packages to iterate a map in a predictable order.

## 📖 Lessons

1. [`01_arrays_and_slices/main.go`](01_arrays_and_slices/main.go) — fixed-size
   arrays versus slices, `len`/`cap`, `make`, and how `append` grows a
   slice.
2. [`02_slice_sharing_and_copy/main.go`](02_slice_sharing_and_copy/main.go) —
   shared backing arrays, surprising `append` aliasing, `copy`, and full
   slice expressions.
3. [`03_maps/main.go`](03_maps/main.go) — map creation, the "comma ok"
   membership idiom, `delete`, and randomized iteration order.
4. [`04_sorting/main.go`](04_sorting/main.go) — `slices.Sort`,
   `slices.SortFunc`, `slices.Reverse`, and combining `slices`/`maps` helpers
   for deterministic map iteration.

## ▶️ Running the lessons

Run any lesson from the repository root:

```bash
go run ./lessons/04_collections/01_arrays_and_slices
go run ./lessons/04_collections/02_slice_sharing_and_copy
go run ./lessons/04_collections/03_maps
go run ./lessons/04_collections/04_sorting
```

Or check that every lesson in this module compiles at once:

```bash
go build ./lessons/04_collections/...
```

## 💡 Concepts covered

- Arrays: a fixed size that is part of the type (`[3]int` and `[4]int` are
  different types), and value semantics (assigning or passing an array
  copies every element).
- Slices: a `(pointer, length, capacity)` header describing a view over a
  backing array; `len`/`cap`; `make([]T, length, capacity)`; and how a `nil`
  slice behaves like an empty one and is safe to `append` to.
- `append`'s growth behavior: while capacity allows it, `append` reuses the
  existing backing array; once length would exceed capacity, it allocates a
  new, larger array and copies the elements over.
- Slicing an existing slice (`s[low:high]`) shares the same backing array as
  the original, so a mutation through one is visible through the other.
- How that sharing interacts with `append`: appending to a sub-slice that
  still has spare capacity can silently overwrite data that a sibling slice
  is also viewing.
- `copy(dst, src)` for making an independent duplicate, and full slice
  expressions (`s[low:high:max]`) for capping a sub-slice's capacity so its
  own `append` calls cannot overwrite a sibling's data.
- Maps: creation with a literal or `make`, the `nil` map (safe to read, but
  panics on write), the "comma ok" idiom (`value, ok := m[key]`) to
  distinguish a missing key from a present key holding a zero value, and
  `delete` (a no-op for a missing key).
- Map iteration order is randomized on purpose so no code accidentally
  depends on an unspecified order; collecting and sorting the keys is the
  standard way to get deterministic, repeatable output.
- Using `map[T]struct{}` as an idiomatic, zero-overhead set.
- Sorting with `slices.Sort` (ordered basic types), `slices.SortFunc` (custom
  or multi-field comparisons, including tie-breaking for determinism),
  `slices.Reverse`, `slices.IsSorted`, and `slices.Sorted(maps.Keys(m))` for
  a deterministic, sorted view of a map's keys.

## ⚠️ Common mistakes

- Forgetting that `s[low:high]` shares memory with the original slice, then
  being surprised when a mutation through the sub-slice changes the
  original (or vice versa).
- Appending to a sub-slice that still has spare capacity and unexpectedly
  overwriting a sibling slice's data. Use a full slice expression
  (`s[low:high:max]`) or `copy` when independence matters.
- Writing to a `nil` map before initializing it with `make` or a literal,
  which panics at runtime (reading a `nil` map is safe and returns zero
  values).
- Reading a map value without the second "comma ok" return and treating a
  zero value as if the key were actually present.
- Printing a map directly and relying on the order of the entries in
  another context (such as a test or a log parser); iteration order is not
  guaranteed, only `fmt`'s own printing of a map value is sorted by key.

## ❓ Review questions

1. What is the difference between an array's and a slice's assignment
   semantics?
2. When does `append` reuse a slice's existing backing array, and when does
   it allocate a new one?
3. Give an example where appending to one slice unexpectedly changes what
   another slice sees, and explain why full slice expressions or `copy`
   prevent it.
4. Why does Go randomize map iteration order, and what is the idiomatic way
   to print a map's entries in a stable, repeatable order?
5. When would you reach for `slices.SortFunc` instead of `slices.Sort`?

Previous:
[Module 3 — Functions and Pointers](../03_functions_and_pointers/README.md).
Next:
[Module 5 — Structs, Methods and Interfaces](../05_structs_methods_interfaces/README.md).
