package har

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// ParseHarFromReader parses HAR data from an io.Reader.
//
// It reads all data from the reader, then parses it as a HAR document.
// Errors are wrapped using NewFileSystemError / WrapJSONUnmarshalError
// to maintain consistency with the rest of the package.
func ParseHarFromReader(r io.Reader) (*Har, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, NewFileSystemError("failed to read from reader", err)
	}

	har, err := ParseHar(data)
	if err != nil {
		if harErr, ok := err.(*HarError); ok && harErr.IsJSONParseError() {
			return nil, err
		}
		return nil, WrapJSONUnmarshalError(err)
	}

	return har, nil
}

// ParseHarFromReaderWithOptions parses HAR data from an io.Reader with custom parse options.
//
// It reads all data from the reader, then delegates to ParseHarWithOptions.
func ParseHarFromReaderWithOptions(r io.Reader, options ParseOptions) (*Har, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, NewFileSystemError("failed to read from reader", err)
	}

	har, err := ParseHarWithOptions(data, options)
	if err != nil {
		if harErr, ok := err.(*HarError); ok && harErr.IsJSONParseError() {
			return nil, err
		}
		return nil, WrapJSONUnmarshalError(err)
	}

	return har, nil
}

// ParseFromReader uses the functional options API to parse HAR data from an io.Reader.
//
// It reads all data from the reader, then delegates to Parse().
func ParseFromReader(r io.Reader, opts ...Option) (HARProvider, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, NewFileSystemError("failed to read from reader", err)
	}

	return Parse(data, opts...)
}

// ParseHarFileGzipped parses a gzip-compressed HAR file.
//
// It opens the file, creates a gzip reader, and delegates to ParseHarFromReader.
func ParseHarFileGzipped(filePath string) (*Har, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, NewFileSystemError(fmt.Sprintf("failed to open gzip HAR file '%s'", filePath), err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, NewFileSystemError(fmt.Sprintf("failed to create gzip reader for '%s'", filePath), err)
	}
	defer gzReader.Close()

	return ParseHarFromReader(gzReader)
}

// ParseHarFileAuto auto-detects whether a HAR file is gzipped and parses accordingly.
//
// Detection strategy:
//   - File extension: .har.gz or .har.gzip are treated as gzipped
//   - Magic bytes: if the first two bytes are 0x1f 0x8b, the file is treated as gzipped
//
// Otherwise, the file is parsed as a plain HAR file.
func ParseHarFileAuto(filePath string) (*Har, error) {
	// Check by file extension first
	if strings.HasSuffix(filePath, ".har.gz") || strings.HasSuffix(filePath, ".har.gzip") {
		return ParseHarFileGzipped(filePath)
	}

	// Open the file to check magic bytes
	file, err := os.Open(filePath)
	if err != nil {
		return nil, NewFileSystemError(fmt.Sprintf("failed to open HAR file '%s'", filePath), err)
	}
	defer file.Close()

	// Read the first two bytes for magic number detection
	buf := make([]byte, 2)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return nil, NewFileSystemError(fmt.Sprintf("failed to read HAR file '%s'", filePath), err)
	}

	// Check gzip magic bytes: 0x1f 0x8b
	if n >= 2 && buf[0] == 0x1f && buf[1] == 0x8b {
		// Rewind and parse as gzipped
		_, seekErr := file.Seek(0, io.SeekStart)
		if seekErr != nil {
			return nil, NewFileSystemError(fmt.Sprintf("failed to seek in HAR file '%s'", filePath), seekErr)
		}

		gzReader, gzErr := gzip.NewReader(file)
		if gzErr != nil {
			return nil, NewFileSystemError(fmt.Sprintf("failed to create gzip reader for '%s'", filePath), gzErr)
		}
		defer gzReader.Close()

		return ParseHarFromReader(gzReader)
	}

	// Not gzipped; close the file and parse as plain HAR
	file.Close()
	return ParseHarFile(filePath)
}

// NewStreamingParserFromReader creates a streaming entry iterator from an io.Reader.
//
// It reads all data into a buffer, then uses NewStreamingHarFromBytes
// to create the streaming parser.
func NewStreamingParserFromReader(r io.Reader, opts ...Option) (EntryIterator, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, NewFileSystemError("failed to read from reader", err)
	}

	streamingHar, err := NewStreamingHarFromBytes(data)
	if err != nil {
		return nil, err
	}

	return streamingHar.Entries(), nil
}

// SaveToFileGzipped saves a HAR object as a gzip-compressed file.
//
// It marshals the HAR to JSON, applies gzip compression, and writes
// the result to the specified file path.
func SaveToFileGzipped(har *Har, filePath string, indent bool) error {
	if har == nil {
		return NewInvalidFormatError("HAR object is nil")
	}

	data, err := har.ToJSON(indent)
	if err != nil {
		return WrapJSONUnmarshalError(err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return NewFileSystemError(fmt.Sprintf("failed to create file '%s'", filePath), err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	if _, err := gzWriter.Write(data); err != nil {
		return NewFileSystemError(fmt.Sprintf("failed to write gzipped data to '%s'", filePath), err)
	}

	// Flush the gzip writer to ensure all data is written
	if err := gzWriter.Close(); err != nil {
		return NewFileSystemError(fmt.Sprintf("failed to flush gzip writer for '%s'", filePath), err)
	}

	return nil
}

// isGzippedByExtension checks if a file path has a gzip extension.
func isGzippedByExtension(filePath string) bool {
	return strings.HasSuffix(filePath, ".har.gz") || strings.HasSuffix(filePath, ".har.gzip")
}

// detectGzipMagicBytes reads the first two bytes from a file and checks
// for the gzip magic number (0x1f 0x8b). The file is closed after detection.
func detectGzipMagicBytes(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, NewFileSystemError(fmt.Sprintf("failed to open file '%s'", filePath), err)
	}
	defer file.Close()

	buf := make([]byte, 2)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return false, NewFileSystemError(fmt.Sprintf("failed to read file '%s'", filePath), err)
	}

	return n >= 2 && buf[0] == 0x1f && buf[1] == 0x8b, nil
}

// unmarshalHarFromBytes is a helper to unmarshal HAR bytes with error wrapping.
func unmarshalHarFromBytes(data []byte) (*Har, error) {
	har := new(Har)
	if err := json.Unmarshal(data, har); err != nil {
		return nil, WrapJSONUnmarshalError(err)
	}
	return har, nil
}
