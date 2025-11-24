package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
)

// Sentinel errors to help callers distinguish failure reasons.
var (
	ErrInvalidObject   = errors.New("storage: invalid object")
	ErrInvalidLocation = errors.New("storage: invalid location")
)

type ContainerType string

const (
	ContainerPublic  ContainerType = "public"
	ContainerPrivate ContainerType = "private"
)

func (c ContainerType) String() string {
	switch c {
	case ContainerPublic, ContainerPrivate:
		return string(c)
	default:
		return "unknown"
	}
}

func (c ContainerType) IsValid() bool {
	return c == ContainerPublic || c == ContainerPrivate
}

// Object represents the payload sent to a storage backend when uploading.
type Object struct {
	Name        string
	Container   ContainerType
	ContentType string
	Size        int64
	Reader      io.Reader
}

// Location represents where an object is stored inside the backend.
type Location struct {
	Container ContainerType
	Path      string
	URL       string
}

// DownloadResult bundles the stream returned by a storage backend and some metadata.
type DownloadResult struct {
	Reader      io.ReadCloser
	ContentType string
	Size        int64
}

// Storage describes the basic operations supported by every storage backend we implement.
type Storage interface {
	Upload(ctx context.Context, obj *Object) (*Location, error)
	Download(ctx context.Context, loc *Location) (*DownloadResult, error)
	Delete(ctx context.Context, loc *Location) error
}

// ValidateObject performs a light validation of the input object before delegating to providers.
func ValidateObject(obj *Object) error {
	if obj == nil || obj.Reader == nil {
		return fmt.Errorf("%w: missing data stream", ErrInvalidObject)
	}
	if !obj.Container.IsValid() {
		return fmt.Errorf("%w: invalid container %q", ErrInvalidObject, obj.Container)
	}
	if obj.Name == "" {
		return fmt.Errorf("%w: missing object name", ErrInvalidObject)
	}
	return nil
}

// ValidateLocation ensures we only interact with safe locations.
func ValidateLocation(loc *Location) error {
	if loc == nil {
		return ErrInvalidLocation
	}
	if !loc.Container.IsValid() {
		return fmt.Errorf("%w: invalid container %q", ErrInvalidLocation, loc.Container)
	}
	if loc.Path == "" {
		return fmt.Errorf("%w: missing path", ErrInvalidLocation)
	}
	return nil
}
