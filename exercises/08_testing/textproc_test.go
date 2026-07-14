package textproc

import "testing"

// TODO(task 1): Replace this body with a table-driven test for Normalize.
// Cover: leading/trailing whitespace, multiple interior spaces, mixed case,
// and an empty string. Use a []struct{name, in, want string} table and
// t.Run(tt.name, ...) for each case.
func TestNormalize(t *testing.T) {
	t.Skip("TODO: implement TestNormalize as a table-driven test (see README task 1)")
}

// TODO(task 2): Replace this body with a table-driven test for WordFrequency.
// Cover: a normal sentence, repeated words, an empty string (expect an empty
// map), and mixed-case words that should be counted together.
func TestWordFrequency(t *testing.T) {
	t.Skip("TODO: implement TestWordFrequency as a table-driven test (see README task 2)")
}

// TODO(task 3): Replace this body with a table-driven test for Reverse.
// Cover: empty string, single rune, ASCII word, and a multi-byte string such
// as "héllo" to prove runes (not bytes) are reversed.
func TestReverse(t *testing.T) {
	t.Skip("TODO: implement TestReverse as a table-driven test (see README task 3)")
}

// TODO(task 4): Replace this body with a table-driven test for SafeSlice.
// Cover at least one success case and at least two failure cases (negative
// start, negative length, start beyond the rune length, and end beyond the
// rune length all count as distinct failure cases -- pick at least two).
// Assert an error is returned; SafeSlice must never panic.
func TestSafeSlice(t *testing.T) {
	t.Skip("TODO: implement TestSafeSlice as a table-driven test (see README task 4)")
}

// TODO(task 5): Replace this body with a real benchmark. Build a
// representative input string once before the loop, call b.ResetTimer, then
// call WordFrequency(input) b.N times inside a "for i := 0; i < b.N; i++"
// loop.
func BenchmarkWordFrequency(b *testing.B) {
	b.Skip("TODO: implement BenchmarkWordFrequency (see README task 5)")
}

// TODO(task 6): Replace this body with a real fuzz target. Seed the corpus
// with f.Add calls for a handful of (s, start, length) combinations
// (including ones you expect to fail), then use f.Fuzz to assert that
// SafeSlice never panics for any input and that whenever it returns a nil
// error, the returned string's rune count equals length.
func FuzzSafeSlice(f *testing.F) {
	f.Skip("TODO: implement FuzzSafeSlice (see README task 6)")
}
