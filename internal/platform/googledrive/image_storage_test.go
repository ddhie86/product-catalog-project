// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package googledrive

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type fakeDriveAPI struct {
	folders      map[string]string
	created      []string
	parents      []string
	uploadNames  []string
	uploadedData [][]byte
	publicIDs    []string
	uploadErr    error
	publicErr    error
}

func (a *fakeDriveAPI) findFolder(_ context.Context, parentID, name string) (string, error) {
	return a.folders[parentID+"/"+name], nil
}

func (a *fakeDriveAPI) createFolder(_ context.Context, parentID, name string) (string, error) {
	id := "folder-" + name
	a.folders[parentID+"/"+name] = id
	a.created = append(a.created, name)
	a.parents = append(a.parents, parentID)
	return id, nil
}

func (a *fakeDriveAPI) uploadFile(_ context.Context, parentID, name, _ string, content io.Reader) (string, string, error) {
	if a.uploadErr != nil {
		return "", "", a.uploadErr
	}
	data, err := io.ReadAll(content)
	if err != nil {
		return "", "", err
	}
	a.parents = append(a.parents, parentID)
	a.uploadNames = append(a.uploadNames, name)
	a.uploadedData = append(a.uploadedData, data)
	return "file-id", name, nil
}

func (a *fakeDriveAPI) makePublic(_ context.Context, fileID string) error {
	a.publicIDs = append(a.publicIDs, fileID)
	return a.publicErr
}

func TestUploadCreatesScopedFoldersOnceAndMakesImagePublic(t *testing.T) {
	api := &fakeDriveAPI{folders: make(map[string]string)}
	storage := newImageStorage(api, "project-root-id", true)
	content := []byte("png bytes")

	result, err := storage.Upload(context.Background(), bytes.NewReader(content), "image/png", ".png")
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if result.ImageURL != "https://drive.google.com/uc?export=view&id=file-id" || result.FileID != "file-id" {
		t.Fatalf("result = %#v", result)
	}
	if len(api.created) != 2 || api.created[0] != "product-images" || api.created[1] != "products" {
		t.Fatalf("created folders = %v", api.created)
	}
	if api.parents[0] != "project-root-id" || api.parents[1] != "folder-product-images" || api.parents[2] != "folder-products" {
		t.Fatalf("Drive parent folders = %v", api.parents)
	}
	if _, err := uuid.Parse(strings.TrimSuffix(api.uploadNames[0], ".png")); err != nil {
		t.Fatalf("unsafe generated filename %q", api.uploadNames[0])
	}
	if string(api.uploadedData[0]) != string(content) || len(api.publicIDs) != 1 || api.publicIDs[0] != "file-id" {
		t.Fatalf("upload/public state = %#v", api)
	}

	if _, err := storage.Upload(context.Background(), bytes.NewReader(content), "image/png", ".png"); err != nil {
		t.Fatalf("second Upload() error = %v", err)
	}
	if len(api.created) != 2 {
		t.Fatalf("folder IDs were not cached: %v", api.created)
	}
}

func TestUploadReturnsDriveErrorsWithoutCreatingSuccessResult(t *testing.T) {
	api := &fakeDriveAPI{folders: make(map[string]string), uploadErr: errors.New("provider response with sensitive details")}
	storage := newImageStorage(api, "project-root-id", false)

	result, err := storage.Upload(context.Background(), bytes.NewReader([]byte("image")), "image/png", ".png")
	if err == nil {
		t.Fatal("Upload() expected Drive API error")
	}
	if result.ImageURL != "" || result.FileID != "" || result.Name != "" {
		t.Fatalf("failed upload returned success result: %#v", result)
	}
}

func TestUploadDoesNotGrantPublicAccessWhenDisabled(t *testing.T) {
	api := &fakeDriveAPI{folders: make(map[string]string)}
	storage := newImageStorage(api, "project-root-id", false)

	if _, err := storage.Upload(context.Background(), bytes.NewReader([]byte("image")), "image/jpeg", ".jpg"); err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if len(api.publicIDs) != 0 {
		t.Fatalf("unexpected public permissions: %v", api.publicIDs)
	}
}
