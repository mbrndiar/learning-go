//go:build !(aix || android || darwin || dragonfly || freebsd || illumos || ios || linux || netbsd || openbsd || solaris)

package contract

import (
	"os"
	"os/exec"
)

func configureProcessGroup(*exec.Cmd) {}

func terminateProcessGroup(pid int) {
	process, err := os.FindProcess(pid)
	if err == nil {
		_ = process.Kill()
	}
}
