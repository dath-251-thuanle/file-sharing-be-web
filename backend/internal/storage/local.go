package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// LocalStorage implements the Storage interface by persisting files on disk.
type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{basePath: basePath}
}

func (s *LocalStorage) Upload(ctx context.Context, obj *Object) (*Location, error) {
	if err := ValidateObject(obj); err != nil {
		return nil, err
	}
	relPath, err := s.safeRelativePath(obj.Name)
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(s.basePath, obj.Container.String(), filepath.Dir(relPath))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("local storage: mkdir failed: %w", err)
	}

	fileName := filepath.Base(relPath)
	if fileName == "." || fileName == "/" {
		fileName = uuid.NewString()
	}

	fullPath := filepath.Join(dir, fileName)
	out, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("local storage: create file failed: %w", err)
	}
	defer out.Close()

	if _, err = io.Copy(out, obj.Reader); err != nil {
		return nil, fmt.Errorf("local storage: write failed: %w", err)
	}

	loc := &Location{
		Container: obj.Container,
		Path:      filepath.ToSlash(filepath.Join(obj.Container.String(), fileName)),
		URL:       fullPath,
	}
	return loc, nil
}

func (s *LocalStorage) Download(ctx context.Context, loc *Location) (*DownloadResult, error) {
	if err := ValidateLocation(loc); err != nil {
		return nil, err
	}
	fullPath := filepath.Join(s.basePath, filepath.FromSlash(loc.Path))
	handle, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("local storage: open failed: %w", err)
	}

	info, err := handle.Stat()
	if err != nil {
		handle.Close()
		return nil, fmt.Errorf("local storage: stat failed: %w", err)
	}

	return &DownloadResult{
		Reader:      handle,
		Size:        info.Size(),
		ContentType: "",
	}, nil
}

func (s *LocalStorage) Delete(ctx context.Context, loc *Location) error {
	if err := ValidateLocation(loc); err != nil {
		return err
	}
	fullPath := filepath.Join(s.basePath, filepath.FromSlash(loc.Path))
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("local storage: delete failed: %w", err)
	}
	return nil
}

func (s *LocalStorage) safeRelativePath(name string) (string, error) {
	clean := filepath.Clean(name)
	if clean == "." || clean == "/" {
		return uuid.NewString(), nil
	}
	if strings.HasPrefix(clean, "..") {
		return "", fmt.Errorf("local storage: invalid object path")
	}
	return clean, nil
}
