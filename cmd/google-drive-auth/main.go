// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const (
	callbackAddress = "127.0.0.1:53682"
	redirectURL     = "http://127.0.0.1:53682/oauth2callback"
	projectFolder   = "Product Catalog Project"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	clientID := os.Getenv("GOOGLE_DRIVE_OAUTH_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_DRIVE_OAUTH_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return errors.New("set GOOGLE_DRIVE_OAUTH_CLIENT_ID and GOOGLE_DRIVE_OAUTH_CLIENT_SECRET before running this helper")
	}

	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return fmt.Errorf("create OAuth state: %w", err)
	}
	state := hex.EncodeToString(stateBytes)
	codeResult := make(chan authorizationResult, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2callback", func(w http.ResponseWriter, r *http.Request) {
		if subtle.ConstantTimeCompare([]byte(r.URL.Query().Get("state")), []byte(state)) != 1 {
			http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
			return
		}
		if oauthError := r.URL.Query().Get("error"); oauthError != "" {
			codeResult <- authorizationResult{err: errors.New("Google authorization was denied")}
			fmt.Fprintln(w, "Authorization was not completed. You may close this window.")
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			codeResult <- authorizationResult{err: errors.New("Google did not return an authorization code")}
			http.Error(w, "Missing authorization code", http.StatusBadRequest)
			return
		}
		codeResult <- authorizationResult{code: code}
		fmt.Fprintln(w, "Authorization complete. You may close this window.")
	})
	listener, err := net.Listen("tcp", callbackAddress)
	if err != nil {
		return fmt.Errorf("start local OAuth callback: %w", err)
	}
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() { _ = server.Serve(listener) }()
	defer server.Close()

	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     google.Endpoint,
		Scopes:       []string{drive.DriveFileScope},
	}
	fmt.Println("Open this URL in the Gmail account that owns the product images:")
	fmt.Println(oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent")))

	var result authorizationResult
	select {
	case result = <-codeResult:
	case <-time.After(5 * time.Minute):
		return errors.New("timed out waiting for Google authorization")
	}
	if result.err != nil {
		return result.err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	token, err := oauthConfig.Exchange(ctx, result.code)
	if err != nil {
		return errors.New("Google OAuth code exchange failed; verify the OAuth client and consent-screen configuration")
	}
	if token.RefreshToken == "" {
		return errors.New("Google did not return a refresh token; revoke this app's access and rerun the helper")
	}

	service, err := drive.NewService(ctx, option.WithHTTPClient(oauthConfig.Client(ctx, token)))
	if err != nil {
		return errors.New("could not initialize Google Drive")
	}
	rootFolderID, err := findOrCreateProjectFolder(ctx, service)
	if err != nil {
		return errors.New("could not create or find the project folder in Google Drive")
	}

	fmt.Println()
	fmt.Println("Store these values in a secret manager or your untracked .env file:")
	fmt.Printf("GOOGLE_DRIVE_REFRESH_TOKEN=%s\n", token.RefreshToken)
	fmt.Printf("GOOGLE_DRIVE_ROOT_FOLDER_ID=%s\n", rootFolderID)
	return nil
}

type authorizationResult struct {
	code string
	err  error
}

func findOrCreateProjectFolder(ctx context.Context, service *drive.Service) (string, error) {
	files, err := service.Files.List().
		Q("'root' in parents and name = 'Product Catalog Project' and mimeType = 'application/vnd.google-apps.folder' and trashed = false").
		Spaces("drive").
		PageSize(100).
		Fields("files(id)").
		Context(ctx).
		Do()
	if err != nil {
		return "", err
	}
	if len(files.Files) > 0 {
		return files.Files[0].Id, nil
	}
	folder, err := service.Files.Create(&drive.File{
		Name:     projectFolder,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{"root"},
	}).Fields("id").Context(ctx).Do()
	if err != nil {
		return "", err
	}
	return folder.Id, nil
}
