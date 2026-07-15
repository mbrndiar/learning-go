// This lesson covers Go's switch statement: expression switches,
// switches with no condition (a clean alternative to long if/else
// chains), fallthrough, and type switches.
package main

import "fmt"

func main() {
	fmt.Println("--- basic expression switch ---")
	day := "Wednesday"
	switch day {
	case "Saturday", "Sunday": // a case can list multiple values
		fmt.Println("weekend")
	case "Wednesday":
		fmt.Println("midweek")
	default:
		fmt.Println("a regular weekday")
	}
	// Unlike C or Java, Go's switch does NOT fall through to the next
	// case by default: each case breaks automatically once its body
	// finishes. This removes an entire class of bugs caused by a
	// forgotten "break".

	fmt.Println("--- switch with an initialization statement ---")
	switch score := 82; {
	case score >= 90:
		fmt.Println("grade: A")
	case score >= 80:
		fmt.Println("grade: B")
	default:
		fmt.Println("grade: C or below")
	}

	fmt.Println("--- switch with no expression (if/else replacement) ---")
	// A switch with no condition after the keyword is equivalent to
	// switch true, and each case is a boolean expression evaluated in
	// order. This reads more clearly than a long if/else if chain when
	// there are several independent conditions.
	temperature := 21
	switch {
	case temperature < 0:
		fmt.Println("freezing")
	case temperature < 15:
		fmt.Println("cold")
	case temperature < 25:
		fmt.Println("comfortable")
	default:
		fmt.Println("hot")
	}

	fmt.Println("--- explicit fallthrough ---")
	// Because cases do not fall through automatically, Go provides the
	// explicit "fallthrough" keyword for the rare cases where you want
	// the next case's body to run regardless of its own condition.
	switch grade := 'B'; grade {
	case 'A':
		fmt.Println("excellent")
		fallthrough
	case 'B':
		fmt.Println("good or better")
		fallthrough
	case 'C':
		fmt.Println("passing or better")
	default:
		fmt.Println("needs improvement")
	}

	fmt.Println("--- type switch ---")
	// A type switch inspects the dynamic type stored in an interface
	// value. "value := x.(type)" is special syntax only valid inside a
	// switch; each case gives "value" the corresponding concrete type.
	// This example previews `any` (interfaces, module 5) and an anonymous
	// function (module 3). For now, focus on how the switch chooses a case
	// from the value's runtime type.
	describe := func(value any) string {
		switch v := value.(type) {
		case int:
			return fmt.Sprintf("int with double %d", v*2)
		case string:
			return fmt.Sprintf("string of length %d", len(v))
		case bool:
			return fmt.Sprintf("bool negated to %t", !v)
		case nil:
			return "nil value"
		default:
			return fmt.Sprintf("unhandled type %T", v)
		}
	}
	for _, sample := range []any{7, "hello", true, nil, 3.14} {
		fmt.Println(describe(sample))
	}
}
