package har

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

// ParseHarFromReader parses HAR data from an io.Reader.
//
// It reads all data from the reader, then parses it as a HAR document.
// Errors are wrapped using NewFileSystemError / WrapJSONUnmarshalError
// to maintain consistency with the rest of the package.
func ParseHarFromReader(r io.Reader) (*Har, error) {
	data, err := readAllFromHARReader(r)
	if err != nil {
		return nil, err
	}

	har, err := ParseHar(data)
	if err != nil {
		if harErr, ok := err.(*HarError); ok {
			return nil, harErr
		}
		return nil, WrapJSONUnmarshalError(err)
	}

	return har, nil
}

// ParseHarFromReaderWithOptions parses HAR data from an io.Reader with custom parse options.
//
// It reads all data from the reader, then delegates to ParseHarWithOptions.
func ParseHarFromReaderWithOptions(r io.Reader, options ParseOptions) (*Har, error) {
	data, err := readAllFromHARReader(r)
	if err != nil {
		return nil, err
	}

	har, err := ParseHarWithOptions(data, options)
	if err != nil {
		if harErr, ok := err.(*HarError); ok {
			return nil, harErr
		}
		return nil, WrapJSONUnmarshalError(err)
	}

	return har, nil
}

// ParseFromReader uses the functional options API to parse HAR data from an io.Reader.
//
// It reads all data from the reader, then delegates to Parse().
func ParseFromReader(r io.Reader, opts ...Option) (HARProvider, error) {
	data, err := readAllFromHARReader(r)
	if err != nil {
		return nil, err
	}

	return Parse(data, opts...)
}

// ParseHarFileGzipped parses a gzip-compressed HAR file.
//
// It opens the file, creates a gzip reader, and delegates to ParseHarFromReader.
func ParseHarFileGzipped(filePath string) (har *Har, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, NewFileSystemError(fmt.Sprintf("failed to open gzip HAR file '%s'", filePath), err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = NewFileSystemError(fmt.Sprintf("failed to close gzip HAR file '%s'", filePath), closeErr)
		}
	}()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, NewFileSystemError(fmt.Sprintf("failed to create gzip reader for '%s'", filePath), err)
	}
	gzipClosed := false
	defer func() {
		if !gzipClosed {
			if closeErr := gzReader.Close(); closeErr != nil && err == nil {
				err = NewFileSystemError(fmt.Sprintf("failed to close gzip reader for '%s'", filePath), closeErr)
			}
		}
	}()

	har, err = ParseHarFromReader(gzReader)
	if err != nil {
		return nil, err
	}

	gzipClosed = true
	if closeErr := gzReader.Close(); closeErr != nil {
		return nil, NewFileSystemError(fmt.Sprintf("failed to close gzip reader for '%s'", filePath), closeErr)
	}

	return har, nil
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
	if isGzippedByExtension(filePath) {
		return ParseHarFileGzipped(filePath)
	}

	isGzipped, err := detectGzipMagicBytes(filePath)
	if err != nil {
		return nil, err
	}
	if isGzipped {
		return ParseHarFileGzipped(filePath)
	}

	return ParseHarFile(filePath)
}

// NewStreamingParserFromReader creates a streaming entry iterator from an io.Reader.
//
// It reads all data into a buffer, then uses NewStreamingHarFromBytes
// to create the streaming parser.
func NewStreamingParserFromReader(r io.Reader, opts ...Option) (EntryIterator, error) {
	data, err := readAllFromHARReader(r)
	if err != nil {
		return nil, err
	}

	return NewStreamingParser(data, opts...)
}

func readAllFromHARReader(r io.Reader) ([]byte, error) {
	if isNilReader(r) {
		return nil, NewInvalidFormatError("reader is nil")
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, NewFileSystemError("failed to read from reader", err)
	}

	return data, nil
}

func isNilReader(r io.Reader) bool {
	if r == nil {
		return true
	}

	value := reflect.ValueOf(r)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
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
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return NewFileSystemError(fmt.Sprintf("failed to create file '%s'", filePath), err)
	}
	return writeGzippedDataToFile(file, filePath, data)
}

func writeGzippedDataToFile(file io.WriteCloser, filePath string, data []byte) (err error) {
	if file == nil {
		return NewInvalidFormatError("file is nil")
	}

	gzWriter := gzip.NewWriter(file)
	gzipClosed := false

	defer func() {
		if !gzipClosed {
			if closeErr := gzWriter.Close(); closeErr != nil && err == nil {
				err = NewFileSystemError(fmt.Sprintf("failed to flush gzip writer for '%s'", filePath), closeErr)
			}
		}
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = NewFileSystemError(fmt.Sprintf("failed to close gzip file '%s'", filePath), closeErr)
		}
	}()

	if _, err := gzWriter.Write(data); err != nil {
		return NewFileSystemError(fmt.Sprintf("failed to write gzipped data to '%s'", filePath), err)
	}

	gzipClosed = true
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
func detectGzipMagicBytes(filePath string) (isGzipped bool, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, NewFileSystemError(fmt.Sprintf("failed to open file '%s'", filePath), err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = NewFileSystemError(fmt.Sprintf("failed to close file '%s'", filePath), closeErr)
		}
	}()

	buf := make([]byte, 2)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return false, NewFileSystemError(fmt.Sprintf("failed to read file '%s'", filePath), err)
	}

	return n >= 2 && buf[0] == 0x1f && buf[1] == 0x8b, nil
}
