//go:build windows

package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	if err := copyDir(src, dst); err != nil {
		return fmt.Errorf("failed to copy %s -> %s: %w", src, dst, err)
	}
	return nil
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
