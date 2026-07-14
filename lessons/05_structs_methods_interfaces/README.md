# 🧱 Module 5 — Structs, Methods and Interfaces

This module moves from plain values to custom types. You will group related
data in structs, attach behavior with methods, model relationships through
composition instead of inheritance, and describe capabilities with Go's
implicitly satisfied interfaces.

## 🎯 Learning goals

By the end of this module you will be able to:

- define struct types and create values with zero values and literals;
- choose value receivers versus pointer receivers deliberately;
- compose types through embedding instead of a class hierarchy;
- design small, consumer-owned interfaces that types satisfy implicitly;
- implement `fmt.Stringer` for readable output;
- build iota-based enum-like types; and
- recognize and avoid the nil-interface pitfall.

## 📦 Lessons

1. [`01_structs_and_literals/`](01_structs_and_literals/) — struct
   definitions, zero values, keyed and positional literals, nested and
   anonymous structs, struct comparison.
2. [`02_methods_and_receivers/`](02_methods_and_receivers/) — methods, value
   versus pointer receivers, mutation, addressability with maps.
3. [`03_composition_and_interfaces/`](03_composition_and_interfaces/) —
   embedding and method promotion, implicit interface satisfaction, small
   consumer-owned interfaces, `fmt.Stringer`.
4. [`04_iota_and_nil_interfaces/`](04_iota_and_nil_interfaces/) — `iota`
   enum-like types, and the nil-interface pitfall with typed nil pointers.

## ▶️ How to run a lesson

From the repository root:

```bash
go run ./lessons/05_structs_methods_interfaces/01_structs_and_literals
go run ./lessons/05_structs_methods_interfaces/02_methods_and_receivers
go run ./lessons/05_structs_methods_interfaces/03_composition_and_interfaces
go run ./lessons/05_structs_methods_interfaces/04_iota_and_nil_interfaces
```

Predict each program's output before running it, then change one field,
receiver, or case at a time and re-run.

## 🚧 Common mistakes

- **Positional literals that outlive the struct's shape.** `Point{3, 4}`
  breaks silently if a field is inserted between `X` and `Y` later. Prefer
  keyed literals (`Point{X: 3, Y: 4}`) once a struct is used outside the file
  that defines it.
- **Mixing receiver kinds on one type.** Calling one method with a value
  receiver and another with a pointer receiver on the same type produces
  a method set that is hard to reason about. Pick pointer receivers for the
  whole type as soon as any one method needs to mutate.
- **Calling a pointer method through a non-addressable value.** A struct
  value stored directly in a map is not addressable
  (`counters["clicks"].Add(1)` does not compile). Read the value out, mutate
  the copy, and store it back.
- **Exporting a big interface from the producing package.** Idiomatic Go
  defines small interfaces close to the code that consumes them (see
  `notifier` in lesson 3), not one large interface exported alongside the
  concrete types.
- **Returning a typed nil pointer as `error`.** `return problem` where
  `problem` is a nil `*NotFoundError` produces a non-nil `error` interface
  value, because an interface is nil only when both its type and its value
  are unset. Return the bare `nil` literal when there is no error.
- **Reordering `iota` constants after they are persisted somewhere.** The
  numeric value is part of the type's contract once it is stored in a file,
  database, or wire format; treat the constant order as append-only.

## ❓ Review questions

1. When does Go copy a struct, and how does that interact with pointer
   receivers?
2. Why does embedding model composition rather than inheritance?
3. What makes an interface satisfied "implicitly," and why does that favor
   small, consumer-owned interfaces?
4. How does `fmt` decide to call a type's `String() string` method?
5. Why is `iota` numbering considered part of a type's external contract once
   the values are stored or transmitted?
6. Concretely, why is an interface holding a typed nil pointer not equal to
   `nil`, and how do you avoid producing one?
