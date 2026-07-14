// This lesson focuses on the sharp edges of slices: two slices can share
// the same backing array, so mutating one can silently affect the other.
// It also covers copy(), which is the standard way to make an
// independent duplicate.
package main

import "fmt"

func main() {
	fmt.Println("--- slicing shares the backing array ---")
	original := []int{1, 2, 3, 4, 5}
	view := original[1:4] // shares memory with "original"
	fmt.Printf("original=%v view=%v\n", original, view)

	view[0] = 999 // mutates original[1] too, because they share storage
	fmt.Printf("after view[0]=999 -> original=%v view=%v\n", original, view)

	fmt.Println("--- append can silently overwrite a sibling slice ---")
	// Both "front" and "back" start out sharing the same backing array as
	// "letters". Appending to "front" writes into the shared array
	// because there is spare capacity, which unexpectedly changes what
	// "letters" and "back" see too.
	letters := make([]byte, 5, 8) // len=5, cap=8: 3 spare slots
	copy(letters, []byte("abcde"))
	front := letters[:2]
	back := letters[2:]
	fmt.Printf("before append: letters=%s front=%s back=%s\n", letters, front, back)
	front = append(front, 'X', 'Y') // writes into letters[2] and letters[3]
	fmt.Printf("after append to front: letters=%s front=%s back=%s\n", letters, front, back)
	fmt.Println("(back changed too, because front's append reused shared capacity)")

	fmt.Println("--- copy() makes an independent duplicate ---")
	// copy(dst, src) copies min(len(dst), len(src)) elements and returns
	// that count. It never grows dst, so dst must already have enough
	// length to hold what you want copied.
	source := []int{1, 2, 3}
	duplicate := make([]int, len(source))
	copiedCount := copy(duplicate, source)
	duplicate[0] = 42
	fmt.Printf("source=%v duplicate=%v copiedCount=%d (independent now)\n",
		source, duplicate, copiedCount)

	fmt.Println("--- full slice expressions limit shared capacity ---")
	// letters[low:high:max] sets the resulting slice's capacity to
	// max-low, not the full remaining capacity of the backing array. This
	// forces the next append to allocate a fresh array instead of
	// silently overwriting a sibling slice's data.
	safeFront := letters[:2:2] // len=2, cap=2: no spare capacity to borrow
	safeFront = append(safeFront, 'Z')
	fmt.Printf("letters=%s safeFront=%s (letters untouched by that append)\n", letters, safeFront)

	fmt.Println("--- appending two independent slices together ---")
	a := []int{1, 2}
	b := []int{3, 4}
	// append(a, b...) is the idiomatic way to concatenate slices. Because
	// "a" may or may not have spare capacity, always reassign the result;
	// never assume append mutates in place.
	combined := append(a, b...)
	fmt.Printf("combined=%v\n", combined)
}
