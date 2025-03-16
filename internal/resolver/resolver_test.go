package resolver

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveFilenames2(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_resolveFilenames2")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFile1 := filepath.Join(tempDir, "file1.txt")
	testFile2 := filepath.Join(tempDir, "file2.sh")
	testSubDir := filepath.Join(tempDir, "subdir")
	testFile3 := filepath.Join(testSubDir, "file3.sh")
	testFile4 := filepath.Join(testSubDir, "file4.xml")

	assert.NoError(t, os.WriteFile(testFile1, []byte("content1"), 0o600))
	assert.NoError(t, os.WriteFile(testFile2, []byte("content2"), 0o600))
	assert.NoError(t, os.Mkdir(testSubDir, 0o755))
	assert.NoError(t, os.WriteFile(testFile3, []byte("content3"), 0o600))
	assert.NoError(t, os.WriteFile(testFile4, []byte("content4"), 0o600))

	tests := []struct {
		name        string
		path        string
		recursive   bool
		expected    []string
		expectError bool
	}{
		{"Single file valid extension", testFile2, false, []string{testFile2}, false},
		{"Glob pattern valid extensions", filepath.Join(tempDir, "*.sh"), false, []string{testFile2}, false},
		{"Directory non-recursive", tempDir, false, []string{testFile2}, false},
		{"Directory recursive", tempDir, true, []string{testFile2, testFile3}, false},
		{"URL handling", "http://example.com/file.sh", false, []string{"http://example.com/file.sh"}, false},
		{"Nonexistent file", filepath.Join(tempDir, "nonexistent.txt"), false, nil, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := resolveFilenamesForPatterns(test.path, test.recursive)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, test.expected, result)
			}
		})
	}
}

func TestIsURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid HTTP URL", "http://example.com", true},
		{"Valid HTTPS URL", "https://example.com", true},
		{"Invalid URL", "example.com", false},
		{"Empty String", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsURL(tt.input))
		})
	}
}

func TestIgnoreFile(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		extensions []string
		expected   bool
	}{
		{"Allowed Extension", "file.sh", []string{".sh"}, false},
		{"Disallowed Extension", "file.txt", []string{".sh"}, true},
		{"Empty Extensions", "file.sh", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ignoreFile(tt.path, tt.extensions))
		})
	}
}

func TestResolveAllFiles(t *testing.T) {
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "file1.sh")
	file2 := filepath.Join(tempDir, "file2.sh")
	file3 := filepath.Join(tempDir, "ignore.txt")
	subDir := filepath.Join(tempDir, "subdir")
	subFile := filepath.Join(subDir, "file3.sh")

	assert.NoError(t, os.WriteFile(file1, []byte("test content"), 0o600))
	assert.NoError(t, os.WriteFile(file2, []byte("test content"), 0o600))
	assert.NoError(t, os.WriteFile(file3, []byte("test content"), 0o600))
	assert.NoError(t, os.Mkdir(subDir, 0o755))
	assert.NoError(t, os.WriteFile(subFile, []byte("test content"), 0o600))

	tests := []struct {
		name      string
		filenames []string
		recursive bool
		want      []string
		wantErr   bool
	}{
		{"Single file", []string{file1}, false, []string{file1}, false},
		{"Glob pattern", []string{filepath.Join(tempDir, "*.sh")}, false, []string{file1, file2}, false},
		{"Directory, non-recursive", []string{tempDir}, false, []string{file1, file2}, false},
		{"Directory, recursive", []string{tempDir}, true, []string{file1, file2, subFile}, false},
		{"If file passed explicitly, don't check its extension", []string{file3}, false, []string{file3}, false},
		{"URL as input", []string{"https://example.com/file.sh"}, false, []string{"https://example.com/file.sh"}, false},
		{"Invalid path", []string{"/invalid/path/file.sh"}, false, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveAllFiles(tt.filenames, tt.recursive)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.want, got)
			}
		})
	}
}
