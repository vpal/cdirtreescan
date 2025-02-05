package filetypes

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/vpal/cdirtreescan/scan"
)

func findDeviceFile(t *testing.T, expectedType FileType) (string, bool) {
	if expectedType == FileTypeCharDevice {
		// Use fixed list for character devices
		candidates := []string{"/dev/null", "/dev/zero", "/dev/tty"}
		for _, path := range candidates {
			info, err := os.Stat(path)
			if err != nil {
				continue
			}
			if info.Mode()&os.ModeCharDevice != 0 {
				return path, true
			}
		}
	} else if expectedType == FileTypeBlockDevice {
		// Scan /dev for block devices
		entries, err := os.ReadDir("/dev")
		if err != nil {
			t.Logf("Failed to read /dev directory: %v", err)
			return "", false
		}

		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if info.Mode()&os.ModeDevice != 0 && info.Mode()&os.ModeCharDevice == 0 {
				return filepath.Join("/dev", entry.Name()), true
			}
		}
	}

	t.Logf("No device of type %d found in common locations, skipping test", expectedType)
	return "", false
}

func TestGetFileType(t *testing.T) {
	root := t.TempDir()

	// Create and test files one by one
	tests := []struct {
		name     string
		setup    func(t *testing.T) (string, FileType, bool) // returns path, expected type, and if test should be skipped
	}{
		{
			name: "dir1",
			setup: func(t *testing.T) (string, FileType, bool) {
				path := filepath.Join(root, "dir1")
				err := os.Mkdir(path, 0755)
				return path, FileTypeDirectory, err != nil
			},
		},
		{
			name: "file1.txt",
			setup: func(t *testing.T) (string, FileType, bool) {
				path := filepath.Join(root, "file1.txt")
				_, err := os.Create(path)
				return path, FileTypeRegular, err != nil
			},
		},
		{
			name: "symlink",
			setup: func(t *testing.T) (string, FileType, bool) {
				path := filepath.Join(root, "symlink")
				err := os.Symlink(filepath.Join(root, "file1.txt"), path)
				return path, FileTypeSymlink, err != nil
			},
		},
		{
			name: "socket",
			setup: func(t *testing.T) (string, FileType, bool) {
				path := filepath.Join(root, "socket")
				err := syscall.Mknod(path, syscall.S_IFSOCK|0666, 0)
				return path, FileTypeSocket, err != nil
			},
		},
		{
			name: "named_pipe",
			setup: func(t *testing.T) (string, FileType, bool) {
				path := filepath.Join(root, "named_pipe")
				err := syscall.Mkfifo(path, 0666)
				return path, FileTypeNamedPipe, err != nil
			},
		},
		{
			name: "char_device",
			setup: func(t *testing.T) (string, FileType, bool) {
				if path, ok := findDeviceFile(t, FileTypeCharDevice); ok {
					return path, FileTypeCharDevice, false
				}
				return "", FileTypeCharDevice, true
			},
		},
		{
			name: "block_device",
			setup: func(t *testing.T) (string, FileType, bool) {
				if path, ok := findDeviceFile(t, FileTypeBlockDevice); ok {
					return path, FileTypeBlockDevice, false
				}
				return "", FileTypeBlockDevice, true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, expectedType, skip := tt.setup(t)
			if skip {
				t.Skip("Skipping test due to setup failure")
				return
			}

			entry, err := os.ReadDir(filepath.Dir(path))
			if err != nil {
				t.Fatalf("Failed to read directory: %v", err)
			}

			var testEntry os.DirEntry
			for _, e := range entry {
				if e.Name() == filepath.Base(path) {
					testEntry = e
					break
				}
			}

			if testEntry == nil {
				t.Fatalf("Could not find test file %s in directory entries", path)
			}

			pathEntry := scan.PathEntry{
				Path:  path,
				Entry: testEntry,
			}

			result := GetFileType(pathEntry)
			if result != expectedType {
				t.Errorf("GetFileType(%s) = %v, want %v", tt.name, result, expectedType)
			}
		})
	}
}
