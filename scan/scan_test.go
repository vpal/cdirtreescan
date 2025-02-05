package scan

import (
	"context"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestNewDirTreeScanner(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	_, err := NewDirTreeScanner(ctx, root, 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDirTreeScanner_Stream(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	// Create test files and directories
	os.Mkdir(filepath.Join(root, "dir1"), 0755)
	os.Create(filepath.Join(root, "file1.txt"))
	os.Symlink(filepath.Join(root, "file1.txt"), filepath.Join(root, "symlink"))
	socketPath := filepath.Join(root, "socket")
	syscall.Mknod(socketPath, syscall.S_IFSOCK|0666, 0)
	namedPipePath := filepath.Join(root, "named_pipe")
	syscall.Mkfifo(namedPipePath, 0666)
	blockDevicePath := filepath.Join(root, "block_device")
	syscall.Mknod(blockDevicePath, syscall.S_IFBLK|0666, int((7<<8)|0))
	charDevicePath := filepath.Join(root, "char_device")
	syscall.Mknod(charDevicePath, syscall.S_IFCHR|0666, int((7<<8)|0))
	os.Create(filepath.Join(root, "other"))

	// Create nested directories and files
	os.Mkdir(filepath.Join(root, "dir1", "subdir1"), 0755)
	os.Create(filepath.Join(root, "dir1", "subdir1", "file2.txt"))

	dts, err := NewDirTreeScanner(ctx, root, 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	entryCh, errCh := dts.Stream()

	for entries := range entryCh {
		for _, entry := range entries {
			if entry.Path == "" {
				t.Errorf("expected valid path, got empty string")
			}
		}
	}

	for err := range errCh {
		t.Errorf("unexpected error: %v", err)
	}
}
