# 🧱 05 — Structs, Methods, Interfaces

Apply [Module 5](../../lessons/05_structs_methods_interfaces/README.md) by
modeling a small catalog of 2D shapes with struct fields, value and pointer
receivers, interfaces, `fmt.Stringer`, and `iota`-based enumerations.

## ▶️ Workflow

```bash
go test ./exercises/05_structs_methods_interfaces
go test ./exercises/05_structs_methods_interfaces/solution
```

The starter package in this folder (`shapes.go`) fails until you implement
every method. Compare with `solution/shapes.go` only after a genuine attempt.

## 🧩 Tasks

1. Implement `Rectangle.Area()` and `Rectangle.Perimeter()` using value
   receivers.
2. Implement `Rectangle.String()` so `Rectangle` satisfies `fmt.Stringer`,
   rendering `"Rectangle(width=%.2f, height=%.2f)"`.
3. Implement `Rectangle.Scale(factor float64)` using a pointer receiver that
   mutates `Width` and `Height` in place. A factor `<= 0` must be a no-op.
4. Implement `Circle.Area()` and `Circle.Perimeter()` using `math.Pi`.
5. Implement `Circle.String()` (`"Circle(radius=%.2f)"`) and
   `Circle.Scale(factor float64)` with the same pointer-receiver and no-op
   rule as task 3.
6. Declare the `Size` enumeration (`Small`, `Medium`, `Large`) with `iota` and
   implement `Size.String()`, returning `"Unknown"` for any other value.
7. Implement the free functions `Classify(s Shape) Size` (thresholds: area
   `< 10` is `Small`, `< 100` is `Medium`, otherwise `Large`) and
   `TotalArea(shapes []Shape) float64`, which sums `Area()` across a slice of
   the `Shape` interface.

## 🔍 What this covers

- Struct field definitions and zero values.
- Value receivers for read-only behavior versus pointer receivers for
  mutation.
- Defining and satisfying an interface (`Shape`) implicitly.
- Implementing `fmt.Stringer` for readable `%v`/`%s` formatting.
- `iota` for compact, ordered enumerations and giving them a `String()`
  method.
- Writing polymorphic functions that accept an interface.

## ⚠️ Common mistakes

- Forgetting that a pointer receiver is required to mutate a struct's fields
  from inside a method — a value receiver only mutates a copy.
- Defining `String()` with a pointer receiver when call sites use a value,
  which silently excludes the value type from satisfying `fmt.Stringer`.
- Comparing floating-point areas with `==` instead of a small epsilon
  tolerance.
