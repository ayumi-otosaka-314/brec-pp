package cmd

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/drive/v3"
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
