package task

// NormalizeTitle trims title and returns it if it then satisfies the Task
// title contract, or a *ValidationError otherwise.
func NormalizeTitle(title string) (string, error) {
	return "", ErrNotImplemented
}

// ValidateTitle reports whether title already satisfies the Task title
// contract without normalizing it.
func ValidateTitle(title string) error {
	return ErrNotImplemented
}

// ValidateID reports whether id is a valid positive task identifier.
func ValidateID(id int64) error {
	return ErrNotImplemented
}

// NormalizeUpdate normalizes the fields present in input and returns the
// result if the update as a whole satisfies the Task update contract.
func NormalizeUpdate(input UpdateInput) (UpdateInput, error) {
	return UpdateInput{}, ErrNotImplemented
}

// ValidateUpdate reports whether the fields present in input already satisfy
// the Task update contract without normalizing them.
func ValidateUpdate(input UpdateInput) error {
	return ErrNotImplemented
}

// NormalizeListFilter normalizes filter and returns the result if it
// satisfies the Task filter contract.
func NormalizeListFilter(filter ListFilter) (ListFilter, error) {
	return ListFilter{}, ErrNotImplemented
}

// ValidateListFilter reports whether filter already satisfies the Task
// filter contract without normalizing it.
func ValidateListFilter(filter ListFilter) error {
	return ErrNotImplemented
}

// ValidateTask reports whether a fully constructed Task satisfies the
// invariants a repository is allowed to rely on.
func ValidateTask(value Task) error {
	return ErrNotImplemented
}
