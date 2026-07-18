# 🧮 Exercises: Functions and Pointers

Practice the concepts from
[Module 3](../../lessons/03_functions_and_pointers/README.md): multiple return
values, variadic parameters, closures, and pointers.

## 🧩 Tasks

1. `Divide` — divide two floats, returning an error instead of panicking on
   division by zero (multiple return values).
2. `MinMax` — return the minimum and maximum of a variadic list of ints, plus
   an error when called with no arguments.
3. `Sum` — sum a variadic list of ints.
4. `Counter` — return a closure that returns an incrementing count on each
   call, starting at 1.
5. `Accumulator` — return a closure that adds its argument to a running
   total (seeded by `start`) and returns the new total.
6. `Increment` — add 1 to the int pointed to by a pointer.
7. `SwapInts` — swap the values pointed to by two int pointers.

## ▶️ Commands

```bash
go test ./exercises/03_functions_and_pointers/...
go test -run '^$' ./exercises/03_functions_and_pointers
go test ./exercises/03_functions_and_pointers/solution
gofmt -l exercises/03_functions_and_pointers
```

## 📝 Notes

- A variadic parameter (`...int`) is a slice inside the function body; a call
  with no arguments gives a nil slice with length zero, which is safe to range
  over.
- Each call to a closure-returning function creates independent state; two
  counters must not share the same counter.
- A pointer receiver lets a function mutate the caller's variable; a plain
  value parameter cannot.
- Compare with `solution/` only after a genuine attempt.
