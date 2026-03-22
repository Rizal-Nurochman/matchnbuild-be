package utils

import (
	"context"
	"errors"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/option"
	"github.com/imagekit-developer/imagekit-go/v2/packages/param"
)

var (
	ErrFileTooLarge       = errors.New("file size exceeds maximum limit of 5MB")
	ErrFileTypeNotAllowed = errors.New("file type is not allowed")
)

var allowedImageTypes = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
}

var allowedFileTypes = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
	".pdf": true, ".dwg": true, ".dxf": true,
}

const maxFileSize int64 = 5 * 1024 * 1024 // 5MB

// UploadToImageKit uploads a file to ImageKit and returns the public URL.
// Requires IMAGEKIT_PRIVATE_KEY env variable to be set.
// SDK reads IMAGEKIT_PRIVATE_KEY from environment by default.
func UploadToImageKit(file multipart.File, header *multipart.FileHeader, folder string) (string, error) {
	if header.Size > maxFileSize {
		return "", ErrFileTooLarge
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedFileTypes[ext] {
		return "", ErrFileTypeNotAllowed
	}

	privateKey := os.Getenv("IMAGEKIT_PRIVATE_KEY")

	client := imagekit.NewClient(
		option.WithPrivateKey(privateKey),
	)

	resp, err := client.Files.Upload(context.Background(), imagekit.FileUploadParams{
		File:     file,
		FileName: header.Filename,
		Folder:   param.NewOpt(folder),
	})
	if err != nil {
		return "", err
	}

	return resp.URL, nil
}

func ValidateImageFile(header *multipart.FileHeader) error {
	if header.Size > maxFileSize {
		return ErrFileTooLarge
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedImageTypes[ext] {
		return ErrFileTypeNotAllowed
	}

	return nil
}
