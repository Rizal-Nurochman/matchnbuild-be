package utils

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/option"
	"github.com/imagekit-developer/imagekit-go/v2/packages/param"
)

type namedReader struct {
	*bytes.Reader
	filename    string
	contentType string
}

func (n *namedReader) Filename() string    { return n.filename }
func (n *namedReader) ContentType() string { return n.contentType }

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

	buf, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = mime.TypeByExtension(ext)
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	reader := &namedReader{
		Reader:      bytes.NewReader(buf),
		filename:    header.Filename,
		contentType: contentType,
	}

	privateKey := os.Getenv("IMAGEKIT_PRIVATE_KEY")
	if privateKey == "" {
		return "", errors.New("IMAGEKIT_PRIVATE_KEY is not set")
	}

	client := imagekit.NewClient(
		option.WithPrivateKey(privateKey),
	)

	resp, err := client.Files.Upload(context.Background(), imagekit.FileUploadParams{
		File:              reader,
		FileName:          header.Filename,
		Folder:            param.NewOpt(folder),
		UseUniqueFileName: param.NewOpt(true),
	})
	if err != nil {
		log.Printf("[imagekit] upload error file=%q folder=%q: %v", header.Filename, folder, err)
		return "", err
	}

	if resp.URL == "" {
		log.Printf("[imagekit] upload succeeded but URL is empty, raw=%s", resp.RawJSON())
		return "", errors.New("imagekit returned empty URL")
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
