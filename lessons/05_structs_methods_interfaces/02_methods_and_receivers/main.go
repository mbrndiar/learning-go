// Package main compares value receivers and pointer receivers on methods.
package main

import "fmt"

// Counter is deliberately tiny so the copy-versus-share behavior is easy to
// see in the output below.
type Counter struct {
	Name  string
	Total int
}

// Snapshot has a value receiver. It receives a *copy* of the Counter, so it
// can read fields but any change it made would be discarded when it returns.
// Value receivers are appropriate for methods that only read data, and for
// small, immutable-style types.
func (c Counter) Snapshot() string {
	return fmt.Sprintf("%s=%d", c.Name, c.Total)
}

// Add has a pointer receiver. It receives the address of the caller's
// Counter, so mutations are visible to the caller. Pointer receivers are
// required whenever a method must mutate the receiver, and are also the
// conventional choice for large structs to avoid copying them on every call.
func (c *Counter) Add(amount int) {
	c.Total += amount
}

// Reset also needs a pointer receiver because it mutates the receiver.
func (c *Counter) Reset() {
	c.Total = 0
}

func main() {
	counter := Counter{Name: "clicks"}

	// Go automatically takes the address of an addressable value, so calling
	// a pointer-receiver method on a plain variable works without writing
	// (&counter).Add(1) yourself.
	counter.Add(1)
	counter.Add(4)
	fmt.Println("after two adds:", counter.Snapshot())

	counter.Reset()
	fmt.Println("after reset:", counter.Snapshot())

	// Mixing receiver kinds on the same type is legal but discouraged: pick
	// one receiver kind per type and use it consistently, because a mixed
	// method set is confusing to callers who cannot tell which methods
	// mutate without reading each signature.
	//
	// Rule of thumb used in this course: if any method needs a pointer
	// receiver, give every method on that type a pointer receiver.

	// Addressability matters for interfaces and map values. A value stored
	// directly in a map is not addressable, so a pointer-receiver method
	// cannot be called on it in place.
	counters := map[string]Counter{"clicks": counter}
	// counters["clicks"].Add(1) // compile error: cannot call pointer method

	// Instead, read the value out, mutate the copy, and store it back.
	entry := counters["clicks"]
	entry.Add(1)
	counters["clicks"] = entry
	fmt.Println("map entry after add:", counters["clicks"].Snapshot())

	// A pointer variable is always addressable, so pointer-receiver methods
	// work directly through it.
	ptr := &Counter{Name: "visits"}
	ptr.Add(10)
	fmt.Println("through pointer:", ptr.Snapshot())
}
