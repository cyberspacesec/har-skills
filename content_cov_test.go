package har

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Cover Content.SaveToFile file-write error branch (lines 373-375).
// os.WriteFile to a path under a non-existent directory fails, surfacing
// as a wrapped *HarError with ErrCodeFileSystem.
func TestCovContentSaveToFile_WriteError(t *testing.T) {
	c := &Content{
		Size:     5,
		MimeType: "text/plain",
		Text:     "hello",
	}

	// A path whose parent directory does not exist -> WriteFile fails.
	badPath := "/nonexistent_dir_xyz/some/deep/path/content.txt"
	err := c.SaveToFile(badPath)
	if err == nil {
		t.Fatal("expected error writing to non-existent path, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeFileSystem)
}

// Cover Content.SaveToFile nil-receiver branch (line 368).
func TestCovContentSaveToFile_NilReceiver(t *testing.T) {
	var c *Content
	err := c.SaveToFile("/tmp/whatever_cov.txt")
	if err == nil {
		t.Fatal("expected error for nil content, got nil")
	}
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

// Cover Content.SaveToFile happy path with empty data (line 376-379).
func TestCovContentSaveToFile_EmptyContent(t *testing.T) {
	c := &Content{
		Size:     0,
		MimeType: "text/plain",
		// Text empty and no Encoding -> DecodeContent returns nil data
	}
	path := t.TempDir() + "/empty-content.txt"
	if err := c.SaveToFile(path); err != nil {
		t.Fatalf("unexpected error saving empty content: %v", err)
	}
	assert.FileExists(t, path)
}

// Cover Content.SaveToFile DecodeContent error branch (lines 373-375):
// invalid base64 encoding makes DecodeContent fail, which SaveToFile
// propagates directly.
func TestCovContentSaveToFile_DecodeError(t *testing.T) {
	c := &Content{
		Size:     10,
		MimeType: "text/plain",
		Text:     "!!!!not-base64!!!!",
		Encoding: "base64", // forces base64 decode path, which fails on invalid input
	}
	err := c.SaveToFile(t.TempDir() + "/decode-err.txt")
	if err == nil {
		t.Fatal("expected error from invalid base64 content, got nil")
	}
}
