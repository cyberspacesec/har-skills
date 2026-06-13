package har

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

// --- NewHarError ---

func TestNewHarError(t *testing.T) {
	innerErr := errors.New("inner")
	err := NewHarError(ErrCodeUnknown, "something went wrong", innerErr)

	if err.Code != ErrCodeUnknown {
		t.Errorf("expected Code %d, got %d", ErrCodeUnknown, err.Code)
	}
	if err.Message != "something went wrong" {
		t.Errorf("expected Message %q, got %q", "something went wrong", err.Message)
	}
	if err.Err != innerErr {
		t.Errorf("expected Err to be innerErr")
	}
}

// --- NewFileSystemError ---

func TestNewFileSystemError(t *testing.T) {
	innerErr := errors.New("permission denied")
	err := NewFileSystemError("cannot read file", innerErr)

	if err.Code != ErrCodeFileSystem {
		t.Errorf("expected Code %d, got %d", ErrCodeFileSystem, err.Code)
	}
	if err.Message != "cannot read file" {
		t.Errorf("expected Message %q, got %q", "cannot read file", err.Message)
	}
	if err.Err != innerErr {
		t.Errorf("expected Err to be innerErr")
	}
}

// --- NewJSONParseError ---

func TestNewJSONParseError(t *testing.T) {
	innerErr := errors.New("bad json")
	err := NewJSONParseError("parse failed", innerErr)

	if err.Code != ErrCodeJSONParse {
		t.Errorf("expected Code %d, got %d", ErrCodeJSONParse, err.Code)
	}
	if err.Message != "parse failed" {
		t.Errorf("expected Message %q, got %q", "parse failed", err.Message)
	}
	if err.Err != innerErr {
		t.Errorf("expected Err to be innerErr")
	}
}

// --- WrapJSONUnmarshalError ---

func TestWrapJSONUnmarshalError_Nil(t *testing.T) {
	result := WrapJSONUnmarshalError(nil)
	if result != nil {
		t.Errorf("expected nil for nil input, got %v", result)
	}
}

func TestWrapJSONUnmarshalError_UnmarshalTypeError(t *testing.T) {
	// Use a simpler approach: just test with a non-nil error
	err := WrapJSONUnmarshalError(fmt.Errorf("cannot unmarshal string into Go value of type int"))
	if err == nil {
		t.Fatal("expected non-nil HarError")
	}
	if err.Code != ErrCodeJSONParse {
		t.Errorf("expected Code %d, got %d", ErrCodeJSONParse, err.Code)
	}
	if err.Err == nil {
		t.Error("expected Err to be set")
	}
}

func TestWrapJSONUnmarshalError_SyntaxError(t *testing.T) {
	// We can construct a real syntax error via json.Unmarshal.
	var target interface{}
	realErr := json.Unmarshal([]byte(`{invalid`), &target)
	if realErr == nil {
		t.Fatal("expected a syntax error from invalid JSON")
	}

	result := WrapJSONUnmarshalError(realErr)
	if result == nil {
		t.Fatal("expected non-nil HarError")
	}
	if result.Code != ErrCodeJSONParse {
		t.Errorf("expected Code %d, got %d", ErrCodeJSONParse, result.Code)
	}
	if _, ok := result.Metadata["offset"]; !ok {
		t.Error("expected Metadata to contain 'offset'")
	}
}

func TestWrapJSONUnmarshalError_GenericJSONError(t *testing.T) {
	err := WrapJSONUnmarshalError(errors.New("cannot unmarshal number into Go struct field"))
	if err == nil {
		t.Fatal("expected non-nil HarError")
	}
	if err.Code != ErrCodeJSONParse {
		t.Errorf("expected Code %d, got %d", ErrCodeJSONParse, err.Code)
	}
}

func TestWrapJSONUnmarshalError_DefaultFallback(t *testing.T) {
	err := WrapJSONUnmarshalError(errors.New("some other error"))
	if err == nil {
		t.Fatal("expected non-nil HarError")
	}
	if err.Code != ErrCodeJSONParse {
		t.Errorf("expected Code %d, got %d", ErrCodeJSONParse, err.Code)
	}
	if err.Message != "JSON解析错误" {
		t.Errorf("expected default message, got %q", err.Message)
	}
}

func TestWrapJSONUnmarshalError_ActualUnmarshalTypeError(t *testing.T) {
	// Trigger a real UnmarshalTypeError
	type Target struct {
		Num int `json:"num"`
	}
	var tgt Target
	err := json.Unmarshal([]byte(`{"num":"not_a_number"}`), &tgt)
	if err == nil {
		t.Fatal("expected an unmarshal error")
	}

	result := WrapJSONUnmarshalError(err)
	if result == nil {
		t.Fatal("expected non-nil HarError")
	}
	if result.Code != ErrCodeJSONParse {
		t.Errorf("expected Code %d, got %d", ErrCodeJSONParse, result.Code)
	}
	if result.Field == "" {
		t.Error("expected Field to be set for UnmarshalTypeError")
	}
}

// --- NewValidationError ---

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("value out of range", "log.entries[0].request.url")
	if err.Code != ErrCodeValidation {
		t.Errorf("expected Code %d, got %d", ErrCodeValidation, err.Code)
	}
	if err.Message != "value out of range" {
		t.Errorf("expected Message %q, got %q", "value out of range", err.Message)
	}
	if err.Field != "log.entries[0].request.url" {
		t.Errorf("expected Field %q, got %q", "log.entries[0].request.url", err.Field)
	}
	if err.Err != nil {
		t.Errorf("expected Err to be nil, got %v", err.Err)
	}
}

// --- NewInvalidFormatError ---

func TestNewInvalidFormatError(t *testing.T) {
	err := NewInvalidFormatError("invalid HAR format")
	if err.Code != ErrCodeInvalidFormat {
		t.Errorf("expected Code %d, got %d", ErrCodeInvalidFormat, err.Code)
	}
	if err.Message != "invalid HAR format" {
		t.Errorf("expected Message %q, got %q", "invalid HAR format", err.Message)
	}
}

// --- NewMissingFieldError ---

func TestNewMissingFieldError(t *testing.T) {
	err := NewMissingFieldError("log.entries")
	if err.Code != ErrCodeMissingField {
		t.Errorf("expected Code %d, got %d", ErrCodeMissingField, err.Code)
	}
	if err.Field != "log.entries" {
		t.Errorf("expected Field %q, got %q", "log.entries", err.Field)
	}
}

// --- NewInvalidValueError ---

func TestNewInvalidValueError(t *testing.T) {
	err := NewInvalidValueError("log.version", "2.0", "unsupported version")
	if err.Code != ErrCodeInvalidValue {
		t.Errorf("expected Code %d, got %d", ErrCodeInvalidValue, err.Code)
	}
	if err.Field != "log.version" {
		t.Errorf("expected Field %q, got %q", "log.version", err.Field)
	}
	if err.Metadata["value"] != "2.0" {
		t.Errorf("expected Metadata value '2.0', got %v", err.Metadata["value"])
	}
}

func TestNewInvalidValueError_NoReason(t *testing.T) {
	err := NewInvalidValueError("log.version", -1, "")
	if err.Code != ErrCodeInvalidValue {
		t.Errorf("expected Code %d, got %d", ErrCodeInvalidValue, err.Code)
	}
	// When reason is empty, the message should just be "字段值无效"
	if err.Message != "字段值无效" {
		t.Errorf("expected message without reason suffix, got %q", err.Message)
	}
}

// --- NewUnsupportedError ---

func TestNewUnsupportedError(t *testing.T) {
	err := NewUnsupportedError("compression not supported")
	if err.Code != ErrCodeUnsupported {
		t.Errorf("expected Code %d, got %d", ErrCodeUnsupported, err.Code)
	}
	if err.Message != "compression not supported" {
		t.Errorf("expected Message %q, got %q", "compression not supported", err.Message)
	}
}

// --- HarError.Error() ---

func TestHarError_Error_Basic(t *testing.T) {
	err := NewHarError(ErrCodeUnknown, "base error", nil)
	got := err.Error()
	if got != "base error" {
		t.Errorf("expected %q, got %q", "base error", got)
	}
}

func TestHarError_Error_WithField(t *testing.T) {
	err := NewHarError(ErrCodeUnknown, "bad value", nil).WithField("log.version")
	got := err.Error()
	// Format: "字段 'log.version': bad value"
	if got == "bad value" {
		t.Errorf("expected field prefix in error string, got %q", got)
	}
	if !contains(got, "log.version") {
		t.Errorf("expected error to contain field name, got %q", got)
	}
}

func TestHarError_Error_WithInnerErr(t *testing.T) {
	innerErr := errors.New("inner detail")
	err := NewHarError(ErrCodeUnknown, "outer", innerErr)
	got := err.Error()
	if !contains(got, "inner detail") {
		t.Errorf("expected error to contain inner error message, got %q", got)
	}
}

func TestHarError_Error_WithPartialErrors(t *testing.T) {
	partial1 := NewHarError(ErrCodeValidation, "field A missing", nil)
	partial2 := NewHarError(ErrCodeValidation, "field B missing", nil)
	err := NewHarError(ErrCodeUnknown, "multiple issues", nil)
	err.AddPartialError(partial1)
	err.AddPartialError(partial2)

	got := err.Error()
	if !contains(got, "field A missing") || !contains(got, "field B missing") {
		t.Errorf("expected error to contain partial error messages, got %q", got)
	}
}

// --- GetCode ---

func TestHarError_GetCode(t *testing.T) {
	tests := []struct {
		name string
		code ErrorCode
	}{
		{"Unknown", ErrCodeUnknown},
		{"FileSystem", ErrCodeFileSystem},
		{"JSONParse", ErrCodeJSONParse},
		{"InvalidFormat", ErrCodeInvalidFormat},
		{"Validation", ErrCodeValidation},
		{"MissingField", ErrCodeMissingField},
		{"InvalidValue", ErrCodeInvalidValue},
		{"Unsupported", ErrCodeUnsupported},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewHarError(tt.code, "msg", nil)
			if err.GetCode() != tt.code {
				t.Errorf("expected Code %d, got %d", tt.code, err.GetCode())
			}
		})
	}
}

// --- HasPartialErrors / GetPartialErrors ---

func TestHarError_HasPartialErrors_False(t *testing.T) {
	err := NewHarError(ErrCodeUnknown, "no partials", nil)
	if err.HasPartialErrors() {
		t.Error("expected HasPartialErrors to be false")
	}
}

func TestHarError_HasPartialErrors_True(t *testing.T) {
	err := NewHarError(ErrCodeUnknown, "with partials", nil)
	err.AddPartialError(NewHarError(ErrCodeValidation, "partial1", nil))
	if !err.HasPartialErrors() {
		t.Error("expected HasPartialErrors to be true")
	}
}

func TestHarError_GetPartialErrors_Empty(t *testing.T) {
	err := NewHarError(ErrCodeUnknown, "no partials", nil)
	partials := err.GetPartialErrors()
	if len(partials) != 0 {
		t.Errorf("expected 0 partial errors, got %d", len(partials))
	}
}

func TestHarError_GetPartialErrors_WithErrors(t *testing.T) {
	err := NewHarError(ErrCodeUnknown, "with partials", nil)
	p1 := NewHarError(ErrCodeValidation, "p1", nil)
	p2 := NewHarError(ErrCodeMissingField, "p2", nil)
	err.AddPartialError(p1)
	err.AddPartialError(p2)

	partials := err.GetPartialErrors()
	if len(partials) != 2 {
		t.Fatalf("expected 2 partial errors, got %d", len(partials))
	}
	if partials[0] != p1 {
		t.Error("expected first partial to be p1")
	}
	if partials[1] != p2 {
		t.Error("expected second partial to be p2")
	}
}

// --- IsFileSystemError ---

func TestIsFileSystemError_True(t *testing.T) {
	err := NewFileSystemError("read failed", nil)
	if !err.IsFileSystemError() {
		t.Error("expected IsFileSystemError to be true")
	}
}

func TestIsFileSystemError_False(t *testing.T) {
	err := NewJSONParseError("bad json", nil)
	if err.IsFileSystemError() {
		t.Error("expected IsFileSystemError to be false for JSONParse error")
	}
}

// --- IsJSONParseError ---

func TestIsJSONParseError_True(t *testing.T) {
	err := NewJSONParseError("bad json", nil)
	if !err.IsJSONParseError() {
		t.Error("expected IsJSONParseError to be true")
	}
}

func TestIsJSONParseError_False(t *testing.T) {
	err := NewFileSystemError("read failed", nil)
	if err.IsJSONParseError() {
		t.Error("expected IsJSONParseError to be false for FileSystem error")
	}
}

// --- IsFormatError ---

func TestIsFormatError_True(t *testing.T) {
	err := NewInvalidFormatError("bad format")
	if !err.IsFormatError() {
		t.Error("expected IsFormatError to be true")
	}
}

func TestIsFormatError_False(t *testing.T) {
	err := NewValidationError("bad value", "field")
	if err.IsFormatError() {
		t.Error("expected IsFormatError to be false for Validation error")
	}
}

// --- IsValidationError ---

func TestIsValidationError_True(t *testing.T) {
	err := NewValidationError("bad value", "field")
	if !err.IsValidationError() {
		t.Error("expected IsValidationError to be true")
	}
}

func TestIsValidationError_False(t *testing.T) {
	err := NewInvalidFormatError("bad format")
	if err.IsValidationError() {
		t.Error("expected IsValidationError to be false for Format error")
	}
}

// --- WithField ---

func TestHarError_WithField_Empty(t *testing.T) {
	err := NewHarError(ErrCodeUnknown, "msg", nil)
	result := err.WithField("log.version")
	if result.Field != "log.version" {
		t.Errorf("expected Field %q, got %q", "log.version", result.Field)
	}
}

func TestHarError_WithField_Prepend(t *testing.T) {
	err := NewHarError(ErrCodeUnknown, "msg", nil)
	err.Field = "url"
	result := err.WithField("request")
	// WithField prepends: "request.url"
	if result.Field != "request.url" {
		t.Errorf("expected Field %q, got %q", "request.url", result.Field)
	}
}

// --- WithMetadata ---

func TestHarError_WithMetadata(t *testing.T) {
	err := NewHarError(ErrCodeUnknown, "msg", nil)
	result := err.WithMetadata("key", "value")
	if result.Metadata == nil {
		t.Fatal("expected Metadata to be non-nil")
	}
	if result.Metadata["key"] != "value" {
		t.Errorf("expected Metadata[key] = %q, got %v", "value", result.Metadata["key"])
	}
}

func TestHarError_WithMetadata_MultipleKeys(t *testing.T) {
	err := NewHarError(ErrCodeUnknown, "msg", nil)
	err.WithMetadata("k1", "v1").WithMetadata("k2", 42)
	if err.Metadata["k1"] != "v1" {
		t.Errorf("expected k1=v1, got %v", err.Metadata["k1"])
	}
	if err.Metadata["k2"] != 42 {
		t.Errorf("expected k2=42, got %v", err.Metadata["k2"])
	}
}

// --- AddPartialError ---

func TestHarError_AddPartialError(t *testing.T) {
	err := NewHarError(ErrCodeUnknown, "main", nil)
	p := NewHarError(ErrCodeValidation, "partial", nil)
	result := err.AddPartialError(p)
	if result != err {
		t.Error("expected AddPartialError to return the same HarError for chaining")
	}
	if len(err.PartialErrors) != 1 {
		t.Fatalf("expected 1 partial error, got %d", len(err.PartialErrors))
	}
	if err.PartialErrors[0] != p {
		t.Error("expected PartialErrors[0] to be p")
	}
}

// --- ParseOptions ---

func TestDefaultParseOptions(t *testing.T) {
	opts := DefaultParseOptions()
	if opts.Lenient != false {
		t.Error("expected Lenient to be false")
	}
	if opts.SkipValidation != false {
		t.Error("expected SkipValidation to be false")
	}
	if opts.CollectWarnings != false {
		t.Error("expected CollectWarnings to be false")
	}
	if opts.MaxWarnings != 100 {
		t.Errorf("expected MaxWarnings to be 100, got %d", opts.MaxWarnings)
	}
}

// --- ErrorCode constants ---

func TestErrorCodeValues(t *testing.T) {
	codes := []struct {
		name  string
		code  ErrorCode
		value int
	}{
		{"ErrCodeUnknown", ErrCodeUnknown, 0},
		{"ErrCodeFileSystem", ErrCodeFileSystem, 1},
		{"ErrCodeJSONParse", ErrCodeJSONParse, 2},
		{"ErrCodeInvalidFormat", ErrCodeInvalidFormat, 3},
		{"ErrCodeValidation", ErrCodeValidation, 4},
		{"ErrCodeMissingField", ErrCodeMissingField, 5},
		{"ErrCodeInvalidValue", ErrCodeInvalidValue, 6},
		{"ErrCodeUnsupported", ErrCodeUnsupported, 7},
	}

	for _, tt := range codes {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.code) != tt.value {
				t.Errorf("expected %s = %d, got %d", tt.name, tt.value, int(tt.code))
			}
		})
	}
}

// --- Combined test for Error() method with all components ---

func TestHarError_Error_AllComponents(t *testing.T) {
	innerErr := errors.New("root cause")
	partial := NewHarError(ErrCodeMissingField, "missing url", nil).WithField("request")
	err := NewHarError(ErrCodeFileSystem, "cannot parse", innerErr)
	err.Field = "log"
	err.AddPartialError(partial)

	got := err.Error()
	if !contains(got, "log") {
		t.Errorf("expected error to contain field 'log', got %q", got)
	}
	if !contains(got, "cannot parse") {
		t.Errorf("expected error to contain message, got %q", got)
	}
	if !contains(got, "root cause") {
		t.Errorf("expected error to contain inner error, got %q", got)
	}
	if !contains(got, "missing url") {
		t.Errorf("expected error to contain partial error, got %q", got)
	}
}

// helper
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
