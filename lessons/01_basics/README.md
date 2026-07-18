# 🌱 Module 1: Basics

This module is the on-ramp to Go. You will write your first program, learn
how Go represents values and types, work through every operator you will use
daily, and understand how Go stores text under the hood. Every lesson is a
runnable, self-contained `package main`.

**Prerequisites:** none — start here.

## 🧠 Mental model

A Go **program** is one or more **packages** compiled together. Executable
statements live inside functions; after package initialization, a command
enters `func main` in `package main`. There is no interpreter executing the
source line by line — `go run`/`go build` compile the whole package first, so a
type error anywhere prevents the command from starting.

Every value has a type, and every variable has one static type fixed when it is
declared — this is what "statically typed" means in practice. A variable you
never assign still holds a well-defined
**zero value** (`0`, `""`, `false`, `nil`, ...) rather than garbage or
"undefined", so a freshly declared value is always safe to read. Because
types are fixed, Go never silently converts one numeric type into another
the way some languages do; you convert explicitly with `T(x)`, and that
conversion is the point where truncation or overflow can happen.

Choosing a numeric representation is a choice about exactness: integers store
whole numbers exactly within their fixed range but can overflow, while
`float64` trades exactness for range and cannot represent most decimal
fractions precisely — after assigning `0.1`, `0.2`, and `0.3` to `float64`
variables, adding the first two need not compare equal to the third. (Untyped
constant arithmetic is evaluated more precisely until a concrete type is
required.) Use integers or a fixed-scale integer such as cents whenever exact
values matter, and reserve floats for measurements that already carry rounding
error.

Strings are read-only, UTF-8-encoded byte sequences, not arrays of
characters. `len(s)` counts bytes; a **rune** is one Unicode code point,
which may itself be more than one byte and still not match what a person
perceives as one printed "character" (an emoji or accented letter can span
several runes). Ranging over a string or converting to `[]rune` walks
code points, not bytes — reach for that boundary whenever you index or
count by "character".

## 🎯 Learning objectives

By the end of this module you will be able to:

- explain what `package main` and `func main` mean, and use `fmt` to print
  and format output;
- declare variables and constants, describe the zero value of every basic
  type, and convert between numeric types explicitly and safely;
- predict the result of arithmetic, comparison, logical, and bitwise
  operators, including integer division and short-circuit evaluation;
- explain the difference between a byte, a rune, and a "character", and
  process Unicode text correctly instead of assuming one byte per character.

## 📖 Lessons

1. [`01_hello_world/main.go`](01_hello_world/main.go) — `package main`,
   `func main`, and the `fmt` package: `Println`, `Printf`, `Sprintf`, and
   the `%v`/`%T` verbs.
2. [`02_variables_and_types/main.go`](02_variables_and_types/main.go) —
   `var` versus `:=`, zero values, Go's static typing, and explicit
   conversions (including truncation and overflow).
3. [`03_operators/main.go`](03_operators/main.go) — arithmetic, comparison,
   logical (with short-circuiting), bitwise, and compound-assignment
   operators, plus operator precedence.
4. [`04_strings_bytes_runes/main.go`](04_strings_bytes_runes/main.go) —
   strings as immutable UTF-8 byte sequences, `len` versus rune count,
   ranging over strings, and the `strings`/`unicode`/`unicode/utf8`
   packages.

## ▶️ Running the lessons

Run any lesson from the repository root:

```bash
go run ./lessons/01_basics/01_hello_world
go run ./lessons/01_basics/02_variables_and_types
go run ./lessons/01_basics/03_operators
go run ./lessons/01_basics/04_strings_bytes_runes
```

Or check that every lesson in this module compiles at once:

```bash
go build ./lessons/01_basics/...
```

**Experiment:** in `04_strings_bytes_runes/main.go`, print both `len(s)` and
`utf8.RuneCountInString(s)` for a string containing an emoji or an accented
letter, then predict which number is larger before you run it.

## 🧩 Matching exercises

[`exercises/01_basics/`](../../exercises/01_basics/README.md) — conversions,
string/rune helpers, and a default-fallback parser.

## 💡 Concepts covered

- Program structure: `package main`, `func main`, and imports.
- Output and formatting with `fmt.Println`, `fmt.Printf`, `fmt.Sprintf`, and
  verbs (`%v`, `%T`, `%d`, `%s`, `%t`, `%q`, `%b`, `%.4f`).
- Declarations with `var` and `:=`, constants with `const`, and grouped
  `const`/`var` blocks.
- Zero values for numbers, strings, booleans, and reference-like types
  (pointers, slices, maps, channels, functions, interfaces are `nil`).
- Static typing and explicit conversion, including truncation
  (`int(floatValue)`) and overflow when converting to a smaller integer type
  (`byte`).
- Binary floating-point as an approximation for most decimal fractions,
  tolerance-based comparison, `NaN`/infinity checks, and why exact decimal
  accounting should use a different representation such as fixed-scale
  integers.
- Arithmetic (`+ - * / %`), comparison (`== != < <= > >=`), logical
  (`&& ||`, with short-circuit evaluation), bitwise (`& | ^ &^ << >>`), and
  compound assignment (`+= *=`) operators.
- Operator precedence and using parentheses to make intent explicit.
- Strings as UTF-8-encoded, immutable byte sequences; the difference between
  `len(s)` (bytes) and `utf8.RuneCountInString(s)` (Unicode code points, not
  necessarily user-perceived characters); converting between `string`,
  `[]byte`, and `[]rune`; and classifying runes with the `unicode` package.

## ⚠️ Common mistakes

- Assuming `len(s)` counts characters. It counts bytes; convert to `[]rune` or
  use `utf8.RuneCountInString` to work with Unicode code points. A visible
  grapheme such as an emoji sequence can still contain several runes.
- Expecting `/` between two integers to produce a fractional result. Integer
  division truncates; convert at least one operand to `float64` first.
- Forgetting that Go requires an explicit conversion between numeric types,
  even between `int` and `float64`. This is enforced at compile time and
  prevents accidental precision loss.
- Assuming a smaller integer type (like `byte`) can hold any value assigned
  to it. Converting a value that does not fit wraps around silently instead
  of raising an error.
- Comparing arbitrary computed `float64` values with `==`, or using binary
  floating point for exact decimal accounting. Choose a domain-specific
  tolerance for approximate measurements and a decimal/fixed-scale model when
  exact decimal values are required.
- Trying to modify a string in place, such as `s[0] = 'H'`. Strings are
  immutable; build a new string instead (via `strings` functions, a
  `[]byte`/`[]rune` conversion, or `strings.Builder`).

## ❓ Review questions

1. What is the zero value of an `int`, a `string`, a `bool`, and a slice?
2. Why does `7 / 2` evaluate to `3` instead of `3.5` in Go, and how would you
   get `3.5`?
3. What does `&&` do differently from `&` in an `if` condition, and why does
   that matter for safety when checking a slice index before using it?
4. Why can `len("héllo")` return a number larger than the number of visible
   characters in the string?
5. What happens if you convert the `int` value `300` to a `byte`, and why?
6. Why can `0.1 + 0.2` differ from the nearest `float64` representation of
   `0.3`, and why is one fixed epsilon not correct for every calculation?

## 📚 References

- [The Go Programming Language Specification](https://go.dev/ref/spec)
- [A Tour of Go: Basics](https://go.dev/tour/basics/1)
- [Strings, bytes, runes and characters in Go](https://go.dev/blog/strings)
- [`unicode/utf8` package docs](https://pkg.go.dev/unicode/utf8)

Next: [Module 2 — Control Flow](../02_control_flow/README.md).
