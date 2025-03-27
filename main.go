package main

import (
	"flag"
	"fmt"
	"github.com/barasher/go-exiftool"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
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

	oneWeekAgo := time.Now().AddDate(0, 0, -7)

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
		if len(metadata) > 0 {
			foundExif := false
			for k, v := range metadata[0].Fields {
				if strings.EqualFold(k, "DateTimeOriginal") || strings.EqualFold(k, "CreateDate") || strings.EqualFold(k, "DateCreated") {
					//We have exif data, short circuit the loop
					logger.Printf("Found %s of %s for %s. Short circuiting\n", k, v, file)
					foundExif = true
					break
				}
			}
			if foundExif {
				continue
			} else {
				var newFileModifyDate interface{}
				for k, v := range metadata[0].Fields {
					if strings.EqualFold(k, "FileModifyDate") {
						fileModifyDate, err := time.Parse("2006:01:02 15:04:05-07:00", v.(string))
						if err != nil {
							logger.Printf("Error parsing FileModifyDate for %s: %v\n", file, err)
							break
						}
						if !fileModifyDate.Before(oneWeekAgo) {
							logger.Printf("File %s was modified within the last week\n", file)
						} else {
							newFileModifyDate = v
							break
						}
					}
				}
				if newFileModifyDate != nil {
					metadata[0].Fields["DateTimeOriginal"] = newFileModifyDate
					et.WriteMetadata(metadata)
					logger.Printf("Updated DateTimeOriginal for %s to %s\n", file, newFileModifyDate)
				}
			}
		} else {
			logger.Printf("No metadata found for %s\n", file)
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
