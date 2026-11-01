package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

const MaxImageSize = 5 * 1024 * 1024 // 5 MB

var (
	ErrFileNotFound      = errors.New("file not found")
	ErrFileTooLarge      = errors.New("file size exceeds 5MB limit")
	ErrInvalidFileExt    = errors.New("unsupported file extension (allowed: jpg, jpeg, png, webp)")
	ErrMissingCloudinary = errors.New("cloudinary credentials are not configured")
)

var AllowedImageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

var AllowedDocumentExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".pdf":  true,
}

func ValidateImageFile(file *multipart.FileHeader) error {
	if file == nil {
		return ErrFileNotFound
	}
	if file.Size > MaxImageSize {
		return ErrFileTooLarge
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !AllowedImageExtensions[ext] {
		return ErrInvalidFileExt
	}
	return nil
}

func ValidateDocumentFile(file *multipart.FileHeader) error {
	if file == nil {
		return ErrFileNotFound
	}
	if file.Size > MaxImageSize {
		return ErrFileTooLarge
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !AllowedDocumentExtensions[ext] {
		return errors.New("format file tidak didukung (format yang diperbolehkan: jpg, jpeg, png, webp, pdf)")
	}
	return nil
}

type CloudinaryUploader struct {
	client      *cloudinary.Cloudinary
	projectName string
}

func NewCloudinaryUploader(cloudName, apiKey, apiSecret, projectName string) (*CloudinaryUploader, error) {
	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, ErrMissingCloudinary
	}

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("init cloudinary client: %w", err)
	}

	if projectName == "" {
		projectName = "proceive"
	}

	return &CloudinaryUploader{
		client:      cld,
		projectName: projectName,
	}, nil
}

func (u *CloudinaryUploader) Upload(ctx context.Context, file multipart.File, filename, moduleName string) (string, error) {
	if moduleName == "" {
		moduleName = "general"
	}

	folderPath := fmt.Sprintf("%s/%s", strings.Trim(u.projectName, "/"), strings.Trim(moduleName, "/"))

	// Bersihkan seluruh ekstensi dari cleanName agar PublicID tidak pernah berakhiran ekstensi
	cleanName := filename
	for {
		e := filepath.Ext(cleanName)
		if e == "" {
			break
		}
		cleanName = strings.TrimSuffix(cleanName, e)
	}
	cleanName = strings.ReplaceAll(cleanName, " ", "_")
	publicID := fmt.Sprintf("%d_%s", time.Now().Unix(), cleanName)

	ext := strings.ToLower(filepath.Ext(filename))
	resourceType := "image"
	// Jika file berupa PDF atau dokumen, simpan sebagai file asli (raw) tanpa dikonversi ke gambar
	if ext == ".pdf" || ext == ".zip" || ext == ".doc" || ext == ".docx" {
		resourceType = "raw"
	}

	uniqueFilename := false
	params := uploader.UploadParams{
		Folder:         folderPath,
		PublicID:       publicID,
		ResourceType:   resourceType,
		UniqueFilename: &uniqueFilename,
	}

	result, err := u.client.Upload.Upload(ctx, file, params)
	if err != nil {
		return "", fmt.Errorf("upload to cloudinary: %w", err)
	}

	finalURL := result.SecureURL
	// Bersihkan jika terjadi ekstensi ganda dari Cloudinary (seperti .pdf.pdf)
	for strings.Contains(finalURL, ".pdf.pdf") {
		finalURL = strings.ReplaceAll(finalURL, ".pdf.pdf", ".pdf")
	}
	for strings.Contains(finalURL, ".jpg.jpg") {
		finalURL = strings.ReplaceAll(finalURL, ".jpg.jpg", ".jpg")
	}
	for strings.Contains(finalURL, ".png.png") {
		finalURL = strings.ReplaceAll(finalURL, ".png.png", ".png")
	}

	return finalURL, nil
}

func UploadToCloudinary(ctx context.Context, file multipart.File, filename, cloudName, apiKey, apiSecret, projectName, moduleName string) (string, error) {
	u, err := NewCloudinaryUploader(cloudName, apiKey, apiSecret, projectName)
	if err != nil {
		return "", err
	}
	return u.Upload(ctx, file, filename, moduleName)
}

func SaveLocalFile(file *multipart.FileHeader, destDir, projectName, moduleName string) (string, error) {
	if projectName == "" {
		projectName = "proceive"
	}
	if moduleName == "" {
		moduleName = "general"
	}

	targetDir := filepath.Join(destDir, projectName, moduleName)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("create storage directory: %w", err)
	}

	cleanName := strings.ReplaceAll(strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename)), " ", "_")
	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), cleanName, filepath.Ext(file.Filename))
	targetPath := filepath.Join(targetDir, filename)

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(targetPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return targetPath, nil
}
