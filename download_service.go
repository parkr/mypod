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
		"--cookies", filepath.Join(ds.storageDir, "yt-dl-cookies.txt"),
		"--abort-on-error",      // tell me if something went wrong
		"--extract-audio",       // just audio
		"--audio-format", "m4a", // m4a format
		"--audio-quality", "0", // best audio quality
		"--add-metadata",    // add metadata to file
		"--embed-thumbnail", // add the thumbnail as cover art
		"--write-thumbnail", // write thumbnail to file as well
		"--exec", fmt.Sprintf("touch {} && mv {} %q", filepath.Join(ds.storageDir, "files")),
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

	// Move thumbnail (and other files generated) to images dir
	fileInfos, err := ioutil.ReadDir(tmpDir)
	if err != nil {
		return err
	}
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			continue
		}

		grohl.Log(grohl.Data{
			"msg":  "moving thumbnail",
			"name": fileInfo.Name(),
		})
		err := os.Rename(
			filepath.Join(tmpDir, fileInfo.Name()),
			filepath.Join(ds.storageDir, "images", fileInfo.Name()),
		)
		if err != nil {
			return err
		}
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

// Shutdown closes the database connection.
func (ds DownloadService) Shutdown(ctx context.Context) {
}
