package task

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// NormalizeTitle trims and validates a title.
func NormalizeTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if !utf8.ValidString(title) {
		return "", validationError("title", "title must contain valid UTF-8")
	}
	count := utf8.RuneCountInString(title)
	if count < 1 || count > MaxTitleLength {
		return "", validationError("title", "title must contain between 1 and 120 characters")
	}
	for _, r := range title {
		if isPhysicalLineBreak(r) {
			return "", validationError("title", "title must occupy one physical line")
		}
		if unicode.IsControl(r) {
			return "", validationError("title", "title must not contain control characters")
		}
	}
	return title, nil
}

// ValidateTitle reports whether a title is already normalized and valid.
func ValidateTitle(title string) error {
	normalized, err := NormalizeTitle(title)
	if err != nil {
		return err
	}
	if normalized != title {
		return validationError("title", "title must not have leading or trailing whitespace")
	}
	return nil
}

// ValidateID requires a positive task ID.
func ValidateID(id int64) error {
	if id <= 0 {
		return validationError("id", "task ID must be a positive integer")
	}
	return nil
}

// NormalizeUpdate validates a partial update and returns normalized copies.
func NormalizeUpdate(input UpdateInput) (UpdateInput, error) {
	if input.Title == nil && input.Completed == nil {
		return UpdateInput{}, validationError("update", "update must include title or completed")
	}

	var normalized UpdateInput
	if input.Title != nil {
		title, err := NormalizeTitle(*input.Title)
		if err != nil {
			return UpdateInput{}, err
		}
		normalized.Title = &title
	}
	if input.Completed != nil {
		completed := *input.Completed
		normalized.Completed = &completed
	}
	return normalized, nil
}

// ValidateUpdate reports whether a partial update is already normalized.
func ValidateUpdate(input UpdateInput) error {
	normalized, err := NormalizeUpdate(input)
	if err != nil {
		return err
	}
	if input.Title != nil && *normalized.Title != *input.Title {
		return validationError("title", "title must not have leading or trailing whitespace")
	}
	return nil
}

// NormalizeListFilter copies an optional filter without losing explicit false.
func NormalizeListFilter(filter ListFilter) (ListFilter, error) {
	if filter.Completed == nil {
		return ListFilter{}, nil
	}
	completed := *filter.Completed
	return ListFilter{Completed: &completed}, nil
}

// ValidateListFilter validates a completion filter.
func ValidateListFilter(filter ListFilter) error {
	_, err := NormalizeListFilter(filter)
	return err
}

// ValidateTask validates a task value returned by storage or a remote server.
func ValidateTask(value Task) error {
	if err := ValidateID(value.ID); err != nil {
		return err
	}
	return ValidateTitle(value.Title)
}

func validationError(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}

func isPhysicalLineBreak(r rune) bool {
	switch r {
	case '\n', '\v', '\f', '\r', '\u001c', '\u001d', '\u001e', '\u0085', '\u2028', '\u2029':
		return true
	default:
		return false
	}
}
