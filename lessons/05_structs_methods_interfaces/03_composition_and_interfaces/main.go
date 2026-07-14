// Package main covers composition through embedding, Go's implicit interface
// satisfaction, small consumer-owned interfaces, and fmt.Stringer.
package main

import "fmt"

// Person is embedded (composed) into Employee below. Go has no inheritance;
// embedding instead promotes a field's fields and methods to the outer type.
type Person struct {
	Name string
	Age  int
}

// Describe is a method on Person. Because Employee embeds Person, this
// method is promoted and callable directly on an Employee value.
func (p Person) Describe() string {
	return fmt.Sprintf("%s (%d)", p.Name, p.Age)
}

// Employee composes Person "by embedding" (no field name, just the type).
// This models "an Employee has the fields and behavior of a Person" without
// claiming an Employee "is a" Person in a type-hierarchy sense.
type Employee struct {
	Person
	Title string
}

// String implements fmt.Stringer. Any type with a String() string method
// satisfies fmt.Stringer, and fmt automatically calls it for %v, %s, and
// Println - no explicit "implements" declaration is needed or possible.
func (e Employee) String() string {
	return fmt.Sprintf("%s - %s", e.Person.Describe(), e.Title)
}

// notifier is a small interface owned by the code that consumes it (main, in
// this lesson), not by the types that implement it. This is idiomatic Go:
// define interfaces where a value is used, sized to exactly what the caller
// needs, rather than exporting one large interface from the producing
// package. Any type with a matching Notify method satisfies this interface
// automatically - that is implicit satisfaction.
type notifier interface {
	Notify(message string) string
}

// EmailNotifier and SMSNotifier never mention the notifier interface. Each
// simply happens to have a Notify method, which is enough to satisfy it.
type EmailNotifier struct{ Address string }

func (e EmailNotifier) Notify(message string) string {
	return fmt.Sprintf("email to %s: %s", e.Address, message)
}

type SMSNotifier struct{ Number string }

func (s SMSNotifier) Notify(message string) string {
	return fmt.Sprintf("sms to %s: %s", s.Number, message)
}

// send depends only on the tiny notifier interface, so it works with any
// current or future type that has a Notify method - including types defined
// in packages that have never heard of this one.
func send(n notifier, message string) string {
	return n.Notify(message)
}

func main() {
	employee := Employee{
		Person: Person{Name: "Ada", Age: 30},
		Title:  "Engineer",
	}

	// The promoted Describe method is reachable directly on Employee.
	fmt.Println("promoted method:", employee.Describe())

	// fmt uses the Stringer implementation automatically for %v and Println.
	fmt.Println("stringer via Println:", employee)
	fmt.Printf("stringer via %%v:      %v\n", employee)

	// Explicit field access to the embedded value is still available when a
	// promoted name would be ambiguous or when being explicit reads better.
	fmt.Println("explicit embedded field:", employee.Person.Name)

	fmt.Println(send(EmailNotifier{Address: "ada@example.com"}, "build passed"))
	fmt.Println(send(SMSNotifier{Number: "+1-555-0100"}, "build passed"))

	// Because notifier is implicit, main can even satisfy it with a type
	// defined right here, without editing EmailNotifier or SMSNotifier.
	fmt.Println(send(consoleNotifier{}, "build passed"))
}

type consoleNotifier struct{}

func (consoleNotifier) Notify(message string) string {
	return "console: " + message
}
