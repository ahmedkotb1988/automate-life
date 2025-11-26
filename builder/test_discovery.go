package builder

import (
	"automateLife/ui"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// TestFile represents a discovered test file
type TestFile struct {
	Path        string
	PackageName string
	RelativePath string
}

// DiscoverTests finds all *_test.go files in a directory recursively
func DiscoverTests(rootDir string) ([]TestFile, error) {
	var testFiles []TestFile

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor, node_modules, and hidden directories
		if info.IsDir() {
			name := info.Name()
			if name == "vendor" || name == "node_modules" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if it's a test file
		if strings.HasSuffix(info.Name(), "_test.go") {
			relPath, _ := filepath.Rel(rootDir, path)

			testFiles = append(testFiles, TestFile{
				Path:         path,
				RelativePath: relPath,
				PackageName:  extractPackageName(path),
			})
		}

		return nil
	})

	return testFiles, err
}

// extractPackageName reads the package name from a Go file
func extractPackageName(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	// Read first 512 bytes to find package declaration
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	content := string(buf[:n])

	// Look for "package <name>"
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}

	return ""
}

// CreateUnifiedTestSuite creates a temporary directory with all test files
func CreateUnifiedTestSuite(testFiles []TestFile, baseDir string) (string, error) {
	// Create a temporary directory for unified tests
	unifiedDir := filepath.Join(baseDir, ".unified_tests")

	// Clean up if it already exists
	if _, err := os.Stat(unifiedDir); err == nil {
		os.RemoveAll(unifiedDir)
	}

	if err := os.MkdirAll(unifiedDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create unified test directory: %w", err)
	}

	ui.Info(fmt.Sprintf("Creating unified test suite in: %s", unifiedDir))
	ui.Info(fmt.Sprintf("Discovered %d test files", len(testFiles)))

	// Copy each test file to the unified directory
	for i, testFile := range testFiles {
		// Create subdirectories to maintain structure
		destPath := filepath.Join(unifiedDir, testFile.RelativePath)
		destDir := filepath.Dir(destPath)

		if err := os.MkdirAll(destDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create directory %s: %w", destDir, err)
		}

		// Copy the file
		if err := copyFile(testFile.Path, destPath); err != nil {
			return "", fmt.Errorf("failed to copy test file %s: %w", testFile.Path, err)
		}

		ui.Success(fmt.Sprintf("  [%d/%d] %s (%s)", i+1, len(testFiles), testFile.RelativePath, testFile.PackageName))
	}

	return unifiedDir, nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}

// CleanupUnifiedTestSuite removes the unified test directory
func CleanupUnifiedTestSuite(baseDir string) error {
	unifiedDir := filepath.Join(baseDir, ".unified_tests")
	if _, err := os.Stat(unifiedDir); err == nil {
		return os.RemoveAll(unifiedDir)
	}
	return nil
}
