//go:build !aix && !android && !darwin && !dragonfly && !freebsd && !hurd && !illumos && !linux && !netbsd && !openbsd && !solaris

package markdown

// syncDirectory is a no-op on non-POSIX platforms (e.g. Windows), where
// directory handles cannot be fsynced; the save still durably fsyncs the
// temp file's contents before rename, so only the directory-entry update
// itself is not additionally flushed here.
func syncDirectory(string) error {
	return nil
}
