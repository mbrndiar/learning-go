// Package contacts is the reference implementation for the errors, files,
// and JSON exercise. See ../contacts.go for the task descriptions.
package contacts

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Contact is one address-book entry.
type Contact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// ValidationError reports which field of a Contact failed validation and
// why. It is returned as a *ValidationError so callers can recover it with
// errors.As.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface for ValidationError.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validate checks that Name is non-empty, Email contains "@", and Age is
// not negative. It returns the first violation found, as a
// *ValidationError, or nil if the contact is valid.
func (c Contact) Validate() error {
	if c.Name == "" {
		return &ValidationError{Field: "Name", Message: "must not be empty"}
	}
	if !strings.Contains(c.Email, "@") {
		return &ValidationError{Field: "Email", Message: "must contain @"}
	}
	if c.Age < 0 {
		return &ValidationError{Field: "Age", Message: "must not be negative"}
	}
	return nil
}

// ErrDuplicateEmail is returned by AddContact when a contact with the same
// email address already exists.
var ErrDuplicateEmail = errors.New("duplicate email")

// ErrNotFound is returned by FindByEmail when no contact matches.
var ErrNotFound = errors.New("contact not found")

// LoadContacts reads and decodes the JSON array of contacts stored at path.
func LoadContacts(path string) ([]Contact, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("load contacts from %s: %w", path, err)
	}
	defer file.Close()

	var contacts []Contact
	if err := json.NewDecoder(file).Decode(&contacts); err != nil {
		return nil, fmt.Errorf("load contacts from %s: %w", path, err)
	}
	return contacts, nil
}

// SaveContacts validates every contact, then encodes contacts as indented
// JSON and writes it to path, creating or truncating the file.
func SaveContacts(path string, contacts []Contact) error {
	for _, c := range contacts {
		if err := c.Validate(); err != nil {
			return fmt.Errorf("save contacts to %s: %w", path, err)
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("save contacts to %s: %w", path, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(contacts); err != nil {
		return fmt.Errorf("save contacts to %s: %w", path, err)
	}
	return nil
}

// AddContact validates c, appends it to the contacts stored at path, and
// saves the result.
func AddContact(path string, c Contact) error {
	if err := c.Validate(); err != nil {
		return fmt.Errorf("add contact: %w", err)
	}

	existing, err := LoadContacts(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("add contact: %w", err)
		}
		existing = nil
	}

	for _, other := range existing {
		if other.Email == c.Email {
			return fmt.Errorf("add contact %s: %w", c.Email, ErrDuplicateEmail)
		}
	}

	existing = append(existing, c)
	return SaveContacts(path, existing)
}

// FindByEmail returns the first contact whose Email matches email
// (case-sensitive). If none match, it returns an error that wraps
// ErrNotFound.
func FindByEmail(contacts []Contact, email string) (Contact, error) {
	for _, c := range contacts {
		if c.Email == email {
			return c, nil
		}
	}
	return Contact{}, fmt.Errorf("find contact %s: %w", email, ErrNotFound)
}
