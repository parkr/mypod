package main

import "testing"

func TestReadAdditionalYtdlArgs(t *testing.T) {
	storage := "testdata/storage"
	actual := readAdditionalYtdlArgs(storage)
	expected := []string{"--foo", "--bar", "--baz"}
	if len(actual) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

func TestReadAdditionalYtdlArgs_NonStringCausesFailure(t *testing.T) {
	storage := "testdata/storage-bad-yt-dlp-args"
	actual := readAdditionalYtdlArgs(storage)
	expected := []string{}
	if len(actual) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}
