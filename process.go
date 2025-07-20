package main

import (
	"os"
	"path/filepath"
	"strings"
)

// FileRenamer is a struct that holds the file processing logic
type RenamerProcessor struct {
	FolderPath     string        // FolderPath
	Files          []os.FileInfo // All files in the folder
	FilteredFiles  []os.FileInfo // Files after filtering
	FilterExt      string        // Extension filter
	NewNames       []string      // New names for files
	PrefixValue    string        // Prefix to add or remove
	SuffixValue    string        // Suffix to add or remove
	ExtensionValue string        // Extension to change
	PrefixMode     string        // "None", "Add", "Remove"
	SuffixMode     string        // "None", "Add", "Remove"
	ExtensionMode  string        // "None", "Change"

}

// Create new RenamerProcessor instance
func NewRenamerProcessor() *RenamerProcessor {
	return &RenamerProcessor{}
}

// Filter the files based on the specified extension
func (rp *RenamerProcessor) FilterFiles() {
	rp.FilteredFiles = make([]os.FileInfo, 0)
	// If no filter is set, copy all files to filtered files
	if rp.FilterExt == "" {
		rp.FilteredFiles = rp.Files
		return
	}
	// Split the filter extensions by semicolon and process each
	exts := strings.Split(rp.FilterExt, ";")
	for _, file := range rp.Files {
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
				rp.FilteredFiles = append(rp.FilteredFiles, file)
				break
			}
		}
	}
}

// Load files rpom the specified path into the RenamerProcessor
func (rp *RenamerProcessor) LoadFiles(path string) error {
	rp.FolderPath = path
	rp.Files = nil
	rp.FilteredFiles = nil

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
		rp.Files = append(rp.Files, info)
	}
	// Filter the files after loading
	rp.FilterFiles()
	return nil
}

// Generate new names for the filtered files based on the specified prefix, suffix, and extension
func (rp *RenamerProcessor) GenerateNewNames() {
	// Initialize the NewNames slice with the same length as FilteredFiles
	rp.NewNames = make([]string, len(rp.FilteredFiles))
	for i, file := range rp.FilteredFiles {
		oldName := file.Name()
		newName := oldName // Edit the name based on the old name
		// Edit prefix according to the specified mode
		switch rp.PrefixMode {
		case "Add":
			newName = rp.PrefixValue + newName
		case "Remove":
			newName = strings.TrimPrefix(newName, rp.PrefixValue)
		}
		// Edit suffix according to the specified mode
		switch rp.SuffixMode {
		case "Add":
			ext := filepath.Ext(newName)
			base := strings.TrimSuffix(newName, ext)
			newName = base + rp.SuffixValue + ext
		case "Remove":
			ext := filepath.Ext(newName)
			base := strings.TrimSuffix(newName, ext)
			if strings.HasSuffix(base, rp.SuffixValue) {
				base = strings.TrimSuffix(base, rp.SuffixValue)
				newName = base + ext
			}
		}
		switch rp.ExtensionMode {
		case "Change":
			if rp.ExtensionValue != "" {
				ext := filepath.Ext(newName)
				newExt := rp.ExtensionValue
				// Ensure the new extension starts with a dot
				if !strings.HasPrefix(newExt, ".") {
					newExt = "." + newExt
				}
				// If the old name has an extension, replace it with the new extension
				// If the old name has no extension, just append the new extension
				if ext != "" {
					newName = strings.TrimSuffix(newName, ext) + newExt
				} else {
					newName += newExt
				}

			}
		}
		// Store the new name in the NewNames slice
		rp.NewNames[i] = newName
	}
}

func (rp *RenamerProcessor) RenameFiles() (int, error) {
	successCount := 0 // Ensure that NewNames is generated before renaming
	for i, file := range rp.FilteredFiles {
		oldPath := filepath.Join(rp.FolderPath, file.Name())    // Combine folder path and old file name
		newPath := filepath.Join(rp.FolderPath, rp.NewNames[i]) // Combine folder path and new file name
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
	if err := rp.LoadFiles(rp.FolderPath); err != nil {
		return successCount, err
	}
	// Return the count of successfully renamed files
	return successCount, nil
}
