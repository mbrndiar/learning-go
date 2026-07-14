// Package contacts persists a small address book to a JSON file to practice
// typed and wrapped errors, defer-based resource cleanup, file I/O, and JSON
// validation.
//
// Implement every function and method below. Replace each
// panic("not implemented") with working code; do not change any signature.
package contacts

import "errors"

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
//
// TODO(task 1): return a message in the form "<Field>: <Message>".
func (e *ValidationError) Error() string {
	panic("not implemented")
}

// Validate checks that Name is non-empty, Email contains "@", and Age is
// not negative. It returns the first violation found, as a
// *ValidationError, or nil if the contact is valid. Check Name, then
// Email, then Age, in that order.
//
// TODO(task 2): implement Validate.
func (c Contact) Validate() error {
	panic("not implemented")
}

// ErrDuplicateEmail is returned by AddContact when a contact with the same
// email address already exists. Wrap it with fmt.Errorf and %w so callers
// can test for it with errors.Is.
var ErrDuplicateEmail = errors.New("duplicate email")

// ErrNotFound is returned by FindByEmail when no contact matches. Wrap it
// with fmt.Errorf and %w so callers can test for it with errors.Is.
var ErrNotFound = errors.New("contact not found")

// LoadContacts reads and decodes the JSON array of contacts stored at path.
//
// Open the file, defer closing it, and decode a []Contact from it. Wrap any
// error with context using fmt.Errorf and %w, preserving the underlying
// error so callers can use errors.Is(err, os.ErrNotExist) and
// errors.As for JSON decode errors.
//
// TODO(task 3): implement LoadContacts.
func LoadContacts(path string) ([]Contact, error) {
	panic("not implemented")
}

// SaveContacts validates every contact, then encodes contacts as indented
// JSON and writes it to path, creating or truncating the file.
//
// Validate before opening any file. Create the file, defer closing it, and
// wrap any I/O or encoding error with fmt.Errorf and %w.
//
// TODO(task 4): implement SaveContacts.
func SaveContacts(path string, contacts []Contact) error {
	panic("not implemented")
}

// AddContact validates c, appends it to the contacts stored at path, and
// saves the result.
//
// If the file at path does not exist, treat the address book as empty
// instead of failing (check the load error with errors.Is(err,
// os.ErrNotExist)). If a contact with the same email (case-sensitive)
// already exists, return an error that wraps ErrDuplicateEmail via %w
// without modifying the file.
//
// TODO(task 5): implement AddContact.
func AddContact(path string, c Contact) error {
	panic("not implemented")
}

// FindByEmail returns the first contact whose Email matches email
// (case-sensitive). If none match, it returns an error that wraps
// ErrNotFound via %w, including email in the message.
//
// TODO(task 6): implement FindByEmail.
func FindByEmail(contacts []Contact, email string) (Contact, error) {
	panic("not implemented")
}
