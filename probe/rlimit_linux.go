package probe

import "golang.org/x/sys/unix"

func applyResourceLimits(pid int) {
	// Limit open file descriptors to prevent FD exhaustion.
	unix.Prlimit(pid, unix.RLIMIT_NOFILE, &unix.Rlimit{Cur: 1024, Max: 1024}, nil)

	// Limit file size to 50 MB to prevent disk filling.
	unix.Prlimit(pid, unix.RLIMIT_FSIZE, &unix.Rlimit{Cur: 50 * 1024 * 1024, Max: 50 * 1024 * 1024}, nil)
}
