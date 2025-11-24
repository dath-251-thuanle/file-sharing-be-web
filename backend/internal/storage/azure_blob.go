package storage

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
)

type AzureBlobStorage struct {
	client           *azblob.Client
	endpoint         string
	publicContainer  string
	privateContainer string
}

func NewAzureBlobStorage(endpoint, accountName, accountKey, publicContainer, privateContainer string) (*AzureBlobStorage, error) {
	if endpoint == "" || accountName == "" || accountKey == "" {
		return nil, fmt.Errorf("azure blob: missing endpoint or credentials")
	}
	cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return nil, fmt.Errorf("azure blob: credential error: %w", err)
	}
	client, err := azblob.NewClientWithSharedKeyCredential(endpoint, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("azure blob: client init failed: %w", err)
	}
	return &AzureBlobStorage{
		client:           client,
		endpoint:         strings.TrimSuffix(endpoint, "/"),
		publicContainer:  publicContainer,
		privateContainer: privateContainer,
	}, nil
}

func (s *AzureBlobStorage) Upload(ctx context.Context, obj *Object) (*Location, error) {
	if err := ValidateObject(obj); err != nil {
		return nil, err
	}
	container, err := s.containerName(obj.Container)
	if err != nil {
		return nil, err
	}

	blobName, err := sanitizeBlobPath(obj.Name)
	if err != nil {
		return nil, err
	}

	options := &azblob.UploadStreamOptions{}
	if obj.ContentType != "" {
		options.HTTPHeaders = &blob.HTTPHeaders{
			BlobContentType: &obj.ContentType,
		}
	}

	if _, err := s.client.UploadStream(ctx, container, blobName, obj.Reader, options); err != nil {
		return nil, fmt.Errorf("azure blob: upload failed: %w", err)
	}

	loc := &Location{
		Container: obj.Container,
		Path:      blobName,
		URL:       fmt.Sprintf("%s/%s/%s", s.endpoint, container, blobName),
	}
	return loc, nil
}

func (s *AzureBlobStorage) Download(ctx context.Context, loc *Location) (*DownloadResult, error) {
	if err := ValidateLocation(loc); err != nil {
		return nil, err
	}
	container, err := s.containerName(loc.Container)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.DownloadStream(ctx, container, loc.Path, nil)
	if err != nil {
		return nil, fmt.Errorf("azure blob: download failed: %w", err)
	}

	contentType := ""
	if resp.ContentType != nil {
		contentType = *resp.ContentType
	}

	size := int64(0)
	if resp.ContentLength != nil {
		size = *resp.ContentLength
	}

	return &DownloadResult{
		Reader:      resp.Body,
		ContentType: contentType,
		Size:        size,
	}, nil
}

func (s *AzureBlobStorage) Delete(ctx context.Context, loc *Location) error {
	if err := ValidateLocation(loc); err != nil {
		return err
	}
	container, err := s.containerName(loc.Container)
	if err != nil {
		return err
	}
	if _, err := s.client.DeleteBlob(ctx, container, loc.Path, nil); err != nil {
		return fmt.Errorf("azure blob: delete failed: %w", err)
	}
	return nil
}

func (s *AzureBlobStorage) containerName(ct ContainerType) (string, error) {
	switch ct {
	case ContainerPublic:
		if s.publicContainer == "" {
			return "", fmt.Errorf("azure blob: public container not configured")
		}
		return s.publicContainer, nil
	case ContainerPrivate:
		if s.privateContainer == "" {
			return "", fmt.Errorf("azure blob: private container not configured")
		}
		return s.privateContainer, nil
	default:
		return "", fmt.Errorf("azure blob: unknown container %q", ct)
	}
}

func sanitizeBlobPath(name string) (string, error) {
	clean := path.Clean(name)
	if clean == "." || clean == "/" {
		return "", fmt.Errorf("azure blob: invalid blob name")
	}
	if strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("azure blob: path traversal detected")
	}
	return clean, nil
}
