//go:build aix || android || darwin || dragonfly || freebsd || hurd || illumos || linux || netbsd || openbsd || solaris

package markdown

import "os"

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
