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
	RemovePrefix  bool          // Remove prefix
	RemoveSuffix  bool          // Remove suffix
}

// Create new RenamerProcessor instance
func NewRenamerProcessor() *RenamerProcessor {
	return &RenamerProcessor{}
}

// Filter the files based on the specified extension
func (fr *RenamerProcessor) FilterFiles() {
	fr.FilteredFiles = make([]os.FileInfo, 0)

	if fr.FilterExt == "" {
		fr.FilteredFiles = fr.Files
		return
	}
	// Split the filter extensions by semicolon and process each
	exts := strings.Split(fr.FilterExt, ";")
	for _, file := range fr.Files {
		fileExt := strings.ToLower(filepath.Ext(file.Name()))
		for _, ext := range exts {
			ext = strings.ToLower(strings.TrimSpace(ext))
			if ext != "" && !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
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

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		info, err := file.Info()
		if err != nil {
			continue
		}
		fr.Files = append(fr.Files, info)
	}

	fr.FilterFiles()
	return nil
}
