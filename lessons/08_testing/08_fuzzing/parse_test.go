package fuzzing

import "testing"

func TestParseKeyValue(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantKey   string
		wantValue string
		wantErr   bool
	}{
		{name: "simple", input: "host=localhost", wantKey: "host", wantValue: "localhost"},
		{name: "spaces trimmed", input: "  host = localhost  ", wantKey: "host", wantValue: "localhost"},
		{name: "empty value", input: "flag=", wantKey: "flag", wantValue: ""},
		{name: "value contains equals", input: "query=a=b", wantKey: "query", wantValue: "a=b"},
		{name: "no separator", input: "no-equals-here", wantErr: true},
		{name: "empty key", input: "=value", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key, value, err := ParseKeyValue(test.input)

			if test.wantErr {
				if err == nil {
					t.Fatalf("ParseKeyValue(%q) = (%q, %q, nil), want an error", test.input, key, value)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseKeyValue(%q) returned unexpected error: %v", test.input, err)
			}
			if key != test.wantKey || value != test.wantValue {
				t.Errorf("ParseKeyValue(%q) = (%q, %q), want (%q, %q)", test.input, key, value, test.wantKey, test.wantValue)
			}
		})
	}
}

// FuzzParseKeyValue is a fuzz target: a function named FuzzXxx taking
// *testing.F. f.Add registers seed inputs that always run, including with
// plain `go test`. f.Fuzz registers the property checked against both the
// seeds and (only under `go test -fuzz`) randomly generated inputs.
//
// Regression seeds - inputs that previously caused a failure - are stored
// as files under testdata/fuzz/FuzzParseKeyValue/. go test always replays
// them too, so a bug that fuzzing once found can never silently come back.
func FuzzParseKeyValue(f *testing.F) {
	for _, seed := range []string{
		"host=localhost",
		"  spaced = value  ",
		"flag=",
		"a=b=c",
		"=novalue",
		"novalue",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		key, value, err := ParseKeyValue(input)
		if err != nil {
			// An error is a perfectly valid outcome for malformed input;
			// there is nothing further to check.
			return
		}

		if key == "" {
			t.Fatalf("ParseKeyValue(%q) returned an empty key with a nil error", input)
		}

		// Invariant: re-parsing "key=value" must reproduce the same key
		// and value. This is the kind of property-style check fuzzing is
		// good at: it does not need to predict every input, only verify a
		// rule that must always hold.
		reconstructed := key + "=" + value
		gotKey, gotValue, err := ParseKeyValue(reconstructed)
		if err != nil {
			t.Fatalf("re-parsing reconstructed %q failed: %v", reconstructed, err)
		}
		if gotKey != key || gotValue != value {
			t.Fatalf("round-trip mismatch for %q: reconstructed %q parsed back as (%q, %q), want (%q, %q)",
				input, reconstructed, gotKey, gotValue, key, value)
		}
	})
}
