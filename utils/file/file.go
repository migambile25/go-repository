package file

import (
	"context"        // context provides support for cancellation and timeouts.
	"fmt"            // fmt provides formatting and printing functions.
	"io"             // io provides interfaces for I/O operations.
	"mime/multipart" // multipart provides MIME multipart parsing.
	"os"             // os provides file system operations.
	"path/filepath"  // filepath provides utilities for file path manipulation.
	"regexp"         // regexp provides regular expression utilities.
	"strings"        // strings provides utilities for string manipulation.
	"time"           // time provides functionality for handling time and durations.

	"github.com/google/uuid"             // uuid provides UUID generation.
	"github.com/migambile25/go-repository/utils/helpers" // helpers provides utility functions.
	"github.com/migambile25/go-repository/utils/log"     // log provides colored logging utilities.
)

// UploadResult represents the result of a file upload operation.
type UploadResult struct {
	Filename     string    // Filename is the unique generated filename
	OriginalName string    // OriginalName is the original filename
	Size         int64     // Size is the file size in bytes
	UploadTime   time.Time // UploadTime is when the file was uploaded
	FileType     string    // FileType is the detected file type
}

// toKebabCase converts a string to kebab-case (lowercase with hyphens).
// Returns the converted string.
func toKebabCase(stringValue string) string {
	if stringValue == "" {
		return ""
	}

	// Replace non-alphanumeric characters with hyphens.
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	kebab := re.ReplaceAllString(stringValue, "-")

	// Insert hyphens between lowercase and uppercase letters (e.g., "camelCase" -> "camel-case").
	re2 := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	kebab = re2.ReplaceAllString(kebab, "${1}-${2}")

	// Convert to lowercase and trim leading/trailing hyphens.
	return strings.Trim(strings.ToLower(kebab), "-")
}

// ensureUploadDirectory ensures the upload directory exists with proper permissions.
func ensureUploadDirectory(uploadDirectory string) error {
	if uploadDirectory == "" {
		return helpers.CreateError("upload directory cannot be empty")
	}

	log.Info("📁 Ensuring upload directory exists: " + uploadDirectory)
	if err := os.MkdirAll(uploadDirectory, 0755); err != nil {
		log.Error("❌ Unable to create upload directory: " + err.Error())
		return helpers.WrapError(err, "failed to create upload directory")
	}

	return nil
}

// generateUniqueFilename generates a unique filename with kebab-case and UUID.
func generateUniqueFilename(originalName string, convertToWebP bool) string {
	ext := filepath.Ext(originalName)

	// Update extension to .webp if converting
	if convertToWebP && (ext == ".jpg" || ext == ".jpeg" || ext == ".png") {
		ext = ".webp"
	}

	// Extract base filename without extension and convert to kebab-case
	base := strings.TrimSuffix(filepath.Base(originalName), ext)
	baseKebab := toKebabCase(base)

	// Generate unique filename with timestamp and UUID
	timestamp := time.Now().Format("20060102-150405")
	uniqueFilename := fmt.Sprintf("%s-%s-%s%s", baseKebab, timestamp, uuid.New().String(), ext)

	return uniqueFilename
}

// UploadFile uploads a single file to the specified directory.
// Optionally converts images to WebP format and generates a unique filename.
// Returns the unique filename or an error if the upload fails.
func UploadFile(file io.Reader, fileName, uploadDirectory string, convertToWebP bool) (string, error) {
	// Create context with timeout for upload operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Validate input parameters
	if file == nil {
		return "", helpers.CreateError("file reader cannot be nil")
	}
	if fileName == "" {
		return "", helpers.CreateError("file name cannot be empty")
	}

	// Ensure upload directory exists
	if err := ensureUploadDirectory(uploadDirectory); err != nil {
		return "", err
	}

	// Convert the file to WebP format if requested and supported
	var processedFile io.Reader = file
	if convertToWebP {
		log.Info("🖼️ Converting image to WebP format: " + fileName)
		converted, err := CheckAndConvertFile(file, fileName)
		if err != nil {
			log.Error("❌ Conversion to WebP failed: " + err.Error())
			return "", helpers.WrapError(err, "WebP conversion failed")
		}
		processedFile = converted
	}

	// Generate unique filename
	uniqueFilename := generateUniqueFilename(fileName, convertToWebP)
	destinationPath := filepath.Join(uploadDirectory, uniqueFilename)

	// Create the destination file
	log.Info("📝 Creating file: " + destinationPath)
	destination, err := os.Create(destinationPath)
	if err != nil {
		log.Error("❌ Failed to create file: " + err.Error())
		return "", helpers.WrapError(err, "failed to create destination file")
	}
	defer destination.Close()

	// Copy the file content to the destination with context support
	log.Info("📤 Copying file content to destination")

	// Use a goroutine for copy with context cancellation
	copyDone := make(chan error, 1)
	go func() {
		_, err := io.Copy(destination, processedFile)
		copyDone <- err
	}()

	select {
	case <-ctx.Done():
		// Context was cancelled, clean up the partially written file
		destination.Close()
		os.Remove(destinationPath)
		return "", helpers.WrapError(ctx.Err(), "upload cancelled during copy")
	case err := <-copyDone:
		if err != nil {
			// Clean up on copy error
			destination.Close()
			os.Remove(destinationPath)
			log.Error("❌ Failed to write file content: " + err.Error())
			return "", helpers.WrapError(err, "failed to copy file content to destination")
		}
	}

	// Log successful file upload
	log.Success("✅ File uploaded successfully: " + uniqueFilename)
	return uniqueFilename, nil
}

// UploadMultipartFile handles file upload from HTTP multipart form.
func UploadMultipartFile(fileHeader *multipart.FileHeader, uploadDirectory string, convertToWebP bool) (*UploadResult, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, helpers.WrapError(err, "failed to open multipart file")
	}
	defer file.Close()

	filename, err := UploadFile(file, fileHeader.Filename, uploadDirectory, convertToWebP)
	if err != nil {
		return nil, err
	}

	return &UploadResult{
		Filename:     filename,
		OriginalName: fileHeader.Filename,
		Size:         fileHeader.Size,
		UploadTime:   time.Now(),
		FileType:     helpers.GetFileType(fileHeader.Filename),
	}, nil
}

// DeleteFile removes a single file from the specified directory.
// Returns an error if the file does not exist or deletion fails.
func DeleteFile(filename, uploadDirectory string) error {
	// Create context with timeout for delete operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check context cancellation
	select {
	case <-ctx.Done():
		return helpers.WrapError(ctx.Err(), "delete operation cancelled")
	default:
		// Continue with deletion
	}

	// Validate input parameters
	if filename == "" {
		return helpers.CreateError("filename cannot be empty")
	}

	// Construct the full file path
	filePath := filepath.Join(uploadDirectory, filename)

	// Log the start of the file deletion process
	log.Info("🗑️ Deleting file: " + filePath)

	// Attempt to remove the file
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			log.Warning("⚠️ File not found: " + filename)
			return helpers.WrapError(err, "file not found")
		}
		log.Error("❌ Failed to delete file: " + err.Error())
		return helpers.WrapError(err, "failed to delete file")
	}

	// Log successful file deletion
	log.Success("✅ File deleted: " + filename)
	return nil
}

// UploadMultipleFiles uploads multiple files and rolls back if any fail.
// Returns a list of uploaded filenames or an error if any upload fails.
func UploadMultipleFiles(files []io.Reader, fileNames []string, uploadDirectory string, convertToWebP bool) ([]string, error) {
	// Create context with timeout for batch upload operation
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Validate that the number of files matches the number of filenames
	if len(files) != len(fileNames) {
		errMsg := "❌ Number of files and filenames mismatch"
		log.Error(errMsg)
		return nil, helpers.CreateError(errMsg)
	}

	// Log the start of the batch upload process
	log.Info("📦 Starting batch file upload for " + fmt.Sprintf("%d", len(files)) + " files")

	// Initialize a slice to store uploaded filenames
	uploadedFiles := make([]string, 0, len(files))

	// Upload each file individually
	for i, file := range files {
		// Check context cancellation before each upload
		select {
		case <-ctx.Done():
			// Rollback uploaded files on cancellation
			rollbackUploads(uploadedFiles, uploadDirectory)
			return nil, helpers.WrapError(ctx.Err(), "batch upload cancelled")
		default:
			// Continue with upload
		}

		// Attempt to upload the file
		uniqueFilename, err := UploadFile(file, fileNames[i], uploadDirectory, convertToWebP)
		if err != nil {
			// Log the failure and initiate rollback of previously uploaded files
			log.Error("❌ Upload failed for file: " + fileNames[i] + " — initiating rollback")
			rollbackUploads(uploadedFiles, uploadDirectory)
			return nil, helpers.WrapErrorf(err, "failed to upload file %s", fileNames[i])
		}

		// Add the uploaded filename to the list
		uploadedFiles = append(uploadedFiles, uniqueFilename)
	}

	// Log successful batch upload
	log.Success("✅ All files uploaded successfully")
	return uploadedFiles, nil
}

// rollbackUploads deletes all uploaded files in case of failure.
func rollbackUploads(uploadedFiles []string, uploadDirectory string) {
	if len(uploadedFiles) == 0 {
		return
	}

	log.Warning("🔄 Rolling back uploaded files due to failure")
	for _, filename := range uploadedFiles {
		if err := DeleteFile(filename, uploadDirectory); err != nil {
			log.Error("⚠️ Rollback deletion failed for: " + filename + " — " + err.Error())
		}
	}
}

// DeleteMultipleFiles deletes multiple files and returns an error if any deletions fail.
// Returns nil if all deletions succeed.
func DeleteMultipleFiles(filenames []string, uploadDirectory string) error {
	// Create context with timeout for batch delete operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check context cancellation
	select {
	case <-ctx.Done():
		return helpers.WrapError(ctx.Err(), "batch delete operation cancelled")
	default:
		// Continue with deletion
	}

	// Log the start of the batch deletion process
	log.Info("🗑️ Deleting " + fmt.Sprintf("%d", len(filenames)) + " files")

	// Track any files that fail to delete
	var failedDeletes []string

	// Attempt to delete each file
	for _, filename := range filenames {
		// Check context cancellation before each deletion
		select {
		case <-ctx.Done():
			log.Warning("⚠️ Batch delete operation cancelled")
			if len(failedDeletes) > 0 {
				return helpers.CreateErrorf("partial deletion completed, failed files: %v", failedDeletes)
			}
			return helpers.WrapError(ctx.Err(), "batch delete cancelled")
		default:
			// Continue with deletion
		}

		if err := DeleteFile(filename, uploadDirectory); err != nil {
			log.Error("❌ Failed to delete file: " + filename + " — " + err.Error())
			failedDeletes = append(failedDeletes, filename)
		}
	}

	// Return an error if any deletions failed
	if len(failedDeletes) > 0 {
		errMsg := fmt.Sprintf("⚠️ Could not delete files: %v", failedDeletes)
		log.Error(errMsg)
		return helpers.CreateError(errMsg)
	}

	// Log successful batch deletion
	log.Success("✅ All files deleted successfully")
	return nil
}

// FileExists checks if a file exists in the upload directory.
func FileExists(filename, uploadDirectory string) bool {
	filePath := filepath.Join(uploadDirectory, filename)
	return helpers.FileExists(filePath)
}

// GetFileInfo returns information about a file in the upload directory.
func GetFileInfo(filename, uploadDirectory string) (os.FileInfo, error) {
	filePath := filepath.Join(uploadDirectory, filename)
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, helpers.WrapError(err, "failed to get file info")
	}
	return info, nil
}

// GetFileSize returns the size of a file in the upload directory.
func GetFileSize(filename, uploadDirectory string) (int64, error) {
	info, err := GetFileInfo(filename, uploadDirectory)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// ListFiles lists all files in the upload directory.
func ListFiles(uploadDirectory string) ([]string, error) {
	if err := ensureUploadDirectory(uploadDirectory); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(uploadDirectory)
	if err != nil {
		return nil, helpers.WrapError(err, "failed to read upload directory")
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// CleanupOldFiles deletes files older than the specified duration.
func CleanupOldFiles(uploadDirectory string, olderThan time.Duration) (int, error) {
	files, err := ListFiles(uploadDirectory)
	if err != nil {
		return 0, err
	}

	cutoffTime := time.Now().Add(-olderThan)
	deletedCount := 0

	for _, filename := range files {
		info, err := GetFileInfo(filename, uploadDirectory)
		if err != nil {
			log.Warning("⚠️ Could not get info for file: " + filename + " — " + err.Error())
			continue
		}

		if info.ModTime().Before(cutoffTime) {
			if err := DeleteFile(filename, uploadDirectory); err != nil {
				log.Warning("⚠️ Could not delete old file: " + filename + " — " + err.Error())
			} else {
				deletedCount++
			}
		}
	}

	log.Info(fmt.Sprintf("🧹 Cleanup completed: deleted %d files older than %v", deletedCount, olderThan))
	return deletedCount, nil
}

// GetFileStats returns statistics about files in the upload directory.
func GetFileStats(uploadDirectory string) (fileCount int, totalSize int64, err error) {
	files, err := ListFiles(uploadDirectory)
	if err != nil {
		return 0, 0, err
	}

	for _, filename := range files {
		size, err := GetFileSize(filename, uploadDirectory)
		if err != nil {
			log.Warning("⚠️ Could not get size for file: " + filename + " — " + err.Error())
			continue
		}
		fileCount++
		totalSize += size
	}

	return fileCount, totalSize, nil
}
