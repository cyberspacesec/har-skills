package har

import (
	"fmt"
	"io"
)

func readAndCloseResponseBody(body io.ReadCloser) ([]byte, error, error) {
	if isNilReader(body) {
		return nil, nil, nil
	}

	bodyBytes, readErr := io.ReadAll(body)
	closeErr := body.Close()
	return bodyBytes, readErr, closeErr
}

func responseBodyErrorMessage(readErr, closeErr error) string {
	if readErr == nil && closeErr == nil {
		return ""
	}

	if readErr != nil && closeErr != nil {
		return fmt.Sprintf("failed to read response body: %v; failed to close response body: %v", readErr, closeErr)
	}
	if readErr != nil {
		return fmt.Sprintf("failed to read response body: %v", readErr)
	}
	return fmt.Sprintf("failed to close response body: %v", closeErr)
}
