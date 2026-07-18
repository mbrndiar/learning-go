# 🔀 Module 2: Control Flow

This module covers how Go directs the order of execution: `if` with an
optional initialization clause, `switch` in its several forms, every shape of
`for` (Go's only loop keyword), and how `range` interacts with mutation.

**Prerequisite:** [Module 1 — Basics](../01_basics/README.md).

## 🧠 Mental model

Go has one decision keyword, `if`, and one loop keyword, `for` — there is no
`while`, `do`/`until`, or ternary operator. `if` always requires a `bool`
condition (no truthy numbers or strings), and it may carry an `init;
condition` form that scopes a helper variable — commonly an error — to just
that chain, so it cannot leak into surrounding code. `switch` is a decision
tool, not a jump table: choose an expression `switch` for one variable
against several values, a condition-less `switch` as a readable
`if`/`else if` chain, and a type `switch` to branch on an interface's
dynamic type; cases do not fall through unless you say `fallthrough`.

`for` covers every loop shape (three-part, condition-only, infinite, and
`range`) through one keyword and different contracts for what runs when.
`range` over a slice, map, or string produces a **copy** of each value, not
a reference into the original — mutating the loop variable never mutates
the source, and (before appending logic runs) `range` fixes the slice's
length at the start, so elements appended during the loop are never visited.
Map `range` order is unspecified and may vary from one iteration to the next;
string `range` yields runes at their byte offset, not sequential indices,
because runes can span multiple bytes.

Labels exist only for one problem: reaching an outer loop from inside a
nested one, since a bare `break`/`continue` only affects the innermost
loop. Reach for a label when — and only when — you have genuine nesting to
escape.

## 🎯 Learning objectives

By the end of this module you will be able to:

- use `if`'s initialization clause to scope a helper variable (commonly an
  error) to a single conditional chain;
- choose between an expression `switch`, a condition-less `switch`, and a
  type `switch`, and know when `fallthrough` is appropriate;
- write every shape of Go's `for` loop and iterate over slices, strings,
  maps, and integers with `range`;
- use `break`, `continue`, and labeled loops correctly in nested loops, and
  avoid the mutation mistakes that come from ranging over a copy.

## 📖 Lessons

1. [`01_if_and_init/main.go`](01_if_and_init/main.go) — `if`/`else if`/`else`,
   the `if init; condition` form, and variable scoping.
2. [`02_switch/main.go`](02_switch/main.go) — expression switches,
   condition-less switches, `fallthrough`, and type switches.
3. [`03_for_and_range/main.go`](03_for_and_range/main.go) — the three-part
   `for`, condition-only `for`, infinite `for`, and `range` over slices,
   strings, and integers.
4. [`04_break_continue_labels/main.go`](04_break_continue_labels/main.go) —
   `break`, `continue`, labeled loops for nested breaks/continues, and
   mutation cautions when ranging over slices.

## ▶️ Running the lessons

Run any lesson from the repository root:

```bash
go run ./lessons/02_control_flow/01_if_and_init
go run ./lessons/02_control_flow/02_switch
go run ./lessons/02_control_flow/03_for_and_range
go run ./lessons/02_control_flow/04_break_continue_labels
```

Or check that every lesson in this module compiles at once:

```bash
go build ./lessons/02_control_flow/...
```

**Experiment:** in `03_for_and_range/main.go`, predict which of the two existing
loops over `numbers` changes the original slice: the loop that multiplies the
range value `n`, or the loop that writes through `numbers[i]`. Run it, then swap
their order and explain why the first printed result changes but the ownership
rule does not.

## 🧩 Matching exercises

[`exercises/02_control_flow/`](../../exercises/02_control_flow/README.md) —
branching, switch, and loop/range helpers.

## 💡 Concepts covered

- `if`/`else if`/`else` chains and the `if init; condition { }` form, whose
  declared variables are scoped only to that chain.
- Conditions must have type `bool`. Go has no implicit truthiness for numbers,
  strings, pointers, slices, or maps; compare explicitly with `0`, `""`, or
  `nil`.
- Variable shadowing: a variable declared inside an `if`'s init clause with
  `:=` is a new variable, even if it shares a name with an outer variable.
- `switch` without `fallthrough` by default (each case exits automatically),
  multi-value `case` lists, `switch` with an init statement, condition-less
  `switch` as a readable alternative to long `if`/`else if` chains, explicit
  `fallthrough`, and type switches (`switch v := x.(type)`).
- The four shapes of `for`: three-part (`init; condition; post`),
  condition-only (a "while" loop), infinite (`for { }`) with `break`, and
  `range`-based.
- `range` over a slice yields `(index, value)`, where `value` is a copy of
  the element; `range` over a string yields `(byteIndex, rune)`; `range` over
  an integer (Go 1.22+) yields `0` through `n-1`.
- `break` and `continue`, including that an unlabeled `break`/`continue`
  affects only the innermost loop.
- Labels placed before a loop, letting `break label` or `continue label`
  target an outer loop directly from a nested one.
- Why mutating a `range` value does not change the original slice, and why
  appending to a slice while ranging over it does not extend the loop.

## ⚠️ Common mistakes

- Expecting a `switch` case to fall through to the next case, as in C or
  Java. Go's cases exit automatically; use the explicit `fallthrough`
  keyword if you truly want that behavior.
- Expecting an `if init; condition` variable to be visible after the
  `if`/`else` chain ends. It is scoped only to that chain.
- Writing `if count` or `if name` as in a language with truthy/falsy values.
  Go requires a Boolean condition such as `count != 0` or `name != ""`.
- Mutating the loop variable from `for _, v := range slice` and expecting
  the original slice to change. `v` is a copy; index into the slice
  (`slice[i]`) to mutate it in place.
- Using a bare `break` inside nested loops and expecting it to exit every
  loop. It only exits the innermost one; use a label for an outer loop.
- Appending to a slice while ranging over that same slice and expecting the
  loop to visit the new elements. `range` captures the length once, before
  the loop starts.

## ❓ Review questions

1. Why does a variable declared in an `if`'s init clause disappear after the
   `else` branch, and what problem does that scoping solve?
2. What is the difference between a condition-less `switch` and a `switch`
   with an explicit boolean expression like `switch x > 0 { }`?
3. Which shape of `for` would you choose to loop "while a condition holds",
   and which for "exactly n times"?
4. Why does `break` inside a nested loop sometimes require a label to have
   the effect you want?
5. If you range over `[]int{1, 2, 3}` with `for _, v := range numbers` and
   set `v = 0` inside the loop, what does `numbers` look like afterward, and
   why?
6. Why does `if count { ... }` fail to compile in Go, and what explicit
   condition should replace it?

## 📚 References

- [The Go Programming Language Specification: Statements](https://go.dev/ref/spec#Statements)
- [A Tour of Go: Flow control statements](https://go.dev/tour/flowcontrol/1)
- [Go 1.22 Release Notes: range over integers](https://go.dev/doc/go1.22#language)

Previous: [Module 1 — Basics](../01_basics/README.md). Next:
[Module 3 — Functions and Pointers](../03_functions_and_pointers/README.md).
