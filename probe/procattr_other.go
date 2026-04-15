//go:build !linux

package probe

import "syscall"

func sysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setpgid: true, // Kill the whole process group on timeout.
	}
}
