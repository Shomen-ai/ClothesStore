package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

type UploadService struct{ dir string }

func NewUploadService(dir string) *UploadService { return &UploadService{dir: dir} }

func (s *UploadService) Save(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

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

func (s *UploadService) Delete(imagePath string) error {
	filename := filepath.Base(imagePath)
	return os.Remove(filepath.Join(s.dir, filename))
}
