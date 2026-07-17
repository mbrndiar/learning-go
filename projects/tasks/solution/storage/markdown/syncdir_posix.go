//go:build aix || android || darwin || dragonfly || freebsd || hurd || illumos || linux || netbsd || openbsd || solaris

package markdown

import "os"

// syncDirectory persists a directory's entries (e.g. after a rename) on
// POSIX platforms, where fsync on an open directory handle is supported and
// required for a rename to survive a crash.
func syncDirectory(path string) (err error) {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := directory.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()
	return directory.Sync()
}
