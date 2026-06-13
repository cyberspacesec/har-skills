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
