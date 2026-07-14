package basictests

import "testing"

// TestReverse shows the smallest possible test: a Test-prefixed function
// taking *testing.T. Go's test runner discovers it automatically because it
// lives in a file ending in _test.go.
func TestReverse(t *testing.T) {
	got := Reverse("Go!")
	want := "!oG"
	if got != want {
		// t.Errorf reports a failure but lets the rest of the test function
		// keep running, so later checks in the same test still execute.
		t.Errorf("Reverse(%q) = %q, want %q", "Go!", got, want)
	}
}

func TestReverseEmpty(t *testing.T) {
	if got := Reverse(""); got != "" {
		// t.Fatalf reports a failure and stops the current test function
		// immediately. Use it when later checks would be meaningless or
		// would panic without the missing precondition.
		t.Fatalf("Reverse(\"\") = %q, want an empty string", got)
	}
}

func TestIsPalindrome(t *testing.T) {
	// t.Log and t.Logf print diagnostic output. It is hidden unless the
	// test fails or you pass -v.
	t.Log("checking a known palindrome")

	if !IsPalindrome("Level") {
		t.Error("expected \"Level\" to be a palindrome")
	}
}

func TestIsPalindromeFalse(t *testing.T) {
	if IsPalindrome("Golang") {
		t.Error("expected \"Golang\" not to be a palindrome")
	}
}

// TestIsPalindromeIgnoresSpaces documents a known limitation: this simple
// implementation compares runes directly, so phrases containing spaces are
// never palindromes even when a human would consider them one. t.Skip marks
// the test as intentionally not run, with a reason visible in -v output.
func TestIsPalindromeIgnoresSpaces(t *testing.T) {
	t.Skip("known limitation: spaces are not ignored; normalize input before calling IsPalindrome")
}
