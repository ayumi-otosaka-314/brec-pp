package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/ayumi-otosaka-314/brec-pp/cmd"
)

func main() {
	ctx := context.Background()

	credentialPath := flag.String("credential", "", "path to the credential json file")
	uploadFilePath := flag.String("upload", "", "path of the file to be uploaded")
	parentFolder := flag.String("parentFolder", "", "ID of the parent folder")

	flag.Parse()

	if *credentialPath == "" {
		log.Println("credential path is required")
		flag.Usage()
		os.Exit(1)
	}
	if *uploadFilePath == "" {
		log.Println("upload file path is required")
		flag.Usage()
		os.Exit(1)
	}
	if *parentFolder == "" {
		log.Println("parent folder ID required")
		flag.Usage()
		os.Exit(1)
	}

	client := cmd.ServiceAccount(*credentialPath)
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("unable to retrieve Drive client: %+v", err)
	}

	upload, err := os.Open(*uploadFilePath)
	if err != nil {
		log.Fatalf("unable to open file on file system: %+v", err)
	}
	defer upload.Close()

	res, err := srv.Files.
		Create(&drive.File{
			Name:    path.Base(*uploadFilePath),
			Parents: []string{*parentFolder},
		}).
		Media(upload).
		ProgressUpdater(func(now, size int64) { log.Printf("now: %d, size: %d\n", now, size) }).
		Do()
	if err != nil {
		log.Fatalf("unable to upload: %+v", err)
	}
	log.Printf("upload success: +%+v\n", res)
}
