//go:build linux

package utils

import (
	"fmt"
	"os"
)

// MoveDir moves a directory from src to dst.
// Strategy:
//  1. Try os.Rename (fast, same-volume)
//  2. Fall back to recursive copy; originals are left in src (tmp) so the OS
//     can clean them up later.
func MoveDir(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	if err := CopyDir(src, dst); err != nil {
		return fmt.Errorf("failed to copy %s -> %s: %w", src, dst, err)
	}

	return nil
}
