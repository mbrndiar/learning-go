# 🧱 Module 5 — Structs, Methods and Interfaces

This module moves from plain values to custom types. You will group related
data in structs, attach behavior with methods, model relationships through
composition instead of inheritance, and describe capabilities with Go's
implicitly satisfied interfaces.

**Prerequisites:** Modules 1–4 (Basics, Control Flow, Functions and Pointers,
Collections). This module assumes comfort with named types, functions,
pointers, and slices/maps, since methods and receivers build directly on how
Go passes values and pointers around.

## 🎯 Learning goals

By the end of this module you will be able to:

- define struct types and create values with zero values and literals;
- choose value receivers versus pointer receivers deliberately;
- compose types through embedding instead of a class hierarchy;
- design small, consumer-owned interfaces that types satisfy implicitly;
- implement `fmt.Stringer` for readable output;
- build iota-based enum-like types; and
- recognize and avoid the nil-interface pitfall.

## 📖 Structs, methods and interfaces, explained

A **struct** is a named collection of fields: a way to give a group of
related values one type of its own instead of passing separate variables
around. Every struct has a **zero value** — `var p Point` and `Point{}` both
produce a struct whose fields are each field type's own zero value. Whether
that state is meaningful for the domain depends on the type's design; aim for
a useful zero value when practical, and validate invariants when zero is not
valid. A struct literal can be **positional** (`Point{3, 4}`) or **keyed**
(`Point{X: 3, Y: 4}`). Positional literals depend on field declaration order,
which is easy to change by accident; once a struct is part of an API other
code depends on, reordering, inserting, or removing a field can silently
change what a positional literal means, or break compilation outright. Keyed
literals are stable against that kind of change, so prefer them for any
struct used outside the file that defines it.

A **method** is an ordinary function with one extra receiver parameter
written before its name (`func (c Counter) Snapshot() string`). A **value
receiver** gets a copy of the struct: reads work, but any mutation inside the
method to the receiver's own fields changes only that copy. Reference-like
fields inside the copy can still point at shared data, just as Module 3
explained for slices and maps. A **pointer receiver**
(`func (c *Counter) Add(amount int)`) gets the address of the caller's value,
so mutations are visible to the caller and the struct is not copied on every
call. A type's **method set** is which methods can be called on a given
value: a plain value only has its value-receiver methods, while a pointer has
both the value- and pointer-receiver methods. Go automatically takes the
address of an addressable variable, so `counter.Add(1)` works without writing
`(&counter).Add(1)` — but that automatic address-of requires the value to be
addressable in the first place. A struct value stored directly inside a map
is **not addressable** (`counters["clicks"].Add(1)` does not compile),
because a map does not expose a stable address for its elements; read the
element out, mutate the copy, and store it back instead.

**Embedding** places one type inside another with no field name
(`type Employee struct { Person; Title string }`). It **promotes** the
embedded type's fields and methods so they are reachable directly on the
outer type (`employee.Describe()`). That is composition, not inheritance:
Go has no class hierarchy and no `extends`, and embedding never establishes
that an `Employee` "is a" `Person` that could be substituted anywhere a
`Person` is expected. Embedding only saves hand-written forwarding methods;
it creates no subtype relationship the type system understands.

An **interface** is a set of method signatures describing a capability, and
Go interfaces are satisfied **implicitly**: any type with the required
methods satisfies the interface automatically, with no `implements` keyword
and no declaration linking the concrete type to it. A type's author does not
need to know an interface exists for the type to satisfy it, which is why
`EmailNotifier` and `SMSNotifier` in lesson 3 never mention the `notifier`
interface they satisfy. Idiomatic Go interfaces are usually small (often one
or two methods) and are typically declared by the package that *consumes* a
value through the interface, not by the package that produces the concrete
type — define an interface where it is used, not next to what implements it.

An **interface value** is really a pair: a dynamic type and a dynamic value.
An interface equals `nil` only when both halves are unset. That explains the
classic pitfall in lesson 4: `return problem` for a nil `*NotFoundError`
produces a non-nil `error`, because the interface's dynamic type is
`*NotFoundError` even though its dynamic value is nil — the pair as a whole
is not the zero interface value. Avoid it by returning the bare `nil`
literal whenever there is truly no error, rather than a variable whose
static type is a concrete pointer.

`iota` is not a distinct enum feature; it is a counter that starts at `0` in
each `const` block and increases by one per constant line, so a single
repeated expression (`Status = iota`) numbers every constant without writing
each value by hand. It adds no extra type safety beyond whatever named type
it is attached to. Once those numbers are stored in a file, a database
column, or sent over a wire format, the specific value tied to each name
becomes part of the type's external contract: reordering or inserting
constants in the middle changes what already-persisted values mean, so treat
the sequence as append-only from that point on.

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
  depends on declaration order; reordering same-typed fields can silently
  change meaning, while adding or removing a field breaks compilation. Prefer
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

## 🔬 Experiment

Starting from lesson 3's `notifier` example, add a `Priority` type built with
`iota` (`Low`, `Normal`, `High`), give `Employee` a pointer-receiver method
`Promote(newTitle string)` that mutates `Title`, and write
`notifyAll(notifiers []notifier, message string) []string` that depends only
on the `notifier` interface — call it with a mix of `EmailNotifier`,
`SMSNotifier`, and a brand-new type you define inline in `main`. Then store an
`Employee` value (not a pointer) in a `map[string]Employee`, try calling
`Promote` directly on the map element, confirm the compiler rejects it, and
fix it with the read-mutate-write pattern from lesson 2. Finally, write a
function that returns a nil `*NotFoundError` through a plain `error` return
type and prove with `== nil` that the result is non-nil.

Practice the same ideas with tests behind them in the matching exercise:
[`exercises/05_structs_methods_interfaces/`](../../exercises/05_structs_methods_interfaces/).

## 🔗 Related reading

- <https://go.dev/ref/spec#Struct_types>
- <https://go.dev/ref/spec#Method_sets>
- <https://go.dev/ref/spec#Interface_types>
- <https://go.dev/doc/effective_go#embedding>
- <https://go.dev/ref/spec#Iota>
