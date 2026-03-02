//go:build windows

package utils

import (
	"fmt"
	"os"
	"syscall"
)

// MoveDir moves a directory from src to dst.
// Strategy:
//  1. Try os.Rename (fast, same-volume)
//  2. Try syscall.MoveFile (works cross-volume on Windows)
//  3. Fall back to recursive copy; originals are left in src (tmp) so the OS
//     can clean them up later.
func MoveDir(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	from, err := syscall.UTF16PtrFromString(src)
	if err == nil {
		to, err := syscall.UTF16PtrFromString(dst)
		if err == nil {
			if err := syscall.MoveFile(from, to); err == nil {
				return nil
			}
		}
	}

	// Fall back to copy; leave originals in tmp.
	if err := CopyDir(src, dst); err != nil {
		return fmt.Errorf("failed to copy %s -> %s: %w", src, dst, err)
	}
	return nil
}
