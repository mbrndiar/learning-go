# 🚦 Exercises: Control Flow

Practice the concepts from
[Module 2](../../lessons/02_control_flow/README.md): classification, loops, and
search using `if`, `switch`, and `for`.

## 🧩 Tasks

1. `ClassifyNumber` — classify an int as `"negative"`, `"zero"`, or
   `"positive"`.
2. `Grade` — convert a numeric score to a letter grade with a `switch`,
   returning an error for out-of-range input.
3. `SumRange` — sum all integers in an inclusive range using a `for` loop,
   handling a start greater than end.
4. `FizzBuzz` — build a `[]string` for `1..n` using the classic FizzBuzz
   rules.
5. `CountDigits` — count the digits of an int with a loop, handling zero and
   negative numbers.
6. `LinearSearch` — find the index of a target in an unsorted slice, or `-1`.
7. `BinarySearch` — find the index of a target in a sorted slice, or `-1`,
   using the classic halving-interval loop.

## ▶️ Commands

```bash
go test ./exercises/02_control_flow/...
go test -run '^$' ./exercises/02_control_flow
go test ./exercises/02_control_flow/solution
gofmt -l exercises/02_control_flow
```

## 📝 Notes

- Prefer early `return` over deeply nested `if`/`else`.
- `switch` without a condition expression reads like a chain of `if`/`else if`.
- Binary search requires the input to already be sorted ascending; do not sort
  inside the function.
- Compare with `solution/` only after a genuine attempt.
