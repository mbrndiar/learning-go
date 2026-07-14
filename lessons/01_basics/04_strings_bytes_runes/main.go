// This lesson explains how Go represents text: strings are immutable
// sequences of bytes that are conventionally (but not enforced to be)
// UTF-8 encoded text. Understanding the difference between a byte, a
// rune, and a "character" is essential before working with any
// user-facing text that might contain non-ASCII characters.
package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

func main() {
	// "héllo" mixes an ASCII-only word with one accented letter. In
	// UTF-8, ASCII characters take 1 byte each, but "é" takes 2 bytes.
	word := "héllo"

	fmt.Println("--- length vs rune count ---")
	// len() on a string returns the number of BYTES, not the number of
	// visible characters. This is a very common source of bugs when
	// slicing user text.
	fmt.Printf("len(%q) = %d bytes\n", word, len(word))
	// utf8.RuneCountInString counts actual Unicode code points (runes).
	fmt.Printf("utf8.RuneCountInString(%q) = %d runes\n", word, utf8.RuneCountInString(word))

	fmt.Println("--- indexing gives bytes, not characters ---")
	// word[0] yields a byte, not a rune. For the ASCII 'h' that byte
	// happens to equal the character, but this does not generalize once
	// multi-byte characters are involved.
	fmt.Printf("word[0] = %d which is byte %q\n", word[0], word[0])

	fmt.Println("--- ranging over a string yields runes ---")
	// range on a string decodes UTF-8 and gives you (byteIndex, rune)
	// pairs. Notice the byte index jumps by 2 after the multi-byte
	// character "é", because "é" occupies two bytes.
	for index, char := range word {
		fmt.Printf("byteIndex=%d rune=%q\n", index, char)
	}

	fmt.Println("--- converting between string, []byte, and []rune ---")
	// []byte(word) exposes the raw UTF-8 bytes; useful for I/O and
	// hashing. []rune(word) decodes it into one entry per character;
	// useful when you need to count, reverse, or index by character.
	bytes := []byte(word)
	runes := []rune(word)
	fmt.Printf("[]byte length = %d, []rune length = %d\n", len(bytes), len(runes))

	// Reversing by rune (not by byte) keeps multi-byte characters intact.
	reversed := make([]rune, len(runes))
	for i, r := range runes {
		reversed[len(runes)-1-i] = r
	}
	fmt.Printf("reversed(%q) = %q\n", word, string(reversed))

	fmt.Println("--- strings are immutable ---")
	// You cannot assign to a byte inside a string: word[0] = 'H' does not
	// compile. To change text you build a new string, often through the
	// strings package or a []byte/[]rune conversion.
	capitalized := strings.ToUpper(word[:1]) + word[1:]
	fmt.Printf("capitalized copy = %q (original unchanged: %q)\n", capitalized, word)

	fmt.Println("--- useful strings package helpers ---")
	sentence := "  Go is fun, Go is fast  "
	fmt.Printf("TrimSpace: %q\n", strings.TrimSpace(sentence))
	fmt.Printf("Contains %q: %t\n", "fun", strings.Contains(sentence, "fun"))
	fmt.Printf("Count of \"Go\": %d\n", strings.Count(sentence, "Go"))
	fmt.Printf("Split by comma: %q\n", strings.Split(strings.TrimSpace(sentence), ", "))
	// Building strings with + in a loop is inefficient because each
	// concatenation allocates a new string. strings.Builder amortizes
	// that cost by growing an internal buffer, similar to append on a
	// slice.
	var builder strings.Builder
	for i := 0; i < 3; i++ {
		fmt.Fprintf(&builder, "part%d ", i)
	}
	fmt.Printf("built with strings.Builder: %q\n", builder.String())

	fmt.Println("--- classifying runes ---")
	for _, r := range "Go2!" {
		fmt.Printf("%q letter=%t digit=%t punctuation=%t\n",
			r, unicode.IsLetter(r), unicode.IsDigit(r), unicode.IsPunct(r))
	}
}
