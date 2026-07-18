# 🌱 Exercises: Basics

Practice the concepts from
[Module 1](../../lessons/01_basics/README.md): conversions and string/rune
helpers using only values, basic types, operators, and `for`/`if` control flow.

## 🧩 Tasks

1. `CelsiusToFahrenheit` — convert a temperature, exact formula.
2. `FahrenheitToCelsius` — convert the other direction.
3. `ParseIntOrDefault` — convert a string to an `int`, falling back to a
   default value when the string is not a valid number.
4. `ReverseString` — reverse a string by rune, not by byte, so multi-byte
   characters stay intact.
5. `CountVowels` — count vowels (`a e i o u`, either case) by iterating runes.
6. `IsPalindrome` — check whether a string reads the same forwards and
   backwards, ignoring case, by comparing runes.
7. `ByteAndRuneLen` — return both the byte length and the rune length of a
   string so the difference is visible for multi-byte input.

## ▶️ Commands

```bash
go test ./exercises/01_basics/...
go test -run '^$' ./exercises/01_basics
go test ./exercises/01_basics/solution
gofmt -l exercises/01_basics
```

## 📝 Notes

- `len(s)` on a string counts bytes, not characters. Convert to `[]rune(s)`
  to work per character.
- `strconv.Atoi` returns `(int, error)`; use the error to decide when to fall
  back to a default.
- Compare with `solution/` only after a genuine attempt.
