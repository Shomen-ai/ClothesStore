package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

// UploadService stores and removes uploaded product images on the local
// filesystem under a single configured directory.
type UploadService struct{ dir string }

// NewUploadService creates an upload service rooted at the given directory.
func NewUploadService(dir string) *UploadService { return &UploadService{dir: dir} }

// Save streams an uploaded file to disk under a collision-resistant,
// nanosecond-timestamped name (keeping the original extension) and returns the
// public "/uploads/..." URL path used by the frontend.
func (s *UploadService) Save(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Derive a unique on-disk name from the current time in nanoseconds so two
	// uploads can't clobber each other; preserve the source file's extension.
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dstPath := filepath.Join(s.dir, filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}
	return "/uploads/" + filename, nil
}

// Delete removes the file backing a stored image URL. Only the base name is
// used, so a stored "/uploads/x.jpg" path can never escape the uploads dir.
func (s *UploadService) Delete(imagePath string) error {
	filename := filepath.Base(imagePath)
	return os.Remove(filepath.Join(s.dir, filename))
}
