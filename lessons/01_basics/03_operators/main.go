// This lesson covers Go's arithmetic, comparison, logical, and bitwise
// operators, along with the surprising edge cases that trip up beginners:
// integer division, operator precedence, and short-circuit evaluation.
package main

import "fmt"

func main() {
	// Arithmetic operators: + - * / %. Division between two integers is
	// integer division: it truncates the fractional part instead of
	// rounding. This is a very common beginner surprise.
	fmt.Println("--- arithmetic ---")
	fmt.Printf("7 / 2 = %d (integer division truncates)\n", 7/2)
	fmt.Printf("7.0 / 2.0 = %v (float division keeps the fraction)\n", 7.0/2.0)
	fmt.Printf("7 %% 2 = %d (remainder, same sign as the dividend)\n", 7%2)
	fmt.Printf("-7 %% 2 = %d (Go's %% keeps the sign of -7, not 2)\n", -7%2)

	// Comparison operators return a bool: == != < <= > >=.
	fmt.Println("--- comparisons ---")
	left, right := 3, 3
	fmt.Printf("3 < 5 = %t, 3 == 3 = %t\n", left < 5, left == right)

	// Logical operators && and || short-circuit: the right-hand side is
	// only evaluated if it can change the result. This means it is safe
	// to guard a risky operation with a check on the left side.
	fmt.Println("--- short-circuit evaluation ---")
	numbers := []int{1, 2, 3}
	index := 5
	// If index is out of range, the left side (index < len(numbers)) is
	// false, so Go never evaluates numbers[index] and never panics.
	if index < len(numbers) && numbers[index] > 0 {
		fmt.Println("found a positive number")
	} else {
		fmt.Println("index guarded safely; no out-of-range access happened")
	}

	// Operator precedence follows familiar math rules: * / %% bind
	// tighter than + -, and comparisons bind tighter than && and ||.
	// Parentheses make intent explicit and are cheap, so prefer them
	// whenever the order is not obvious at a glance.
	fmt.Println("--- precedence ---")
	withoutParens := 2 + 3*4
	withParens := (2 + 3) * 4
	fmt.Printf("2 + 3 * 4 = %d, (2 + 3) * 4 = %d\n", withoutParens, withParens)

	// Bitwise operators work on the binary representation of integers:
	// & (AND), | (OR), ^ (XOR as a binary operator, NOT as unary),
	// &^ (AND NOT, i.e. "clear bits"), << and >> (shifts).
	fmt.Println("--- bitwise ---")
	const (
		readPermission  = 1 << 0 // 0b001
		writePermission = 1 << 1 // 0b010
		execPermission  = 1 << 2 // 0b100
	)
	permissions := readPermission | writePermission // grant read and write
	fmt.Printf("permissions = %03b\n", permissions)
	fmt.Printf("has write? %t\n", permissions&writePermission != 0)
	permissions &^= writePermission // revoke write ("AND NOT")
	fmt.Printf("after revoking write: %03b\n", permissions)
	fmt.Printf("has exec? %t\n", permissions&execPermission != 0)

	// Compound assignment operators combine an operator with =.
	total := 10
	total += 5
	total *= 2
	fmt.Printf("compound assignment: %d\n", total)

	// ++ and -- are statements, not expressions, in Go: they cannot be
	// used inside a larger expression like x := y++. This avoids the
	// confusing pre/post-increment puzzles found in C-like languages.
	counter := 0
	counter++
	fmt.Printf("counter after counter++ = %d\n", counter)
}
