# 🗂️ Exercises: Collections

Practice the concepts from
[Module 4](../../lessons/04_collections/README.md): slices, maps, `copy`,
sorting, and shared-backing-array awareness.

## 🧩 Tasks

1. `Sum` — sum the elements of an `[]int`.
2. `Unique` — return the distinct elements of an `[]int`, preserving the
   order of first appearance, using a map to track what has been seen.
3. `WordFrequency` — split text on whitespace and count occurrences of each
   word in a `map[string]int`.
4. `MergeCounts` — merge two `map[string]int` count maps, summing values for
   keys present in both, without mutating either input map.
5. `SortDescending` — return a new slice with the elements sorted from
   largest to smallest, without mutating the input slice.
6. `RemoveAt` — return a new slice with the element at an index removed,
   returning an error for an out-of-range index, without mutating the input.
7. `CloneInts` — return an independent copy of an `[]int` using the `copy`
   builtin, so mutating the clone never affects the original backing array.

## ▶️ Commands

```bash
go test ./exercises/04_collections/...
go test -run '^$' ./exercises/04_collections
go test ./exercises/04_collections/solution
gofmt -l exercises/04_collections
```

## 📝 Notes

- Map iteration order is random; never rely on it for output order. `Unique`
  keeps order by iterating the original slice, using the map only for
  membership.
- Slicing (`s[a:b]`) shares the underlying array with `s`; `append` may or
  may not allocate a new array depending on capacity. `copy` is the
  deliberate way to get an independent array.
- Functions that must not mutate their input should build a new slice or map
  rather than sorting or writing in place.
- Compare with `solution/` only after a genuine attempt.
