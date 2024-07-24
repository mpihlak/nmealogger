package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func main() {
	logDirectory := flag.String("logDir", "data", "Directory where downloaded log files are stored")
	credentialsFile := flag.String("credentials", "nmealogger-5cf95ba688f5.json", "Location of Google Drive client credentials")
	parentFolderID := flag.String("folderId", "1Jes5cUmB_MMk4U2qkC7SiJCeT_jBFV0Y", "ID of the data folder in Google Drive")
	deleteFiles := flag.Bool("delete", false, "Delete files from Drive after successful download")
	download := flag.Bool("download", true, "Download files from Drive")
	flag.Parse()

	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithCredentialsFile(*credentialsFile), option.WithScopes(drive.DriveScope))
	if err != nil {
		log.Fatalf("Warning: Unable to create drive Client %v", err)
	}

	// List files at the folder
	log.Printf("Listing files at Drive folder %s", *parentFolderID)
	var files []*drive.File
	numFiles := 0
	pageToken := ""
	for {
		q := srv.Files.List().Q(fmt.Sprintf("'%s' in parents", *parentFolderID))
		if pageToken != "" {
			q = q.PageToken(pageToken)
		}
		r, err := q.Do()
		if err != nil {
			log.Fatalf("Error listing files in Drive: %v\n", err)
		}

		files = append(files, r.Files...)
		numFiles += len(r.Files)
		pageToken = r.NextPageToken
		if pageToken == "" {
			break
		}
	}

	log.Printf("Processing files to %s", *logDirectory)
	for _, file := range files {
		if *download {
			log.Printf("Downloading: %s\n", file.Name)
			resp, err := srv.Files.Get(file.Id).Download()
			if err != nil {
				log.Fatalf("Error downloading file %s %s: %v", file.Id, file.Name, err)
			}

			fileName := filepath.Join(*logDirectory, file.Name)
			log.Printf("Writing to %s", fileName)
			outFile, err := os.Create(fileName)
			if err != nil {
				log.Fatalf("Error creating output file %s: %v", fileName, err)
			}

			io.Copy(outFile, resp.Body)
			outFile.Close()
			resp.Body.Close()
		}

		if *deleteFiles {
			log.Printf("Deleting: %s\n", file.Name)
			if err := srv.Files.Delete(file.Id).Do(); err != nil {
				log.Fatalf("Error deleting file %s %s: %v", file.Id, file.Name, err)
			}
		}
	}

	log.Printf("Done, %d files processed.", numFiles)
}
