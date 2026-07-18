# 🧩 Module 3: Functions and Pointers

This module covers how Go organizes behavior into functions and how it
handles memory addresses. It builds toward one of Go's most important ideas:
the difference between passing a copy of a value and passing a pointer to
it — a distinction that motivates pointer receivers on methods later in the
course.

**Prerequisites:** [Module 1 — Basics](../01_basics/README.md) and
[Module 2 — Control Flow](../02_control_flow/README.md).

## 🧠 Mental model

A function's signature is a contract: the parameter types and the return
types are what a caller can rely on. Go has no exceptions, so a function
that can fail returns an extra `error` value alongside its result, and
convention (not the compiler) says to check it before trusting the other
return values.

Functions themselves are values — you can hold one in a variable, put it in
a map, or return one from another function. A **closure** is a function
literal that refers to variables from its enclosing scope. It retains those
variables themselves, not merely a snapshot of each value at creation time;
each call to a factory function creates a fresh set of captured variables,
which is why two calls produce two independent counters.

Recursion trades a straightforward problem decomposition (a base case plus
a smaller sub-problem) for stack frames and, without memoization, possibly
exponential repeated work — prefer an iterative or memoized version once
performance matters.

Go passes every argument **by value**: a function parameter is a copy, so
reassigning it or mutating a plain `int`/`string`/`struct` parameter never
changes the caller's variable. This is not "pass by reference" — it is
always a copy — but a copied slice, map, pointer, channel, or function value
still holds the same header or address, so mutating the data it *points to*
(as opposed to the header itself) is visible to the caller. That distinction
is exactly what Module 4 explores in depth for slices and maps.

## 🎯 Learning objectives

By the end of this module you will be able to:

- write functions with multiple parameters, multiple return values, and
  named returns, and use an error return instead of throwing an exception;
- write variadic functions, pass functions as values, and write closures
  that capture variables from an enclosing scope;
- write and reason about recursive functions, including their base case and
  performance trade-offs versus an iterative solution;
- explain Go's value-by-default parameter passing, and use pointers to let a
  function modify a caller's variable.

## 📖 Lessons

1. [`01_functions_basics/main.go`](01_functions_basics/main.go) — parameters,
   multiple return values, returning an error, and named returns.
2. [`02_variadic_closures/main.go`](02_variadic_closures/main.go) — variadic
   parameters, functions as first-class values, and closures.
3. [`03_recursion/main.go`](03_recursion/main.go) — recursive functions,
   base cases, and naive recursion versus an iterative equivalent.
4. [`04_pointers/main.go`](04_pointers/main.go) — value semantics, the `&`
   and `*` operators, `nil` pointers, and why mutable state motivates
   pointer receivers.

## ▶️ Running the lessons

Run any lesson from the repository root:

```bash
go run ./lessons/03_functions_and_pointers/01_functions_basics
go run ./lessons/03_functions_and_pointers/02_variadic_closures
go run ./lessons/03_functions_and_pointers/03_recursion
go run ./lessons/03_functions_and_pointers/04_pointers
```

Or check that every lesson in this module compiles at once:

```bash
go build ./lessons/03_functions_and_pointers/...
```

**Experiment:** in `04_pointers/main.go`, write two versions of a `swap`
function — one taking `int` parameters and one taking `*int` parameters —
call each from `main`, and print the caller's variables afterward to see
which one actually swaps them.

## 🧩 Matching exercises

[`exercises/03_functions_and_pointers/`](../../exercises/03_functions_and_pointers/README.md)
— multi-return functions, closures, recursion, and pointer mutation.

## 💡 Concepts covered

- Function declarations with multiple parameters (including the shared-type
  shorthand `func add(a, b int) int`) and multiple return values.
- Returning `(value, error)` pairs and checking the error immediately after
  the call, instead of relying on exceptions.
- Named return values, and how a bare `return` sends back their current
  values — useful for documenting what each return value means.
- The blank identifier `_` for discarding a return value you do not need.
- Variadic parameters (`func sum(numbers ...int) int`) and spreading a slice
  into one with `values...`.
- Functions as values: storing them in variables and maps, and passing them
  as arguments.
- Closures: anonymous functions that capture variables from their enclosing
  scope, including why each call to a factory function produces an
  independent captured variable, and how Go 1.22+ gives each loop iteration
  its own copy of the loop variable.
- Recursion: a base case that stops the recursion, a recursive case that
  solves a smaller sub-problem, and the performance cost of naive recursion
  (recomputing the same sub-problems) versus an iterative or memoized
  approach.
- Value semantics: every function parameter receives a copy. Reassigning that
  local parameter never changes the caller's variable, but copied slices,
  maps, pointers, channels, and functions can still refer to shared data.
  Module 4 develops that distinction for slices and maps.
- Pointers: `&x` takes the address of `x`; `*p` dereferences a pointer to
  read or write the value it points to; a pointer's zero value is `nil`, and
  dereferencing a `nil` pointer panics.
- Why pointer parameters are necessary to let a function modify a caller's
  variable (for example, an in-place `swap`), which is the same reasoning
  behind pointer receivers on methods in Module 5.

## ⚠️ Common mistakes

- Expecting a function to modify a plain (non-pointer) parameter and have
  the caller see the change. Go passes arguments by value; you must pass a
  pointer to mutate the original.
- Dereferencing a pointer without checking whether it is `nil` first, which
  panics at runtime.
- Writing a recursive function without a reachable base case, or one whose
  argument never actually approaches the base case, causing a stack
  overflow.
- Assuming naive recursive Fibonacci is fast. It repeats the same
  sub-computations exponentially many times; prefer an iterative or
  memoized version for real workloads.
- Forgetting that named return values still need explicit `return`
  statements in most functions; a bare `return` only makes sense once the
  named variables already hold the right values.

## ❓ Review questions

1. Why does `func increment(n int) { n++ }` fail to change the caller's
   variable, while `func increment(n *int) { *n++ }` succeeds?
2. What is the zero value of a pointer, and what happens if you dereference
   it without checking?
3. How does a variadic function like `func sum(numbers ...int) int` differ,
   from the caller's perspective, from a function that takes a `[]int`
   parameter directly?
4. Why does calling a "counter factory" function twice produce two
   independent counters instead of one shared counter?
5. What is the base case of a recursive function, and what goes wrong if a
   recursive function never reaches it?

## 📚 References

- [The Go Programming Language Specification: Function declarations](https://go.dev/ref/spec#Function_declarations)
- [A Tour of Go: Pointers](https://go.dev/tour/moretypes/1)
- [Go FAQ: Pointers and Allocation](https://go.dev/doc/faq#Pointers)
- [Effective Go: Errors are values](https://go.dev/blog/errors-are-values)

Previous: [Module 2 — Control Flow](../02_control_flow/README.md). Next:
[Module 4 — Collections](../04_collections/README.md).
