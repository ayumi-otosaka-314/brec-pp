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

	"golang.org/x/oauth2/jwt"
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
	fileName := flag.String("fileName", "", "name of the file to download")

	flag.Parse()

	if *credentialPath == "" {
		log.Println("credential path is required")
		flag.Usage()
		os.Exit(1)
	}
	if *fileName == "" {
		log.Println("fileName is required")
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
		Q(fmt.Sprintf("name = '%s'", *fileName)).
		PageSize(10).
		Fields("files(id, name, sha256Checksum)").
		Do()
	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
	}
	if len(r.Files) == 0 {
		log.Fatalln("No files found.")
	} else if len(r.Files) > 1 {
		for _, i := range r.Files {
			log.Printf("%s (%s)", i.Name, i.Id)
		}
		os.Exit(1)
	}
	fileID := r.Files[0].Id

	resp, err := srv.Files.Get(fileID).Download()
	if err != nil {
		log.Fatalf("error downloading file: %+v\n", err)
	}
	defer resp.Body.Close()

	file, err := os.Create(*fileName)
	if err != nil {
		log.Fatalf("error create file for download: %+v\n", err)
	}
	defer file.Close()

	log.Println("start downloading")
	bytesRead, err := file.ReadFrom(resp.Body)
	if err != nil {
		log.Fatalf("error writing content: %+v\n", err)
	}
	log.Printf("Download completed; %d bytes transferred", bytesRead)
	resp.Body.Close()

	if r.Files[0].Sha256Checksum != "" {
		hasher := sha256.New()

		_, _ = file.Seek(0, io.SeekStart)
		if _, err = io.Copy(hasher, file); err != nil {
			log.Fatalf("unable to check sha256 of file written: %+v", err)
		}
		gdriveChecksum := r.Files[0].Sha256Checksum
		localChecksum := hex.EncodeToString(hasher.Sum(nil))
		if gdriveChecksum != localChecksum {
			log.Fatalf("sha256 mismatch; gdrive [%s]; local [%s]", gdriveChecksum, localChecksum)
		}
	}
}
