package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/ayumi-otosaka-314/brec-pp/cmd"
)

func main() {
	ctx := context.Background()

	credentialPath := flag.String("credential", "", "path to the credential json file")
	nameContains := flag.String("nameContains", "", "keyword in name for search")

	flag.Parse()

	if *credentialPath == "" {
		log.Println("credential path is required")
		flag.Usage()
		os.Exit(1)
	}
	if *nameContains == "" {
		log.Println("keyword in name for search is required")
		flag.Usage()
		os.Exit(1)
	}

	client := cmd.ServiceAccount(*credentialPath)
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("unable to retrieve Drive client: %+v", err)
	}

	r, err := srv.Files.
		List().
		Q(fmt.Sprintf("name contains '%s'", *nameContains)).
		PageSize(10).
		Fields("files(id, name, sha256Checksum)").
		Do()
	if err != nil {
		log.Fatalf("unable to retrieve files: %v", err)
	}
	if len(r.Files) == 0 {
		log.Fatalln("no files found")
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(4)

	for _, f := range r.Files {
		file := f
		g.Go(func() error {
			return doDownload(gCtx, srv, file)
		})
	}

	if err = g.Wait(); err != nil {
		log.Fatalf("error downloading file: %+v", err)
	}
}

func doDownload(ctx context.Context, srv *drive.Service, driveFile *drive.File) error {
	resp, err := srv.Files.Get(driveFile.Id).Context(ctx).Download()
	if err != nil {
		return errors.Wrap(err, "error downloading file")
	}
	defer resp.Body.Close()

	file, err := os.Create(driveFile.Name)
	if err != nil {
		return errors.Wrap(err, "error create file for download")
	}
	defer file.Close()

	log.Printf("start downloading: %s\n", driveFile.Name)
	bytesRead, err := file.ReadFrom(resp.Body)
	if err != nil {
		return errors.Wrap(err, "error writing content")
	}
	log.Printf("Download %s completed; %d bytes transferred", driveFile.Name, bytesRead)
	resp.Body.Close()

	if driveChecksum := driveFile.Sha256Checksum; driveChecksum != "" {
		hasher := sha256.New()

		_, _ = file.Seek(0, io.SeekStart)
		if _, err = io.Copy(hasher, file); err != nil {
			return errors.Wrap(err, "unable to check sha256 of file written")
		}
		localChecksum := hex.EncodeToString(hasher.Sum(nil))
		if driveChecksum != localChecksum {
			return errors.Errorf(
				"sha256 mismatch for file [%s]; gdrive [%s]; local [%s]",
				driveFile.Name, driveChecksum, localChecksum)
		}
	}

	return nil
}
