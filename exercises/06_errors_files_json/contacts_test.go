package contacts

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestContactValidate(t *testing.T) {
	tests := []struct {
		name      string
		contact   Contact
		wantField string
		wantErr   bool
	}{
		{"valid", Contact{Name: "Ada Lovelace", Email: "ada@example.com", Age: 36}, "", false},
		{"missing name", Contact{Name: "", Email: "ada@example.com", Age: 36}, "Name", true},
		{"invalid email", Contact{Name: "Ada", Email: "ada-example.com", Age: 36}, "Email", true},
		{"negative age", Contact{Name: "Ada", Email: "ada@example.com", Age: -1}, "Age", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.contact.Validate()
			if !tt.wantErr {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			var valErr *ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("Validate() error = %v, want *ValidationError", err)
			}
			if valErr.Field != tt.wantField {
				t.Errorf("ValidationError.Field = %q, want %q", valErr.Field, tt.wantField)
			}
		})
	}
}

func TestSaveAndLoadContacts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "contacts.json")

	want := []Contact{
		{Name: "Ada Lovelace", Email: "ada@example.com", Age: 36},
		{Name: "Grace Hopper", Email: "grace@example.com", Age: 85},
	}

	if err := SaveContacts(path, want); err != nil {
		t.Fatalf("SaveContacts() error = %v", err)
	}

	got, err := LoadContacts(path)
	if err != nil {
		t.Fatalf("LoadContacts() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("LoadContacts() = %+v, want %+v", got, want)
	}
}

func TestSaveContactsValidation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "contacts.json")

	err := SaveContacts(path, []Contact{{Name: "", Email: "bad", Age: 1}})
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("SaveContacts() error = %v, want *ValidationError", err)
	}
	if _, statErr := os.Stat(path); statErr == nil {
		t.Errorf("SaveContacts() should not create %s when validation fails", path)
	}
}

func TestLoadContactsMissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.json")

	_, err := LoadContacts(path)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("LoadContacts() error = %v, want errors.Is(err, os.ErrNotExist)", err)
	}
}

func TestLoadContactsInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.json")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := LoadContacts(path)
	if err == nil {
		t.Fatal("LoadContacts() error = nil, want a JSON decode error")
	}
	var syntaxErr *json.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Fatalf("LoadContacts() error = %v, want errors.As to find a *json.SyntaxError", err)
	}
}

func TestAddContactDuplicate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "contacts.json")

	first := Contact{Name: "Ada Lovelace", Email: "ada@example.com", Age: 36}
	if err := AddContact(path, first); err != nil {
		t.Fatalf("AddContact() first call error = %v", err)
	}

	second := Contact{Name: "Ada Again", Email: "ada@example.com", Age: 40}
	err := AddContact(path, second)
	if !errors.Is(err, ErrDuplicateEmail) {
		t.Fatalf("AddContact() error = %v, want errors.Is(err, ErrDuplicateEmail)", err)
	}

	got, loadErr := LoadContacts(path)
	if loadErr != nil {
		t.Fatalf("LoadContacts() error = %v", loadErr)
	}
	if len(got) != 1 {
		t.Fatalf("LoadContacts() returned %d contacts, want 1", len(got))
	}
}

func TestAddContactCreatesMissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new", "contacts.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	c := Contact{Name: "Ada Lovelace", Email: "ada@example.com", Age: 36}
	if err := AddContact(path, c); err != nil {
		t.Fatalf("AddContact() error = %v", err)
	}

	got, err := LoadContacts(path)
	if err != nil {
		t.Fatalf("LoadContacts() error = %v", err)
	}
	if len(got) != 1 || got[0].Email != c.Email {
		t.Fatalf("LoadContacts() = %+v, want one contact with email %q", got, c.Email)
	}
}

func TestFindByEmail(t *testing.T) {
	contacts := []Contact{
		{Name: "Ada Lovelace", Email: "ada@example.com", Age: 36},
		{Name: "Grace Hopper", Email: "grace@example.com", Age: 85},
	}

	t.Run("found", func(t *testing.T) {
		got, err := FindByEmail(contacts, "grace@example.com")
		if err != nil {
			t.Fatalf("FindByEmail() error = %v", err)
		}
		if got.Name != "Grace Hopper" {
			t.Errorf("FindByEmail() = %+v, want Grace Hopper", got)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := FindByEmail(contacts, "missing@example.com")
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("FindByEmail() error = %v, want errors.Is(err, ErrNotFound)", err)
		}
	})
}
