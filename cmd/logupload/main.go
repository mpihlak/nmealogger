package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const (
	FileAgeCutOffMinutes = 10
)

func main() {
	logDirectory := flag.String("logDir", "data", "Directory where log files are stored")
	credentialsFile := flag.String("credentials", "nmealogger-5cf95ba688f5.json", "Location of Google Drive client credentials")
	parentFolderID := flag.String("folderId", "1Jes5cUmB_MMk4U2qkC7SiJCeT_jBFV0Y", "ID of the upload folder in Google Drive")
	dontRenameFiles := flag.Bool("dontRenameFiles", false, "Don't rename the uploaded log files to .uploaded")
	flag.Parse()

	entries, err := os.ReadDir(*logDirectory)
	if err != nil {
		log.Fatalf("Error reading log directory: %v", err)
	}

	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithCredentialsFile(*credentialsFile), option.WithScopes(drive.DriveScope))
	if err != nil {
		log.Fatalf("Warning: Unable to create drive Client %v", err)
	}

	log.Printf("Looking at files in %s", *logDirectory)
	filesUploaded := 0
	uploadErrors := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !(strings.HasPrefix(e.Name(), "nmea") && strings.HasSuffix(e.Name(), ".log")) {
			continue
		}

		fileInfo, err := e.Info()
		if err != nil {
			log.Print("Error getting file info for %s: %v", e.Name(), err)
			continue
		}

		if fileInfo.ModTime().After(time.Now().Add(-FileAgeCutOffMinutes * time.Minute)) {
			log.Printf("File is newer than %d minutes, skipping: %s", FileAgeCutOffMinutes, e.Name())
			continue
		}

		pathName := filepath.Join(*logDirectory, e.Name())
		if err := uploadFile(srv, *parentFolderID, pathName); err != nil {
			log.Printf("Error uploading file to Drive: %v", err)
			uploadErrors++
		} else {
			filesUploaded++
			if !*dontRenameFiles {
				if err := os.Rename(pathName, pathName+".uploaded"); err != nil {
					log.Printf("Error renaming file %s: %v", pathName, err)
				}
			}
		}
	}
	log.Printf("Done, %d files uploaded, %d errors.", filesUploaded, uploadErrors)
}

func uploadFile(srv *drive.Service, parentFolder string, fileName string) error {
	log.Printf("Uploading %s", fileName)

	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("error opening file for reading: %w", err)
	}
	defer file.Close()

	baseName := filepath.Base(fileName)

	driveFile := &drive.File{
		Name:    baseName,
		Parents: []string{parentFolder},
	}

	_, err = srv.Files.
		Create(driveFile).
		Media(file).
		ProgressUpdater(func(now, size int64) { fmt.Printf("%d, %d\r", now, size) }).
		Do()

	return err
}
