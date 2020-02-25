package mypod

import (
	"context"

	"github.com/parkr/radar"
	"github.com/pkg/errors"
)

type DownloadService struct {
	storageDir string
}

func NewDownloadService(storageDir string) DownloadService {
	return DownloadService{storageDir: storageDir}
}

// Create adds a RadarItem to the database.
func (ds DownloadService) Create(ctx context.Context, m radar.RadarItem) error {
	// Download URL into storage service and convert to MP3.
	// Metadata?
	return errors.New("not implemented yet")
}

// List returns a list of all radar items.
func (ds DownloadService) List(ctx context.Context, limit int) ([]radar.RadarItem, error) {
	// Not necessary.
	return nil, errors.New("not implemented yet")
}

// Delete removes a RadarItem from the database by its ID.
func (ds DownloadService) Get(ctx context.Context, id int64) (radar.RadarItem, error) {
	// Not necessary.
	return radar.RadarItem{}, errors.New("not implemented yet")
}

// Delete removes a RadarItem from the database by its ID.
func (ds DownloadService) Delete(ctx context.Context, id int64) error {
	return errors.New("not implemented yet")
}

// Shutdown closes the database connection.
func (ds DownloadService) Shutdown(ctx context.Context) {
}
