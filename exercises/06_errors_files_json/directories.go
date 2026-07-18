package contacts

import "errors"

// ErrNotDirectory reports a path that exists but is not a directory.
var ErrNotDirectory = errors.New("not a directory")

// EnsureWorkspace creates the inbox, archive, and reports/daily directories
// below root.
//
// TODO(task 10): implement EnsureWorkspace with os.MkdirAll.
func EnsureWorkspace(root string) error {
	panic("not implemented")
}

// ListRegularFiles recursively returns regular files below root as relative,
// slash-separated paths in deterministic order.
//
// TODO(task 11): implement ListRegularFiles with filepath.WalkDir. Return an
// error wrapping ErrNotDirectory when root exists but is not a directory.
func ListRegularFiles(root string) ([]string, error) {
	panic("not implemented")
}

// MoveFile moves source to destination, creating destination's parent
// directory first.
//
// TODO(task 12): implement MoveFile with os.MkdirAll and os.Rename.
func MoveFile(source, destination string) error {
	panic("not implemented")
}

// RemoveEmptyDirectory removes path only when it is a directory and empty.
//
// TODO(task 13): implement RemoveEmptyDirectory with os.Stat and os.Remove.
// Return an error wrapping ErrNotDirectory when path is not a directory.
func RemoveEmptyDirectory(path string) error {
	panic("not implemented")
}
