package mypod

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/parkr/radar"
	"github.com/pkg/errors"
	"github.com/technoweenie/grohl"
)

type DownloadService struct {
	storageDir string
}

func NewDownloadService(storageDir string) DownloadService {
	return DownloadService{storageDir: storageDir}
}

// Create adds a RadarItem to the database.
func (ds DownloadService) Create(ctx context.Context, m radar.RadarItem) error {
	// Generate temporary dir to download in.
	tmpDir, err := ioutil.TempDir("", "mypod")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir) // clean up

	// Gonna be really naughty and ignore the context... downloads take a long time.
	// Download URL into storage service and convert to M4A.
	dlCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(dlCtx,
		"youtube-dl",
		"--abort-on-error",
		"--extract-audio",
		"--audio-format", "m4a",
		"--audio-quality", "0",
		"--xattrs",
		"--exec", fmt.Sprintf("mv {} \"%s\"", filepath.Join(ds.storageDir, "files")),
		m.URL,
	)
	cmd.Dir = tmpDir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	grohl.Log(grohl.Data{
		"msg":  "starting download",
		"url":  m.URL,
		"dir":  tmpDir,
		"args": cmd.Args,
	})
	if err := cmd.Run(); err != nil {
		return err
	}
	grohl.Log(grohl.Data{
		"msg":            "completed download",
		"url":            m.URL,
		"dir":            tmpDir,
		"elapsed_user":   cmd.ProcessState.UserTime().String(),
		"elapsed_system": cmd.ProcessState.SystemTime().String(),
	})

	// Metadata?
	return err
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
