package mypod

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestGetCommandArgs(t *testing.T) {
	ds := NewDownloadService("/strg", []string{"--foo", "--bar"})

	actual := ds.getCommandArgs("https://example.com")
	expected := []string{"--cookies", filepath.Join(ds.storageDir, "yt-dl-cookies.txt"),
		"--abort-on-error",      // tell me if something went wrong
		"--extract-audio",       // just audio
		"--audio-format", "m4a", // m4a format
		"--audio-quality", "0", // best audio quality
		"--add-metadata",    // add metadata to file
		"--embed-thumbnail", // add the thumbnail as cover art
		"--write-thumbnail", // write thumbnail to file as well
		"--embed-chapters",  // embed chapters into the output file
		"--exec", fmt.Sprintf("touch {} && mv {} %q", filepath.Join(ds.storageDir, "files")),
		"--foo",
		"--bar",
		"https://example.com",
	}
	if len(actual) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}
