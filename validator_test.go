package har

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateHarFile(t *testing.T) {
	// 测试有效的HAR文件
	t.Run("ValidHar", func(t *testing.T) {
		har := &Har{
			Log: Log{
				Version: HarSpecVersion12,
				Creator: Creator{
					Name:    "Test",
					Version: "1.0",
				},
				Entries: []Entries{
					{
						StartedDateTime: time.Now(),
						Request: Request{
							Method:      "GET",
							URL:         "https://example.com",
							HTTPVersion: "HTTP/1.1",
						},
						Response: Response{
							Status:      200,
							StatusText:  "OK",
							HTTPVersion: "HTTP/1.1",
							Content: Content{
								Size:     100,
								MimeType: "text/html",
							},
						},
					},
				},
			},
		}

		err := ValidateHarFile(har)
		assert.NoError(t, err)
	})

	// 测试缺少必要字段
	t.Run("MissingRequiredFields", func(t *testing.T) {
		har := &Har{
			Log: Log{
				// 缺少 Version
				Creator: Creator{
					// 缺少 Name
					Version: "1.0",
				},
				Entries: []Entries{},
			},
		}

		err := ValidateHarFile(har)
		assert.Error(t, err)

		harErr, ok := err.(*HarError)
		require.True(t, ok)
		assert.Equal(t, ErrCodeValidation, harErr.Code)
		assert.True(t, harErr.HasPartialErrors())
		assert.GreaterOrEqual(t, len(harErr.GetPartialErrors()), 2) // 至少有两个错误（缺少版本和创建者名称）
	})

	// 测试不支持的版本
	t.Run("UnsupportedVersion", func(t *testing.T) {
		har := &Har{
			Log: Log{
				Version: "2.0", // 不支持的版本
				Creator: Creator{
					Name:    "Test",
					Version: "1.0",
				},
				Entries: []Entries{},
			},
		}

		err := ValidateHarFile(har)
		assert.Error(t, err)

		harErr, ok := err.(*HarError)
		require.True(t, ok)
		assert.Equal(t, ErrCodeValidation, harErr.Code)
		assert.True(t, harErr.HasPartialErrors())

		// 验证是否有关于不支持版本的错误
		found := false
		for _, pe := range harErr.GetPartialErrors() {
			if pe.Field == "log.version" {
				found = true
				break
			}
		}
		assert.True(t, found, "应该有关于不支持版本的错误")
	})

	// 测试条目验证
	t.Run("InvalidEntries", func(t *testing.T) {
		har := &Har{
			Log: Log{
				Version: HarSpecVersion12,
				Creator: Creator{
					Name:    "Test",
					Version: "1.0",
				},
				Entries: []Entries{
					{
						// 缺少 StartedDateTime
						Request: Request{
							// 缺少 Method
							URL: "https://example.com",
							// 缺少 HTTPVersion
						},
						Response: Response{
							Status: 200,
							// 缺少 HTTPVersion
							Content: Content{
								Size: 100,
								// 缺少 MimeType
							},
						},
					},
				},
			},
		}

		err := ValidateHarFile(har)
		assert.Error(t, err)

		harErr, ok := err.(*HarError)
		require.True(t, ok)
		assert.Equal(t, ErrCodeValidation, harErr.Code)
		assert.True(t, harErr.HasPartialErrors())
	})

	// 测试无效的URL
	t.Run("InvalidURL", func(t *testing.T) {
		har := &Har{
			Log: Log{
				Version: HarSpecVersion12,
				Creator: Creator{
					Name:    "Test",
					Version: "1.0",
				},
				Entries: []Entries{
					{
						StartedDateTime: time.Now(),
						Request: Request{
							Method:      "GET",
							URL:         "://invalid-url", // 无效的URL
							HTTPVersion: "HTTP/1.1",
						},
						Response: Response{
							Status:      200,
							HTTPVersion: "HTTP/1.1",
							Content: Content{
								Size:     100,
								MimeType: "text/html",
							},
						},
					},
				},
			},
		}

		err := ValidateHarFile(har)
		assert.Error(t, err)

		harErr, ok := err.(*HarError)
		require.True(t, ok)
		assert.True(t, harErr.HasPartialErrors())

		// 验证是否有关于无效URL的错误
		found := false
		for _, pe := range harErr.GetPartialErrors() {
			if pe.Field == "log.entries[0].request.url" {
				found = true
				break
			}
		}
		assert.True(t, found, "应该有关于无效URL的错误")
	})
}

func TestVersionDetection(t *testing.T) {
	// 测试版本检测
	testCases := []struct {
		name            string
		version         string
		expectedVersion string
	}{
		{"ExactVersion11", HarSpecVersion11, HarSpecVersion11},
		{"ExactVersion12", HarSpecVersion12, HarSpecVersion12},
		{"ExactVersion13", HarSpecVersion13, HarSpecVersion13},
		{"PrefixVersion11", "1.1.2", HarSpecVersion11},
		{"PrefixVersion12", "1.2.1", HarSpecVersion12},
		{"PrefixVersion13", "1.3.0", HarSpecVersion13},
		{"InvalidVersion", "0.9", HarSpecVersion12}, // 默认为1.2
		{"EmptyVersion", "", HarSpecVersion12},      // 默认为1.2
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			har := &Har{
				Log: Log{
					Version: tc.version,
				},
			}
			detected := DetectHarVersion(har)
			assert.Equal(t, tc.expectedVersion, detected)
		})
	}
}

func TestHarVersionOptions(t *testing.T) {
	// 测试指定版本选项
	t.Run("WithHarVersion", func(t *testing.T) {
		opts := applyOptions(WithHarVersion(HarSpecVersion11))
		assert.Equal(t, HarSpecVersion11, opts.harVersion)
		assert.False(t, opts.autoDetectVersion)
	})

	// 测试指定无效版本
	t.Run("WithInvalidVersion", func(t *testing.T) {
		opts := applyOptions(WithHarVersion("0.9"))
		assert.Equal(t, HarSpecVersion12, opts.harVersion) // 应该保持默认值
		assert.True(t, opts.autoDetectVersion)             // 应该保持默认值
	})

	// 测试禁用自动检测
	t.Run("DisableAutoDetect", func(t *testing.T) {
		opts := applyOptions(WithAutoDetectVersion(false))
		assert.False(t, opts.autoDetectVersion)
	})
}

func TestValidatePostDataAndQueryString(t *testing.T) {
	t.Run("MissingPostDataMimeType", func(t *testing.T) {
		har := &Har{
			Log: Log{
				Version: HarSpecVersion12,
				Creator: Creator{Name: "Test", Version: "1.0"},
				Entries: []Entries{
					{
						StartedDateTime: time.Now(),
						Request: Request{
							Method:      "POST",
							URL:         "https://example.com/api",
							HTTPVersion: "HTTP/1.1",
							PostData:    &PostData{MimeType: "", Text: "data"},
						},
						Response: Response{
							Status:      200,
							HTTPVersion: "HTTP/1.1",
							Content:     Content{Size: 100, MimeType: "text/html"},
						},
					},
				},
			},
		}
		err := ValidateHarFile(har)
		assert.Error(t, err)
	})

	t.Run("ValidPostData", func(t *testing.T) {
		har := &Har{
			Log: Log{
				Version: HarSpecVersion12,
				Creator: Creator{Name: "Test", Version: "1.0"},
				Entries: []Entries{
					{
						StartedDateTime: time.Now(),
						Request: Request{
							Method:      "POST",
							URL:         "https://example.com/api",
							HTTPVersion: "HTTP/1.1",
							PostData:    &PostData{MimeType: "application/json", Text: `{"key":"value"}`},
						},
						Response: Response{
							Status:      200,
							HTTPVersion: "HTTP/1.1",
							Content:     Content{Size: 100, MimeType: "text/html"},
						},
					},
				},
			},
		}
		err := ValidateHarFile(har)
		assert.NoError(t, err)
	})

	t.Run("MissingQueryStringName", func(t *testing.T) {
		har := &Har{
			Log: Log{
				Version: HarSpecVersion12,
				Creator: Creator{Name: "Test", Version: "1.0"},
				Entries: []Entries{
					{
						StartedDateTime: time.Now(),
						Request: Request{
							Method:      "GET",
							URL:         "https://example.com/api",
							HTTPVersion: "HTTP/1.1",
							QueryString: []QueryString{{Name: "", Value: "test"}},
						},
						Response: Response{
							Status:      200,
							HTTPVersion: "HTTP/1.1",
							Content:     Content{Size: 100, MimeType: "text/html"},
						},
					},
				},
			},
		}
		err := ValidateHarFile(har)
		assert.Error(t, err)
	})
}

func TestValidateBrowser(t *testing.T) {
	t.Run("BrowserNameWithoutVersion", func(t *testing.T) {
		har := &Har{
			Log: Log{
				Version: HarSpecVersion12,
				Creator: Creator{Name: "Test", Version: "1.0"},
				Browser: Browser{Name: "Chrome", Version: ""},
				Entries: []Entries{
					{
						StartedDateTime: time.Now(),
						Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
						Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
					},
				},
			},
		}
		err := ValidateHarFile(har)
		assert.Error(t, err)
	})

	t.Run("BrowserWithVersion", func(t *testing.T) {
		har := &Har{
			Log: Log{
				Version: HarSpecVersion12,
				Creator: Creator{Name: "Test", Version: "1.0"},
				Browser: Browser{Name: "Chrome", Version: "100.0"},
				Entries: []Entries{
					{
						StartedDateTime: time.Now(),
						Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
						Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
					},
				},
			},
		}
		err := ValidateHarFile(har)
		assert.NoError(t, err)
	})
}

func TestValidateContentEncoding(t *testing.T) {
	t.Run("InvalidEncoding", func(t *testing.T) {
		har := &Har{
			Log: Log{
				Version: HarSpecVersion12,
				Creator: Creator{Name: "Test", Version: "1.0"},
				Entries: []Entries{
					{
						StartedDateTime: time.Now(),
						Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
						Response: Response{
							Status:      200,
							HTTPVersion: "HTTP/1.1",
							Content:     Content{Size: 100, MimeType: "text/html", Encoding: "invalid"},
						},
					},
				},
			},
		}
		err := ValidateHarFile(har)
		assert.Error(t, err)
	})

	t.Run("Base64Encoding", func(t *testing.T) {
		har := &Har{
			Log: Log{
				Version: HarSpecVersion12,
				Creator: Creator{Name: "Test", Version: "1.0"},
				Entries: []Entries{
					{
						StartedDateTime: time.Now(),
						Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
						Response: Response{
							Status:      200,
							HTTPVersion: "HTTP/1.1",
							Content:     Content{Size: 100, MimeType: "text/html", Encoding: "base64"},
						},
					},
				},
			},
		}
		err := ValidateHarFile(har)
		assert.NoError(t, err)
	})
}

func TestValidateNegativeTime(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					Time:            -100,
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidatePostDataParamsV11(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion11,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "POST",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
						PostData: &PostData{
							MimeType: "application/x-www-form-urlencoded",
							Params:   []Param{{Name: "", Value: "test"}},
						},
					},
					Response: Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

// --- Comprehensive tests for uncovered branches ---

func TestValidateHarFile_Nil(t *testing.T) {
	err := ValidateHarFile(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HAR")
}

func TestValidateHarFile_NilEntries(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: nil,
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
	harErr, ok := err.(*HarError)
	require.True(t, ok)
	assert.True(t, harErr.HasPartialErrors())
}

func TestValidateHarFile_MissingCreatorVersion(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: ""},
			Entries: []Entries{},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
	harErr, ok := err.(*HarError)
	require.True(t, ok)
	found := false
	for _, pe := range harErr.GetPartialErrors() {
		if pe.Field == "log.creator.version" {
			found = true
		}
	}
	assert.True(t, found, "should have error for missing creator version")
}

func TestValidateHarFile_BrowserNameNoVersion(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Browser: Browser{Name: "Chrome", Version: ""},
			Entries: []Entries{},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
	harErr, ok := err.(*HarError)
	require.True(t, ok)
	found := false
	for _, pe := range harErr.GetPartialErrors() {
		if pe.Field == "log.browser.version" {
			found = true
		}
	}
	assert.True(t, found, "should have error for browser name without version")
}

func TestValidateHarFile_Version11(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion11,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
					},
					Response: Response{
						Status:      200,
						HTTPVersion: "HTTP/1.1",
						Content:     Content{Size: 100, MimeType: "text/html"},
					},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.NoError(t, err)
}

func TestValidateHarFile_Version13(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion13,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
					},
					Response: Response{
						Status:      200,
						HTTPVersion: "HTTP/1.1",
						Content:     Content{Size: 100, MimeType: "text/html"},
					},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.NoError(t, err)
}

func TestValidateHarFile_V11PostDataParamsWithParamsNil(t *testing.T) {
	// V1.1 with PostData but nil Params — should not error on params
	har := &Har{
		Log: Log{
			Version: HarSpecVersion11,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "POST",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
						PostData:    &PostData{MimeType: "text/plain", Text: "hello", Params: nil},
					},
					Response: Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.NoError(t, err)
}

func TestValidateHarFile_V12QueryStringWithoutName(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
						QueryString: []QueryString{{Name: "", Value: "test"}},
					},
					Response: Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidateHarFile_V12PostDataMissingMimeType(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "POST",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
						PostData:    &PostData{MimeType: "", Text: "data"},
					},
					Response: Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidateHarFile_V12ContentInvalidEncoding(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response: Response{
						Status:      200,
						HTTPVersion: "HTTP/1.1",
						Content:     Content{Size: 100, MimeType: "text/html", Encoding: "gzip"},
					},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidateHarFile_V13InheritsV12(t *testing.T) {
	// V1.3 includes all V1.2 validation — test invalid encoding under v1.3
	har := &Har{
		Log: Log{
			Version: HarSpecVersion13,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "POST",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
						PostData:    &PostData{MimeType: "", Text: "data"},
					},
					Response: Response{
						Status:      200,
						HTTPVersion: "HTTP/1.1",
						Content:     Content{Size: 100, MimeType: "text/html", Encoding: "invalid"},
					},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidateRequest_MissingMethod(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						// Method missing
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
					},
					Response: Response{
						Status:      200,
						HTTPVersion: "HTTP/1.1",
						Content:     Content{Size: 100, MimeType: "text/html"},
					},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
	harErr, ok := err.(*HarError)
	require.True(t, ok)
	found := false
	for _, pe := range harErr.GetPartialErrors() {
		if pe.Field == "log.entries[0].request.method" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestValidateRequest_MissingURL(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "GET",
						URL:         "", // missing
						HTTPVersion: "HTTP/1.1",
					},
					Response: Response{
						Status:      200,
						HTTPVersion: "HTTP/1.1",
						Content:     Content{Size: 100, MimeType: "text/html"},
					},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
	harErr, ok := err.(*HarError)
	require.True(t, ok)
	found := false
	for _, pe := range harErr.GetPartialErrors() {
		if pe.Field == "log.entries[0].request.url" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestValidateRequest_MissingHTTPVersion(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com",
						HTTPVersion: "",
					},
					Response: Response{
						Status:      200,
						HTTPVersion: "HTTP/1.1",
						Content:     Content{Size: 100, MimeType: "text/html"},
					},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidateRequest_WithPostDataAndParamsMissingName(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "POST",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
						PostData: &PostData{
							MimeType: "application/json",
							Params:   []Param{{Name: "", Value: "v"}, {Name: "ok", Value: "v"}},
						},
					},
					Response: Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidateResponse_MissingHTTPVersion(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
					},
					Response: Response{
						Status:      200,
						HTTPVersion: "", // missing
						Content:     Content{Size: 100, MimeType: "text/html"},
					},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidateResponse_NegativeStatus(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
					},
					Response: Response{
						Status:      -1,
						HTTPVersion: "HTTP/1.1",
						Content:     Content{Size: 100, MimeType: "text/html"},
					},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidateResponse_ZeroStatus(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
					},
					Response: Response{
						Status:      0, // zero is also invalid (<=0)
						HTTPVersion: "HTTP/1.1",
						Content:     Content{Size: 100, MimeType: "text/html"},
					},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidateContent_MissingMimeType(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
					},
					Response: Response{
						Status:      200,
						HTTPVersion: "HTTP/1.1",
						Content:     Content{Size: 100, MimeType: ""}, // missing MIME type
					},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidateContent_NegativeSize(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
					},
					Response: Response{
						Status:      200,
						HTTPVersion: "HTTP/1.1",
						Content:     Content{Size: -1, MimeType: "text/html"},
					},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidateContent_InvalidEncoding(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
					},
					Response: Response{
						Status:      200,
						HTTPVersion: "HTTP/1.1",
						Content:     Content{Size: 100, MimeType: "text/html", Encoding: "unknown"},
					},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidateContent_Base64EncodingCaseInsensitive(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
					},
					Response: Response{
						Status:      200,
						HTTPVersion: "HTTP/1.1",
						Content:     Content{Size: 100, MimeType: "text/html", Encoding: "Base64"},
					},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.NoError(t, err)
}

func TestValidateHeaders_MissingName(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
						Headers:     []Headers{{Name: "", Value: "test"}},
					},
					Response: Response{
						Status:      200,
						HTTPVersion: "HTTP/1.1",
						Content:     Content{Size: 100, MimeType: "text/html"},
					},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidateHeaders_ValidHeaders(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
						Headers:     []Headers{{Name: "Content-Type", Value: "text/html"}},
					},
					Response: Response{
						Status:      200,
						HTTPVersion: "HTTP/1.1",
						Content:     Content{Size: 100, MimeType: "text/html"},
						Headers:     []Headers{{Name: "X-Custom", Value: "value"}},
					},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.NoError(t, err)
}

func TestValidateCookies_MissingName(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
						Cookies:     []Cookie{{Name: "", Value: "test"}},
					},
					Response: Response{
						Status:      200,
						HTTPVersion: "HTTP/1.1",
						Content:     Content{Size: 100, MimeType: "text/html"},
					},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidateCookies_ValidCookies(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
						Cookies:     []Cookie{{Name: "session", Value: "abc123"}},
					},
					Response: Response{
						Status:      200,
						HTTPVersion: "HTTP/1.1",
						Content:     Content{Size: 100, MimeType: "text/html"},
						Cookies:     []Cookie{{Name: "tracking", Value: "xyz"}},
					},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.NoError(t, err)
}

func TestValidatePages_MissingID(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Pages: []Pages{
				{
					ID:              "", // missing
					StartedDateTime: time.Now(),
					Title:           "Test Page",
				},
			},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidatePages_MissingStartedDateTime(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Pages: []Pages{
				{
					ID:              "page_1",
					StartedDateTime: time.Time{}, // zero
					Title:           "Test Page",
				},
			},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidatePages_MissingTitle(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Pages: []Pages{
				{
					ID:              "page_1",
					StartedDateTime: time.Now(),
					Title:           "", // missing
				},
			},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidatePages_EmptySlice(t *testing.T) {
	// Empty pages slice should not cause any errors
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Pages:   []Pages{},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.NoError(t, err)
}

func TestValidatePageTimings_OnContentLoadTooNegative(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Pages: []Pages{
				{
					ID:              "page_1",
					StartedDateTime: time.Now(),
					Title:           "Test Page",
					PageTimings:     PageTimings{OnContentLoad: -5.0}, // less than -1
				},
			},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidatePageTimings_OnLoadTooNegative(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Pages: []Pages{
				{
					ID:              "page_1",
					StartedDateTime: time.Now(),
					Title:           "Test Page",
					PageTimings:     PageTimings{OnLoad: -10.0}, // less than -1
				},
			},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidatePageTimings_OnContentLoadMinusOne(t *testing.T) {
	// -1 is a valid value (means unavailable)
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Pages: []Pages{
				{
					ID:              "page_1",
					StartedDateTime: time.Now(),
					Title:           "Test Page",
					PageTimings:     PageTimings{OnContentLoad: -1, OnLoad: -1},
				},
			},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.NoError(t, err)
}

func TestValidatePageTimings_ZeroValues(t *testing.T) {
	// Zero values should be valid
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Pages: []Pages{
				{
					ID:              "page_1",
					StartedDateTime: time.Now(),
					Title:           "Test Page",
					PageTimings:     PageTimings{OnContentLoad: 0, OnLoad: 0},
				},
			},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.NoError(t, err)
}

func TestValidateTimings_NegativeWait(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
					Timings:         Timings{Wait: -10},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidateTimings_NegativeReceive(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
					Timings:         Timings{Receive: -5},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidateTimings_NegativeSend(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
					Timings:         Timings{Send: -1},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidateTimings_ValidTimings(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
					Timings:         Timings{Blocked: 10, DNS: 5, Connect: 20, Ssl: 15, Send: 5, Wait: 50, Receive: 30},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.NoError(t, err)
}

func TestValidateEntries_ZeroStartedDateTime(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Time{}, // zero
					Request:         Request{Method: "GET", URL: "https://example.com", HTTPVersion: "HTTP/1.1"},
					Response:        Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestDetectHarVersion_Nil(t *testing.T) {
	detected := DetectHarVersion(nil)
	assert.Equal(t, HarSpecVersion12, detected)
}

func TestDetectHarVersion_EmptyVersion(t *testing.T) {
	har := &Har{Log: Log{Version: ""}}
	detected := DetectHarVersion(har)
	assert.Equal(t, HarSpecVersion12, detected)
}

func TestDetectHarVersion_WhitespaceVersion(t *testing.T) {
	har := &Har{Log: Log{Version: " 1.2 "}}
	detected := DetectHarVersion(har)
	assert.Equal(t, HarSpecVersion12, detected)
}

func TestIsValidHarVersion(t *testing.T) {
	assert.True(t, IsValidHarVersion(HarSpecVersion11))
	assert.True(t, IsValidHarVersion(HarSpecVersion12))
	assert.True(t, IsValidHarVersion(HarSpecVersion13))
	assert.False(t, IsValidHarVersion("2.0"))
	assert.False(t, IsValidHarVersion(""))
	assert.False(t, IsValidHarVersion("0.9"))
}

func TestValidateQueryString_MissingName(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
						QueryString: []QueryString{{Name: "", Value: "test"}},
					},
					Response: Response{Status: 200, HTTPVersion: "HTTP/1.1", Content: Content{Size: 100, MimeType: "text/html"}},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidateResponse_HeadersMissingName(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
					},
					Response: Response{
						Status:      200,
						HTTPVersion: "HTTP/1.1",
						Content:     Content{Size: 100, MimeType: "text/html"},
						Headers:     []Headers{{Name: "", Value: "test"}},
					},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidateResponse_CookiesMissingName(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
					},
					Response: Response{
						Status:      200,
						HTTPVersion: "HTTP/1.1",
						Content:     Content{Size: 100, MimeType: "text/html"},
						Cookies:     []Cookie{{Name: "", Value: "test"}},
					},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
}

func TestValidateMultipleEntriesWithMultipleErrors(t *testing.T) {
	har := &Har{
		Log: Log{
			Version: HarSpecVersion12,
			Creator: Creator{Name: "Test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
						Cookies:     []Cookie{{Name: "", Value: "v"}},
						Headers:     []Headers{{Name: "", Value: "v"}},
					},
					Response: Response{
						Status:      200,
						HTTPVersion: "HTTP/1.1",
						Content:     Content{Size: -5, MimeType: "", Encoding: "gzip"},
						Cookies:     []Cookie{{Name: "", Value: "v"}},
						Headers:     []Headers{{Name: "", Value: "v"}},
					},
					Timings: Timings{Send: -1, Wait: -1, Receive: -1},
				},
			},
		},
	}
	err := ValidateHarFile(har)
	assert.Error(t, err)
	harErr, ok := err.(*HarError)
	require.True(t, ok)
	// Should have many partial errors
	assert.GreaterOrEqual(t, len(harErr.GetPartialErrors()), 5)
}
