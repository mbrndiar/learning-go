//go:build !aix && !android && !darwin && !dragonfly && !freebsd && !hurd && !illumos && !linux && !netbsd && !openbsd && !solaris

package markdown

// Directory syncing is unavailable or unsupported on these platforms.
func syncDirectory(string) error {
	return nil
}
