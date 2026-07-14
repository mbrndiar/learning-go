// This lesson covers pointers: what they are, why Go's default "pass by
// value" semantics matter, and how a pointer lets a function modify a
// caller's variable. It deliberately avoids structs (module 5 covers
// those); the same motivation applies to any value type, including plain
// numbers and arrays.
package main

import "fmt"

// incrementByValue receives a COPY of the caller's int. Changing "n"
// here has no effect outside this function, because "n" is an entirely
// separate variable that just happens to start with the same value.
func incrementByValue(n int) {
	n++
}

// incrementByPointer receives the memory address of the caller's int (a
// *int, "pointer to int"). *n dereferences the pointer, meaning "the
// value stored at this address". Assigning through *n changes the
// original variable, because there is only one int in memory, and both
// the caller and this function refer to it via the pointer.
func incrementByPointer(n *int) {
	*n++
}

// doubleAll shows the same idea applied to an array. Go arrays are value
// types: passing one to a function copies every element. A pointer to
// the array lets the function mutate the caller's original array
// in place instead of a throwaway copy.
func doubleAll(numbers *[3]int) {
	for i := range numbers {
		numbers[i] *= 2 // Go automatically dereferences numbers[i] for you
	}
}

func main() {
	fmt.Println("--- value semantics: a copy cannot change the original ---")
	value := 10
	incrementByValue(value)
	fmt.Printf("after incrementByValue: value=%d (unchanged)\n", value)

	fmt.Println("--- pointers let a function modify the caller's variable ---")
	incrementByPointer(&value) // &value takes the address of "value"
	fmt.Printf("after incrementByPointer: value=%d (changed)\n", value)

	fmt.Println("--- the zero value of a pointer is nil ---")
	var pointer *int
	fmt.Printf("pointer=%v (nil means it points at nothing yet)\n", pointer)
	// Dereferencing a nil pointer panics at runtime: *pointer = 5 would
	// crash the program. Always ensure a pointer is non-nil (or check
	// explicitly) before dereferencing it.
	pointer = &value
	fmt.Printf("after assigning &value, *pointer=%d\n", *pointer)

	fmt.Println("--- pointers to arrays avoid copying and enable mutation ---")
	numbers := [3]int{1, 2, 3}
	doubleAll(&numbers)
	fmt.Printf("numbers after doubleAll: %v\n", numbers)

	fmt.Println("--- why this matters for methods later ---")
	// Module 5 introduces methods with pointer receivers, such as
	// func (t *Task) Complete(). The motivation is exactly what you just
	// saw: a method with a value receiver only ever sees a copy of the
	// data, so it cannot permanently change the caller's original value.
	// A pointer receiver lets the method mutate the real data in place,
	// which is why types that represent mutable state usually define
	// their methods on a pointer receiver.
	fmt.Println("(see module 5 for methods with pointer receivers)")

	fmt.Println("--- swapping two values needs pointers ---")
	a, b := 1, 2
	swap(&a, &b)
	fmt.Printf("after swap: a=%d b=%d\n", a, b)
}

// swap exchanges the values stored at two addresses. Without pointers,
// a swap function could only exchange its own local copies, leaving the
// caller's variables untouched.
func swap(x, y *int) {
	*x, *y = *y, *x
}
