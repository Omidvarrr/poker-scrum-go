package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func ConvertToURL(filePath string, baseURL string) string {
	if filePath == "" {
		return ""
	}

	if strings.HasPrefix(filePath, "http") {
		return filePath
	}

	cleanPath := filepath.Clean(filePath)

	if strings.HasPrefix(cleanPath, "data/uploads/") {
		return baseURL + "/uploads/" + strings.TrimPrefix(cleanPath, "data/uploads/")
	}

	if strings.HasPrefix(cleanPath, "uploads/") {
		return baseURL + "/uploads/" + strings.TrimPrefix(cleanPath, "uploads/")
	}

	return baseURL + "/uploads/" + filepath.Base(cleanPath)
}

func DownloadAndSaveAvatar(avatarURL string) (string, error) {
	if avatarURL == "" {
		return "", nil
	}

	// Skip if it's already a local file
	if !strings.HasPrefix(avatarURL, "http") {
		return avatarURL, nil
	}

	// Create uploads directory if it doesn't exist
	uploadsDir := "data/uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create uploads directory: %w", err)
	}

	// Download the image
	resp, err := http.Get(avatarURL)
	if err != nil {
		return "", fmt.Errorf("failed to download avatar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download avatar: status %d", resp.StatusCode)
	}

	// Get file extension from content type
	ext := ".jpg" // default
	contentType := resp.Header.Get("Content-Type")
	switch contentType {
	case "image/png":
		ext = ".png"
	case "image/jpeg", "image/jpg":
		ext = ".jpg"
	case "image/gif":
		ext = ".gif"
	case "image/webp":
		ext = ".webp"
	}

	// Generate random filename
	randomName := RandomURLSafe(16) + ext
	filePath := filepath.Join(uploadsDir, randomName)

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy the image data
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save avatar: %w", err)
	}

	return filePath, nil
}

func SaveUploadedFile(fileHeader *multipart.FileHeader, category string) (string, error) {
	// Create uploads directory if it doesn't exist
	uploadsDir := filepath.Join("data", "uploads", category)
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create uploads directory: %w", err)
	}

	// Open the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	// Get file extension from the original filename
	ext := filepath.Ext(fileHeader.Filename)
	if ext == "" {
		ext = ".jpg" // default extension
	}

	// Generate random filename
	randomName := RandomURLSafe(16) + ext
	filePath := filepath.Join(uploadsDir, randomName)

	// Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy the uploaded file to destination
	_, err = io.Copy(dst, file)
	if err != nil {
		return "", fmt.Errorf("failed to save uploaded file: %w", err)
	}

	return filePath, nil
}
