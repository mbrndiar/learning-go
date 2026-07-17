package markdown

import "testing"

// FuzzParseDocument exercises parseDocument directly, since it is a pure
// function of its input bytes with no filesystem or context dependencies.
// Every input must parse without panicking, and every successfully parsed
// document must have valid, unique, ascending, positive task IDs, a
// coherent next-id, and titles that pass the project's own validation.
func FuzzParseDocument(f *testing.F) {
	f.Add([]byte("<!-- rest-task-api:v1 next-id=1 -->\n# Tasks\n\n"))
	f.Add([]byte("<!-- rest-task-api:v1 next-id=3 -->\n# Tasks\n\n" +
		"- [ ] 1: literal *Markdown*\n- [x] 2: second\n"))
	f.Add([]byte(""))
	f.Add([]byte{0xff, '\n'})
	f.Add([]byte("<!-- rest-task-api:v1 next-id=1 -->\n# Tasks\n"))
	f.Add([]byte("<!-- rest-task-api:v1 next-id=1 -->\n# Tasks\n\n\n"))
	f.Add([]byte("<!-- rest-task-api:v01 next-id=1 -->\n# Tasks\n\n"))
	f.Add([]byte("<!-- rest-task-api:v1 next-id=0 -->\n# Tasks\n\n"))
	f.Add([]byte("<!-- rest-task-api:v1 next-id=2 -->\n# Tasks\n\n- [ ] 0: title\n"))
	f.Add([]byte("<!-- rest-task-api:v1 next-id=3 -->\n# Tasks\n\n- [ ] 1: one\n- [x] 1: two\n"))
	f.Add([]byte("<!-- rest-task-api:v1 next-id=3 -->\n# Tasks\n\n- [ ] 2: two\n- [x] 1: one\n"))
	f.Add([]byte("<!-- rest-task-api:v1 next-id=2 -->\n# Tasks\n\n- [ ] 1: trailing \n"))
	f.Add([]byte("<!-- rest-task-api:v1 next-id=9223372036854775807 -->\n# Tasks\n\n" +
		"- [x] 9223372036854775806: near max\n"))
	f.Fuzz(func(t *testing.T, content []byte) {
		document, err := parseDocument(content)
		if err != nil {
			return
		}

		if document.NextID <= 0 {
			t.Fatalf("parsed next-id %d is not positive", document.NextID)
		}
		seen := make(map[int64]bool, len(document.Tasks))
		var previousID int64
		for _, item := range document.Tasks {
			if item.ID <= 0 {
				t.Fatalf("parsed task ID %d is not positive", item.ID)
			}
			if seen[item.ID] {
				t.Fatalf("parsed task ID %d is not unique", item.ID)
			}
			if item.ID <= previousID {
				t.Fatalf("parsed task IDs are not strictly ascending: %d after %d", item.ID, previousID)
			}
			seen[item.ID] = true
			previousID = item.ID
		}
		if document.NextID <= previousID {
			t.Fatalf("next-id %d is not greater than every task ID (max %d)", document.NextID, previousID)
		}

		// A successfully parsed document must round-trip: serializing it and
		// parsing the result again must reproduce the same document.
		reparsed, err := parseDocument([]byte(serialize(document)))
		if err != nil {
			t.Fatalf("round-trip parse failed for %q: %v", serialize(document), err)
		}
		if reparsed.NextID != document.NextID || len(reparsed.Tasks) != len(document.Tasks) {
			t.Fatalf("round-trip mismatch: got %+v, want %+v", reparsed, document)
		}
		for index := range document.Tasks {
			if reparsed.Tasks[index] != document.Tasks[index] {
				t.Fatalf("round-trip task mismatch at %d: got %+v, want %+v",
					index, reparsed.Tasks[index], document.Tasks[index])
			}
		}
	})
}
