package main

import (
	"os"
	"path/filepath"
	"strings"
)

// FileRenamer is a struct that holds the file processing logic
type RenamerProcessor struct {
	FolderPath    string        // FolderPath
	Files         []os.FileInfo // All files in the folder
	FilteredFiles []os.FileInfo // Files after filtering
	FilterExt     string        // Extension filter
	NewNames      []string      // New names for files
	Prefix        string        // Prefix
	Suffix        string        // Suffix
	NewExtension  string        // New extension
	RemovePrefix  bool          // Remove prefix or not
	RemoveSuffix  bool          // Remove suffix or not
}

// Create new RenamerProcessor instance
func NewRenamerProcessor() *RenamerProcessor {
	return &RenamerProcessor{}
}

// Filter the files based on the specified extension
func (fr *RenamerProcessor) FilterFiles() {
	fr.FilteredFiles = make([]os.FileInfo, 0)
	// If no filter is set, copy all files to filtered files
	if fr.FilterExt == "" {
		fr.FilteredFiles = fr.Files
		return
	}
	// Split the filter extensions by semicolon and process each
	exts := strings.Split(fr.FilterExt, ";")
	for _, file := range fr.Files {
		// Check if the file matches any of the specified extensions
		fileExt := strings.ToLower(filepath.Ext(file.Name()))
		for _, ext := range exts {
			ext = strings.ToLower(strings.TrimSpace(ext))
			// Ensure the extension starts with a dot
			if ext != "" && !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			// If the file's extension matches the filter, add it to the filtered files
			if ext == fileExt {
				fr.FilteredFiles = append(fr.FilteredFiles, file)
				break
			}
		}
	}
}

// Load files from the specified path into the RenamerProcessor
func (fr *RenamerProcessor) LoadFiles(path string) error {
	fr.FolderPath = path
	fr.Files = nil
	fr.FilteredFiles = nil

	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	// Initialize the Files slice
	for _, file := range files {
		// Check if the file is a directory, if so, skip it
		if file.IsDir() {
			continue
		}
		// Get file info and append to the Files slice if no error
		info, err := file.Info()
		if err != nil {
			continue
		}
		fr.Files = append(fr.Files, info)
	}
	// Filter the files after loading
	fr.FilterFiles()
	return nil
}

// Generate new names for the filtered files based on the specified prefix, suffix, and extension
func (fr *RenamerProcessor) GenerateNewNames() {
	// Initialize the NewNames slice with the same length as FilteredFiles
	fr.NewNames = make([]string, len(fr.FilteredFiles))
	for i, file := range fr.FilteredFiles {
		oldName := file.Name()
		newName := oldName // Edit the name based on the old name
		// Edit prefix if the prefix is set
		if fr.Prefix != "" {
			// Delete the prefix if user chooses RemovePrefix and the old name has the prefix
			// Add the prefix only when RemovePrefix is false
			if fr.RemovePrefix && strings.HasPrefix(oldName, fr.Prefix) {
				newName = strings.TrimPrefix(newName, fr.Prefix)
			} else if !fr.RemovePrefix {
				newName = fr.Prefix + newName
			}
		}
		// Edit suffix if the suffix is set
		if fr.Suffix != "" {
			ext := filepath.Ext(newName)             // Get the file extension
			base := strings.TrimSuffix(newName, ext) // Split the name and extension
			// Delete the suffix if user chooses Removesuffix and the old name has the suffix
			// Add the suffix only when Removesuffix is false
			if fr.RemoveSuffix && strings.HasSuffix(base, fr.Suffix) {
				base = strings.TrimSuffix(base, fr.Suffix) // Delete the the suffix and get base name
				newName = base + ext                       //Reconstruct the base name with the extension
			} else if !fr.RemoveSuffix {
				newName = base + fr.Suffix + ext
			}
		}
		// Edit extension if the new extension is set
		if fr.NewExtension != "" {
			ext := filepath.Ext(newName) // Get old file extension
			newExt := fr.NewExtension    // Load new extension
			// Ensure the new extension starts with a dot
			if !strings.HasPrefix(newExt, ".") {
				newExt = "." + newExt
			}
			// If the old name has an extension, replace it with the new extension
			// If the old name has no exrtension, just append the new extension
			if ext != "" {
				newName = strings.TrimSuffix(newName, ext) + newExt
			} else {
				newName += newExt
			}
		}
		// Store the new name in the NewNames slice
		fr.NewNames[i] = newName
	}
}

func (fr *RenamerProcessor) RenameFiles() (int, error) {
	successCount := 0 // Ensure that NewNames is generated before renaming
	for i, file := range fr.FilteredFiles {
		oldPath := filepath.Join(fr.FolderPath, file.Name())    // Combine folder path and old file name
		newPath := filepath.Join(fr.FolderPath, fr.NewNames[i]) // Combine folder path and new file name
		// Skip renaming if the old and new paths are the same
		if oldPath == newPath {
			continue
		}
		// Rename the file and check for errors
		err := os.Rename(oldPath, newPath)
		if err != nil {
			return successCount, err // If error occurs, return the count and error
		}
		successCount++ // If no error occurs, increse the success count
	}
	// Reload the Files and check for any errors
	if err := fr.LoadFiles(fr.FolderPath); err != nil {
		return successCount, err
	}
	// Return the count of successfully renamed files
	return successCount, nil
}
