package main

import (
	"flag"
	"fmt"
	"github.com/barasher/go-exiftool"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Define and parse the dirPath flag
	dirPath := flag.String("path", "", "Directory path to scan for JPG files")
	flag.Parse()

	if dirPath == nil || *dirPath == "" {
		fmt.Println("Please provide a directory path to scan for JPG files")
		return
	}
	// Open log file
	logFile, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer logFile.Close()

	// Create a new logger
	logger := log.New(logFile, "", log.LstdFlags)

	// Get all files in the directory
	files, err := getAllJPGs(*dirPath)
	if err != nil {
		logger.Printf("Error getting files: %v\n", err)
		return
	}

	for _, file := range files {
		// Create a new ExifTool instance
		et, err := exiftool.NewExiftool()
		if err != nil {
			logger.Printf("Error when intializing: %v\n", err)
			return
		}
		defer et.Close()

		// Extract metadata from the file
		metadata := et.ExtractMetadata(file)
		if metadata[0].Err != nil {
			logger.Printf("Error extracting metadata for %s: %v\n", file, metadata[0].Err)
			continue
		}

		if len(metadata) > 0 {
			foundExif := false
			for k, v := range metadata[0].Fields {
				if strings.EqualFold(k, "DateTimeOriginal") || strings.EqualFold(k, "CreateDate") || strings.EqualFold(k, "DateCreated") {
					//We have date exif data, short circuit the loop
					logger.Printf("Found %s of %s for %s. Short circuiting\n", k, v, file)
					foundExif = true
					break
				}
			}
			if !foundExif {
				var fileModifyDate interface{}
				for k, v := range metadata[0].Fields {
					if strings.EqualFold(k, "FileModifyDate") {
						fileModifyDate = v
						break
					}
				}
				if fileModifyDate != nil {
					metadata[0].Fields["DateTimeOriginal"] = fileModifyDate
					metadata[0].Err = nil

					et.WriteMetadata(metadata)

					if metadata[0].Err != nil {
						logger.Printf("Error writing metadata for %s: %v\n", file, metadata[0].Err)
						continue
					} else {
						logger.Printf("Updated DateTimeOriginal for %s to %s\n", file, fileModifyDate)
					}
				}
			}
		} else {
			logger.Printf("No metadata found for %s\n", file)
		}
		err = et.Close()
		if err != nil {
			logger.Printf("Error closing exiftool: %v\n", err)
			return
		}
	}
	fmt.Printf("Done\n")
}

func getAllJPGs(dirPath string) ([]string, error) {
	var files []string
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".jpg" || ext == ".jpeg" {
				files = append(files, path)
			}
		}
		return nil
	})
	return files, err
}
