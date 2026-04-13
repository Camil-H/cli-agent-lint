//go:build !linux

package probe

func applyResourceLimits(pid int) {}
