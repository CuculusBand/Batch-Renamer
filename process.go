package main

import "os"

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
	RemovePrefix  bool          // Remove prefix
	RemoveSuffix  bool          // Remove suffix
}

// Create new FileRenamer instance
func NewFileProcessor() *RenamerProcessor {
	return &RenamerProcessor{}
}
