// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package googledrive

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	"github.com/dedysetiawan/product-catalog-project/internal/config"
	"github.com/dedysetiawan/product-catalog-project/internal/modules/product/domain"
)

const (
	productImagesFolderName = "product-images"
	productsFolderName      = "products"
)

type driveAPI interface {
	findFolder(ctx context.Context, parentID, name string) (string, error)
	createFolder(ctx context.Context, parentID, name string) (string, error)
	uploadFile(ctx context.Context, parentID, name, contentType string, content io.Reader) (id, storedName string, err error)
	makePublic(ctx context.Context, fileID string) error
}

type imageStorage struct {
	api          driveAPI
	rootFolderID string
	publicImages bool

	mu               sync.Mutex
	productsFolderID string
}

func NewGoogleDriveStorage(ctx context.Context, cfg config.GoogleDriveConfig) (domain.ProductImageStorage, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.OAuthClientID,
		ClientSecret: cfg.OAuthClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{drive.DriveFileScope},
	}
	token := &oauth2.Token{RefreshToken: cfg.RefreshToken}
	client := oauthConfig.Client(ctx, token)
	service, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("initialize Google Drive client: %w", err)
	}
	return newImageStorage(&googleDriveAPI{service: service}, cfg.RootFolderID, cfg.PublicImages), nil
}

func newImageStorage(api driveAPI, rootFolderID string, publicImages bool) *imageStorage {
	return &imageStorage{api: api, rootFolderID: rootFolderID, publicImages: publicImages}
}

func (s *imageStorage) Upload(ctx context.Context, content io.Reader, contentType, extension string) (domain.ProductImageUploadResult, error) {
	expectedType, ok := imageTypeForExtension(extension)
	if !ok || expectedType != contentType {
		return domain.ProductImageUploadResult{}, errors.New("unsupported product image type")
	}
	folderID, err := s.getProductsFolder(ctx)
	if err != nil {
		return domain.ProductImageUploadResult{}, fmt.Errorf("resolve product image folder: %w", err)
	}

	name := uuid.NewString() + extension
	fileID, storedName, err := s.api.uploadFile(ctx, folderID, name, contentType, content)
	if err != nil {
		return domain.ProductImageUploadResult{}, fmt.Errorf("upload product image: %w", err)
	}
	if fileID == "" {
		return domain.ProductImageUploadResult{}, errors.New("Google Drive returned an empty file ID")
	}
	if s.publicImages {
		if err := s.api.makePublic(ctx, fileID); err != nil {
			return domain.ProductImageUploadResult{}, fmt.Errorf("make product image public: %w", err)
		}
	}
	if storedName == "" {
		storedName = name
	}
	imageURL := "https://drive.google.com/uc?" + url.Values{
		"export": {"view"},
		"id":     {fileID},
	}.Encode()
	return domain.ProductImageUploadResult{ImageURL: imageURL, FileID: fileID, Name: storedName}, nil
}

func (s *imageStorage) getProductsFolder(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.productsFolderID != "" {
		return s.productsFolderID, nil
	}
	imageFolderID, err := s.findOrCreateFolder(ctx, s.rootFolderID, productImagesFolderName)
	if err != nil {
		return "", err
	}
	folderID, err := s.findOrCreateFolder(ctx, imageFolderID, productsFolderName)
	if err != nil {
		return "", err
	}
	if folderID == "" {
		return "", errors.New("Google Drive returned an empty products folder ID")
	}
	s.productsFolderID = folderID
	return folderID, nil
}

func (s *imageStorage) findOrCreateFolder(ctx context.Context, parentID, name string) (string, error) {
	folderID, err := s.api.findFolder(ctx, parentID, name)
	if err != nil {
		return "", err
	}
	if folderID == "" {
		folderID, err = s.api.createFolder(ctx, parentID, name)
		if err != nil {
			return "", err
		}
	}
	if folderID == "" {
		return "", errors.New("Google Drive returned an empty folder ID")
	}
	return folderID, nil
}

type googleDriveAPI struct {
	service *drive.Service
}

func (a *googleDriveAPI) findFolder(ctx context.Context, parentID, name string) (string, error) {
	query := fmt.Sprintf(
		"'%s' in parents and name = '%s' and mimeType = '%s' and trashed = false",
		escapeQueryValue(parentID),
		escapeQueryValue(name),
		"application/vnd.google-apps.folder",
	)
	files, err := a.service.Files.List().
		Q(query).
		Spaces("drive").
		PageSize(100).
		Fields("files(id,name)").
		Context(ctx).
		Do()
	if err != nil {
		return "", err
	}
	if len(files.Files) == 0 {
		return "", nil
	}
	return files.Files[0].Id, nil
}

func (a *googleDriveAPI) createFolder(ctx context.Context, parentID, name string) (string, error) {
	folder, err := a.service.Files.Create(&drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{parentID},
	}).Fields("id").Context(ctx).Do()
	if err != nil {
		return "", err
	}
	return folder.Id, nil
}

func (a *googleDriveAPI) uploadFile(ctx context.Context, parentID, name, contentType string, content io.Reader) (string, string, error) {
	file, err := a.service.Files.Create(&drive.File{
		Name:     name,
		MimeType: contentType,
		Parents:  []string{parentID},
	}).Media(content, googleapi.ContentType(contentType)).Fields("id,name").Context(ctx).Do()
	if err != nil {
		return "", "", err
	}
	return file.Id, file.Name, nil
}

func (a *googleDriveAPI) makePublic(ctx context.Context, fileID string) error {
	_, err := a.service.Permissions.Create(fileID, &drive.Permission{
		Type: "anyone",
		Role: "reader",
	}).Fields("id").Context(ctx).Do()
	return err
}

func escapeQueryValue(value string) string {
	return strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(value)
}

func imageTypeForExtension(extension string) (string, bool) {
	switch strings.ToLower(extension) {
	case ".jpg", ".jpeg":
		return "image/jpeg", true
	case ".png":
		return "image/png", true
	case ".webp":
		return "image/webp", true
	default:
		return "", false
	}
}
