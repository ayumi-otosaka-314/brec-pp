package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/jwt"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func ServiceAccount(credentialFile string) *http.Client {
	b, err := os.ReadFile(credentialFile)
	if err != nil {
		log.Fatal(err)
	}
	var c = struct {
		PrivateKeyID string `json:"private_key_id"`
		PrivateKey   string `json:"private_key"`
		ClientEmail  string `json:"client_email"`
		TokenURI     string `json:"token_uri"`
	}{}
	if err = json.Unmarshal(b, &c); err != nil {
		log.Fatalf("unable to unmarshal credential json: %+v", err)
	}
	config := &jwt.Config{
		PrivateKeyID: c.PrivateKeyID,
		PrivateKey:   []byte(c.PrivateKey),
		Email:        c.ClientEmail,
		Scopes:       []string{drive.DriveScope},
		TokenURL:     c.TokenURI,
	}
	client := config.Client(context.Background())
	return client
}

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

	client := ServiceAccount(*credentialPath)
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %+v", err)
	}

	r, err := srv.Files.
		List().
		Q(fmt.Sprintf("name contains '%s'", *nameContains)).
		PageSize(10).
		Fields("files(id, name, sha256Checksum)").
		Do()
	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
	}
	if len(r.Files) == 0 {
		log.Fatalln("No files found.")
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

	log.Println("start downloading")
	bytesRead, err := file.ReadFrom(resp.Body)
	if err != nil {
		return errors.Wrap(err, "error writing content")
	}
	log.Printf("Download completed; %d bytes transferred", bytesRead)
	resp.Body.Close()

	if driveChecksum := driveFile.Sha256Checksum; driveChecksum != "" {
		hasher := sha256.New()

		_, _ = file.Seek(0, io.SeekStart)
		if _, err = io.Copy(hasher, file); err != nil {
			return errors.Wrap(err, "unable to check sha256 of file written")
		}
		localChecksum := hex.EncodeToString(hasher.Sum(nil))
		if driveChecksum != localChecksum {
			return errors.Errorf("sha256 mismatch; gdrive [%s]; local [%s]", driveChecksum, localChecksum)
		}
	}

	return nil
}
