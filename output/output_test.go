package output

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/vpal/cdirtreescan/scan"
)

func setupTestFiles(t *testing.T, root string) (map[string]string, error) {
	paths := make(map[string]string)

	// Create basic test files only
	os.Mkdir(filepath.Join(root, "dir1"), 0755)
	os.Create(filepath.Join(root, "file1.txt"))
	os.Symlink(filepath.Join(root, "file1.txt"), filepath.Join(root, "symlink"))
	socketPath := filepath.Join(root, "socket")
	syscall.Mknod(socketPath, syscall.S_IFSOCK|0666, 0)
	namedPipePath := filepath.Join(root, "named_pipe")
	syscall.Mkfifo(namedPipePath, 0666)

	// Create nested structure
	os.Mkdir(filepath.Join(root, "dir1", "subdir1"), 0755)
	os.Create(filepath.Join(root, "dir1", "subdir1", "file2.txt"))

	paths["dir"] = filepath.Join(root, "dir1")
	paths["file"] = filepath.Join(root, "file1.txt")
	paths["symlink"] = filepath.Join(root, "symlink")
	paths["socket"] = socketPath
	paths["pipe"] = namedPipePath

	return paths, nil
}

func TestDirTreePrinter_PrintCount(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	paths, err := setupTestFiles(t, root)
	if err != nil {
		t.Fatalf("Failed to setup test files: %v", err)
	}

	dts, err := scan.NewDirTreeScanner(ctx, root, 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var outBuf, errBuf bytes.Buffer
	dtp := NewDirTreePrinter(dts, &outBuf, &errBuf, true)

	if err := dtp.PrintCount(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Basic expectations only
	expectedOutput := []string{
		"Regular files: 2",
		"Directories: 3",
		"Symbolic links: 1",
		"Sockets: 1",
		"FIFOs (named pipe): 1",
	}

	for _, eo := range expectedOutput {
		if !strings.Contains(outBuf.String(), eo) {
			t.Errorf("expected output to contain %q, but it didn't", eo)
		}
	}
}

func TestDirTreePrinter_PrintList(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	paths, err := setupTestFiles(t, root)
	if err != nil {
		t.Fatalf("Failed to setup test files: %v", err)
	}

	dts, err := scan.NewDirTreeScanner(ctx, root, 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var outBuf, errBuf bytes.Buffer
	dtp := NewDirTreePrinter(dts, &outBuf, &errBuf, true)

	err = dtp.PrintList()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedOutput := []string{
		"d " + root,
		"d " + filepath.Join(root, "dir1"),
		"- " + filepath.Join(root, "file1.txt"),
		"l " + filepath.Join(root, "symlink"),
		"s " + paths["socket"],
		"p " + paths["pipe"],
		"d " + filepath.Join(root, "dir1", "subdir1"),
		"- " + filepath.Join(root, "dir1", "subdir1", "file2.txt"),
	}

	for _, eo := range expectedOutput {
		if !strings.Contains(outBuf.String(), eo) {
			t.Errorf("expected output to contain %q, but it didn't", eo)
		}
	}
}
