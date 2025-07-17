package main

import (
	"log"
	"os"
)

func LoadFont(fontPath string) []byte {
	// Load font file from the specified path
	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		log.Fatalf("Failed to load the font: %v", err)
	}
	return fontData
}
