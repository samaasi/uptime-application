package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/samaasi/uptime-application/services/api-services/internal/utils"
)

const LocalStorageName = "local"

// LocalStorageDriver implements StorageDriver for local disk storage.
type LocalStorageDriver struct {
	basePath string // Base path for storing files (e.g., "/var/www/assets")
	baseURL  string // Base URL for accessing files (e.g., "http://localhost:5005/assets")
}

// NewLocalStorageDriver creates a new LocalStorageDriver instance.
func NewLocalStorageDriver(basePath, baseURL string) (*LocalStorageDriver, error) {
	if basePath == "" {
		return nil, fmt.Errorf("local storage base path cannot be empty")
	}
	if baseURL == "" {
		return nil, fmt.Errorf("local storage base URL cannot be empty")
	}

	if !filepath.IsAbs(basePath) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		basePath = filepath.Join(wd, basePath)
	}

	basePath = filepath.Clean(basePath)
	baseURL = strings.TrimRight(baseURL, "/") + "/"

	if err := os.MkdirAll(basePath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create local storage base path '%s': %w", basePath, err)
	}

	return &LocalStorageDriver{
		basePath: basePath,
		baseURL:  baseURL,
	}, nil
}

// Upload saves the given data to the specified key on the local disk.
func (l *LocalStorageDriver) Upload(ctx context.Context, key string, data io.Reader, mimeType string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("key cannot be empty")
	}

	key, err := sanitizeKey(key)
	if err != nil {
		return "", fmt.Errorf("invalid key: %w", err)
	}

	fullPath := filepath.Join(l.basePath, key)
	if err := verifyPathWithinBase(l.basePath, fullPath); err != nil {
		return "", fmt.Errorf("path verification failed: %w", err)
	}

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := openFileSecurely(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, data)
	if err != nil {
		// Attempt to remove the file if writing failed
		_ = os.Remove(fullPath)
		return "", fmt.Errorf("failed to write data to file: %w", err)
	}

	publicURL := l.baseURL + url.PathEscape(key)
	return publicURL, nil
}

// Download retrieves the data for the given key from the local disk.
func (l *LocalStorageDriver) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if key == "" {
		return nil, fmt.Errorf("key cannot be empty")
	}

	key, err := sanitizeKey(key)
	if err != nil {
		return nil, fmt.Errorf("invalid key: %w", err)
	}

	fullPath := filepath.Join(l.basePath, key)

	if err := verifyPathWithinBase(l.basePath, fullPath); err != nil {
		return nil, fmt.Errorf("path verification failed: %w", err)
	}

	file, err := openFileSecurely(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("asset not found")
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

// Delete removes the asset at the specified key from the local disk.
func (l *LocalStorageDriver) Delete(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	key, err := sanitizeKey(key)
	if err != nil {
		return fmt.Errorf("invalid key: %w", err)
	}

	fullPath := filepath.Join(l.basePath, key)
	if err := verifyPathWithinBase(l.basePath, fullPath); err != nil {
		return fmt.Errorf("path verification failed: %w", err)
	}

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// Exists checks if an asset with the given key exists on the local disk.
func (l *LocalStorageDriver) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("key cannot be empty")
	}

	key, err := sanitizeKey(key)
	if err != nil {
		return false, fmt.Errorf("invalid key: %w", err)
	}

	fullPath := filepath.Join(l.basePath, key)
	if err := verifyPathWithinBase(l.basePath, fullPath); err != nil {
		return false, fmt.Errorf("path verification failed: %w", err)
	}

	_, err = os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("failed to check file existence: %w", err)
}

// GetName returns the name of the local storage driver.
func (l *LocalStorageDriver) GetName() string {
	return LocalStorageName
}

// GenerateSignedURL returns a direct URL since local storage doesn't support signatures.
func (l *LocalStorageDriver) GenerateSignedURL(ctx context.Context, key string, operation string, expires time.Duration) (string, error) {
	if key == "" {
		return "", fmt.Errorf("key cannot be empty")
	}

	key, err := sanitizeKey(key)
	if err != nil {
		return "", fmt.Errorf("invalid key: %w", err)
	}

	switch operation {
	case "GET":
		return l.baseURL + url.PathEscape(key), nil
	default:
		return "", fmt.Errorf("operation '%s' not supported for local storage", operation)
	}
}

// sanitizeKey prevents directory traversal attacks and cleans paths
func sanitizeKey(key string) (string, error) {
	for _, r := range key {
		if r == 0 {
			return "", fmt.Errorf("null byte in path")
		}
		if unicode.IsControl(r) {
			return "", fmt.Errorf("control character in path")
		}
	}

	key = filepath.FromSlash(key)
	key = filepath.Clean(key)

	if strings.HasPrefix(key, "..") || strings.Contains(key, "/..") ||
		strings.Contains(key, `\..`) || strings.Contains(key, ":") {
		return "", fmt.Errorf("directory traversal attempt detected")
	}

	if runtime.GOOS == "windows" {
		if isWindowsReservedName(key) {
			return "", fmt.Errorf("windows reserved name detected")
		}
	}

	if strings.Contains(key, "\u2028") || strings.Contains(key, "\u2029") {
		return "", fmt.Errorf("unicode separator detected")
	}

	return key, nil
}

// isWindowsReservedName checks for Windows reserved names and stream attempts
func isWindowsReservedName(key string) bool {
	base := filepath.Base(key)
	upper := strings.ToUpper(base)

	if len(base) >= 4 {
		prefix := upper[:4]
		if (prefix == "COM" || prefix == "LPT") && len(base) >= 5 {
			if _, err := strconv.Atoi(base[4:5]); err == nil {
				return true
			}
		}
	}

	return strings.Contains(key, ":")
}

// verifyPathWithinBase ensures the target path is within the base directory
func verifyPathWithinBase(base, target string) error {
	base = filepath.Clean(base)
	target = filepath.Clean(target)

	if runtime.GOOS == "windows" {
		base = strings.ToLower(base)
		target = strings.ToLower(target)
	}

	rel, err := filepath.Rel(base, target)
	if err != nil {
		return fmt.Errorf("failed to determine path relation")
	}

	if strings.HasPrefix(rel, "..") {
		return fmt.Errorf("path escapes base directory")
	}

	return nil
}

// openFileSecurely opens a file with secure flags and permissions
func openFileSecurely(path string) (*os.File, error) {
	flags := os.O_RDWR | os.O_CREATE

	if runtime.GOOS != "windows" {
		if fi, err := os.Lstat(path); err == nil && fi.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("symlinks not allowed")
		}

		const O_NOFOLLOW = 0x200
		flags |= O_NOFOLLOW
	}

	file, err := os.OpenFile(path, flags, 0600)
	if err != nil {
		if runtime.GOOS != "windows" && strings.Contains(err.Error(), "invalid argument") {
			file, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
			if err == nil {
				if fi, err := os.Lstat(path); err == nil && fi.Mode()&os.ModeSymlink != 0 {
					utils.CheckError(file.Close())
					_ = os.Remove(path)
					return nil, fmt.Errorf("symlinks not allowed")
				}
			}
		}
		return nil, err
	}

	return file, nil
}
