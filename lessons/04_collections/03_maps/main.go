// This lesson covers maps: Go's built-in hash table type. It shows
// creation, membership checks ("comma ok"), deletion, and why map
// iteration order is intentionally randomized.
package main

import (
	"fmt"
	"slices"
)

func main() {
	fmt.Println("--- creating and populating a map ---")
	// A map literal creates keys and values together. Map keys must be a
	// "comparable" type (numbers, strings, bools, structs of comparable
	// fields, and so on); slices and other maps cannot be keys.
	scores := map[string]int{
		"Ada":   95,
		"Grace": 88,
	}
	scores["Alan"] = 91 // adding a new key
	fmt.Printf("scores=%v\n", scores)

	fmt.Println("--- the zero value of a map is nil, and reading is safe ---")
	var empty map[string]int
	fmt.Printf("empty map: %v, lookup on nil map: %d\n", empty, empty["anything"])
	// Reading a nil map returns the zero value for every key, but WRITING
	// to a nil map panics. Always initialize a map with make() or a
	// literal before writing to it.
	// empty["x"] = 1 // would panic: assignment to entry in nil map

	fmt.Println("--- membership: the \"comma ok\" idiom ---")
	// A map access can return a second boolean telling you whether the
	// key was actually present. This distinguishes "the key is missing"
	// from "the key exists and its value happens to be the zero value".
	if value, ok := scores["Ada"]; ok {
		fmt.Printf("Ada's score is %d\n", value)
	}
	if _, ok := scores["Ismael"]; !ok {
		fmt.Println("Ismael has no score recorded")
	}
	// Without the second return value, a missing key silently returns 0,
	// which looks identical to a real score of 0.
	missingScore := scores["Ismael"]
	fmt.Printf("scores[\"Ismael\"] without comma-ok = %d (looks like a real zero!)\n", missingScore)

	fmt.Println("--- delete removes a key; deleting a missing key is a no-op ---")
	delete(scores, "Grace")
	delete(scores, "NotPresent") // safe; does nothing
	fmt.Printf("scores after delete: %v\n", scores)

	fmt.Println("--- len works on maps too ---")
	fmt.Printf("number of entries: %d\n", len(scores))

	fmt.Println("--- map iteration order is randomized on purpose ---")
	// Go deliberately randomizes the order of "for range" over a map so
	// that no code accidentally depends on an order that was never
	// guaranteed. Running this program multiple times can print the pairs
	// in a different order each time.
	for name, score := range scores {
		_ = name
		_ = score
	}
	fmt.Println("(iterated once above; order is not guaranteed or printed)")

	fmt.Println("--- for deterministic output, sort the keys first ---")
	names := make([]string, 0, len(scores))
	for name := range scores {
		names = append(names, name)
	}
	slices.Sort(names) // now the order is alphabetical and repeatable
	for _, name := range names {
		fmt.Printf("%s: %d\n", name, scores[name])
	}

	fmt.Println("--- maps as sets: using an empty struct value ---")
	// A common Go idiom for a "set" is map[T]struct{}: struct{} takes
	// zero bytes, so it signals "this key is present" as cheaply as
	// possible, without a meaningless bool or int value.
	visited := map[string]struct{}{}
	visited["home"] = struct{}{}
	visited["about"] = struct{}{}
	_, wasVisited := visited["about"]
	fmt.Printf("was \"about\" visited? %t\n", wasVisited)
}
