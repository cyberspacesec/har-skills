package har

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 创建测试用HAR文件
func setupTestFiles(t *testing.T) {
	// 创建测试目录
	testDataDir := "testdata"
	_ = os.MkdirAll(testDataDir, 0755)

	// 有效的HAR文件 - 最小配置
	minimalHar := Har{
		Log: Log{
			Version: "1.2",
			Creator: Creator{
				Name:    "Go-HAR Test",
				Version: "1.0",
			},
			Entries: []Entries{},
		},
	}
	writeHarFile(t, filepath.Join(testDataDir, "minimal.har"), minimalHar)

	// 有效的HAR文件 - 完整配置
	fullHar := createFullHar()
	writeHarFile(t, filepath.Join(testDataDir, "full.har"), fullHar)

	// 有效的HAR文件 - 1.1版本
	har11 := Har{
		Log: Log{
			Version: "1.1",
			Creator: Creator{
				Name:    "Go-HAR Test",
				Version: "1.0",
			},
			Entries: []Entries{},
		},
	}
	writeHarFile(t, filepath.Join(testDataDir, "v1.1.har"), har11)

	// 无效的HAR文件 - 缺少必要字段
	invalidHar := map[string]interface{}{
		"log": map[string]interface{}{
			// 缺少 version
			"creator": map[string]interface{}{
				// 缺少 name
				"version": "1.0",
			},
			"entries": []interface{}{},
		},
	}
	writeJSONFile(t, filepath.Join(testDataDir, "invalid.har"), invalidHar)

	// 无效的HAR文件 - 不是JSON格式
	writeTextFile(t, filepath.Join(testDataDir, "not_json.har"), "This is not a JSON file")

	// 无效的HAR文件 - 错误的日期格式
	invalidDateHar := Har{
		Log: Log{
			Version: "1.2",
			Creator: Creator{
				Name:    "Go-HAR Test",
				Version: "1.0",
			},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(), // 有效日期
					Request: Request{
						Method: "GET",
						URL:    "https://example.com",
					},
					Response: Response{
						Status: 200,
					},
				},
			},
		},
	}
	writeHarFile(t, filepath.Join(testDataDir, "invalid_date.har"), invalidDateHar)

	// 大型HAR文件 - 用于性能测试
	largeHar := createLargeHar(1000) // 1000个条目
	writeHarFile(t, filepath.Join(testDataDir, "large.har"), largeHar)
}

// 辅助函数：写入HAR文件
func writeHarFile(t *testing.T, filename string, har Har) {
	data, err := json.MarshalIndent(har, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(filename, data, 0644)
	require.NoError(t, err)
}

// 辅助函数：写入任意JSON文件
func writeJSONFile(t *testing.T, filename string, data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(filename, jsonData, 0644)
	require.NoError(t, err)
}

// 辅助函数：写入文本文件
func writeTextFile(t *testing.T, filename string, content string) {
	err := os.WriteFile(filename, []byte(content), 0644)
	require.NoError(t, err)
}

// 辅助函数：创建完整的HAR对象
func createFullHar() Har {
	now := time.Now()
	return Har{
		Log: Log{
			Version: "1.2",
			Creator: Creator{
				Name:    "Go-HAR Test",
				Version: "1.0",
			},
			Pages: []Pages{
				{
					StartedDateTime: now,
					ID:              "page_1",
					Title:           "Test Page",
					PageTimings: PageTimings{
						OnContentLoad: 150.5,
						OnLoad:        250.75,
						Comment:       "Page timing comment",
					},
				},
			},
			Entries: []Entries{
				{
					Pageref:         "page_1",
					StartedDateTime: now,
					Time:            350.25,
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com/test",
						HTTPVersion: "HTTP/1.1",
						Headers: []Headers{
							{Name: "Accept", Value: "application/json"},
							{Name: "User-Agent", Value: "Go-HAR Test"},
						},
						QueryString: []QueryString{
							{Name: "id", Value: "12345"},
							{Name: "format", Value: "json"},
						},
						Cookies: []Cookie{
							{
								Name:     "session",
								Value:    "abc123",
								Path:     "/",
								Domain:   "example.com",
								Expires:  now.Add(24 * time.Hour),
								HTTPOnly: true,
								Secure:   true,
							},
						},
						HeadersSize: 150,
						BodySize:    0,
					},
					Response: Response{
						Status:      200,
						StatusText:  "OK",
						HTTPVersion: "HTTP/1.1",
						Headers: []Headers{
							{Name: "Content-Type", Value: "application/json"},
							{Name: "Cache-Control", Value: "no-cache"},
						},
						Cookies: []Cookie{},
						Content: Content{
							Size:     1024,
							MimeType: "application/json",
						},
						RedirectURL:  "",
						HeadersSize:  120,
						BodySize:     1024,
						TransferSize: 1144,
					},
					Cache: Cache{},
					Timings: Timings{
						Blocked: 12.5,
						DNS:     10.0,
						Connect: 25.5,
						Send:    5.5,
						Wait:    75.25,
						Receive: 15.75,
						Ssl:     20.0,
					},
					ServerIPAddress: "192.168.1.1",
					Connection:      "close",
				},
			},
		},
	}
}

// 辅助函数：创建大型HAR文件
func createLargeHar(entriesCount int) Har {
	now := time.Now()
	har := Har{
		Log: Log{
			Version: "1.2",
			Creator: Creator{
				Name:    "Go-HAR Test",
				Version: "1.0",
			},
			Pages:   []Pages{},
			Entries: make([]Entries, entriesCount),
		},
	}

	for i := 0; i < entriesCount; i++ {
		har.Log.Entries[i] = Entries{
			StartedDateTime: now.Add(time.Duration(i) * time.Second),
			Time:            float64(100 + i),
			Request: Request{
				Method:      "GET",
				URL:         "https://example.com/api/item/" + string(rune(i)),
				HTTPVersion: "HTTP/1.1",
				Headers: []Headers{
					{Name: "Accept", Value: "application/json"},
				},
				HeadersSize: 100,
				BodySize:    0,
			},
			Response: Response{
				Status:      200,
				StatusText:  "OK",
				HTTPVersion: "HTTP/1.1",
				Headers: []Headers{
					{Name: "Content-Type", Value: "application/json"},
				},
				Content: Content{
					Size:     500,
					MimeType: "application/json",
				},
				HeadersSize: 80,
				BodySize:    500,
			},
			Timings: Timings{
				Blocked: 10.0,
				DNS:     5.0,
				Connect: 15.0,
				Send:    5.0,
				Wait:    50.0,
				Receive: 15.0,
			},
		}
	}

	return har
}

// TestParseHarBasic 测试基本解析功能
func TestParseHarBasic(t *testing.T) {
	setupTestFiles(t)

	// 测试解析最小HAR文件
	t.Run("ParseMinimalHar", func(t *testing.T) {
		data, err := os.ReadFile("testdata/minimal.har")
		require.NoError(t, err)

		har, err := ParseHar(data)
		require.NoError(t, err)
		assert.Equal(t, "1.2", har.Log.Version)
		assert.Equal(t, "Go-HAR Test", har.Log.Creator.Name)
		assert.Empty(t, har.Log.Entries)
	})

	// 测试解析完整HAR文件
	t.Run("ParseFullHar", func(t *testing.T) {
		data, err := os.ReadFile("testdata/full.har")
		require.NoError(t, err)

		har, err := ParseHar(data)
		require.NoError(t, err)
		assert.Equal(t, "1.2", har.Log.Version)
		assert.Equal(t, "Go-HAR Test", har.Log.Creator.Name)
		assert.Equal(t, "Test Page", har.Log.Pages[0].Title)
		assert.Len(t, har.Log.Pages, 1)
		assert.Len(t, har.Log.Entries, 1)

		// 验证entry详情
		entry := har.Log.Entries[0]
		assert.Equal(t, "page_1", entry.Pageref)
		assert.Equal(t, "GET", entry.Request.Method)
		assert.Equal(t, "https://example.com/test", entry.Request.URL)
		assert.Equal(t, 200, entry.Response.Status)
		assert.Equal(t, "application/json", entry.Response.Content.MimeType)
	})

	// 测试解析无效HAR文件
	t.Run("ParseInvalidHar", func(t *testing.T) {
		data, err := os.ReadFile("testdata/invalid.har")
		require.NoError(t, err)

		har, err := ParseHar(data)
		assert.Error(t, err)
		assert.Nil(t, har)

		// 验证错误类型
		harErr, ok := err.(*HarError)
		if assert.True(t, ok, "Expected HarError type") {
			assert.Equal(t, ErrCodeValidation, harErr.Code)
			assert.True(t, harErr.HasPartialErrors())
		}
	})

	// 测试非JSON文件
	t.Run("ParseNonJsonFile", func(t *testing.T) {
		data, err := os.ReadFile("testdata/not_json.har")
		require.NoError(t, err)

		har, err := ParseHar(data)
		assert.Error(t, err)
		assert.Nil(t, har)
	})
}

// 测试不同版本的HAR规范
func TestHarVersions(t *testing.T) {
	t.Run("ParseHarV1.1", func(t *testing.T) {
		data, err := os.ReadFile("testdata/version_11.har")
		require.NoError(t, err)

		har, err := ParseHar(data)
		require.NoError(t, err)
		assert.Equal(t, "1.1", har.Log.Version)
	})

	t.Run("ParseHarV1.3", func(t *testing.T) {
		data, err := os.ReadFile("testdata/version_13.har")
		require.NoError(t, err)

		har, err := ParseHar(data)
		require.NoError(t, err)
		assert.Equal(t, "1.3", har.Log.Version)
	})
}

// 测试内存优化模式
func TestMemoryOptimizedParsing(t *testing.T) {
	t.Run("ParseWithMemoryOptimized", func(t *testing.T) {
		data, err := os.ReadFile("testdata/full.har")
		require.NoError(t, err)

		har, err := Parse(data, WithMemoryOptimized())
		require.NoError(t, err)
		assert.IsType(t, &OptimizedHar{}, har)
		assert.Equal(t, "1.2", har.GetVersion())
		assert.Equal(t, "Go-HAR Test", har.GetCreator().Name)
		assert.Len(t, har.GetEntries(), 1)
	})
}

// 测试懒加载模式
func TestLazyLoadingParsing(t *testing.T) {
	t.Run("ParseWithLazyLoading", func(t *testing.T) {
		data, err := os.ReadFile("testdata/full.har")
		require.NoError(t, err)

		har, err := Parse(data, WithLazyLoading())
		require.NoError(t, err)
		assert.Equal(t, "1.2", har.GetVersion())

		// 验证懒加载
		entries := har.GetEntries()
		assert.Len(t, entries, 1)

		content := entries[0].GetResponse().GetContent()
		assert.Equal(t, 1024, content.GetSize())
		assert.Equal(t, "application/json", content.GetMimeType())
	})
}

// 测试跳过验证
func TestSkipValidation(t *testing.T) {
	t.Run("ParseInvalidHarWithSkipValidation", func(t *testing.T) {
		data, err := os.ReadFile("testdata/invalid.har")
		require.NoError(t, err)

		// 使用跳过验证选项，应该能解析成功
		har, err := Parse(data, WithSkipValidation())
		assert.NoError(t, err)
		assert.NotNil(t, har)
	})
}

// 测试流式解析
func TestStreamingParsing(t *testing.T) {
	t.Run("EmptyEntries", func(t *testing.T) {
		data, err := os.ReadFile("testdata/minimal.har")
		require.NoError(t, err)

		parser, err := NewStreamingParser(data)
		require.NoError(t, err)

		count := 0
		for parser.Next() {
			count++
		}
		assert.NoError(t, parser.Err())
		assert.Equal(t, 0, count) // minimal.har doesn't have any entries
	})

	t.Run("WithEntries", func(t *testing.T) {
		data, err := os.ReadFile("testdata/example.har")
		require.NoError(t, err)

		parser, err := NewStreamingParser(data)
		require.NoError(t, err)

		count := 0
		for parser.Next() {
			entry := parser.Entry()
			assert.NotNil(t, entry)

			// 验证条目内容
			if count == 0 {
				assert.Equal(t, "GET", entry.Request.Method)
				assert.Equal(t, "https://example.com/test", entry.Request.URL)
				assert.Equal(t, 200, entry.Response.Status)
				assert.Equal(t, "text/plain", entry.Response.Content.MimeType)
				assert.Equal(t, 100, entry.Response.Content.Size)
			}
			count++
		}
		assert.NoError(t, parser.Err())
		assert.Equal(t, 1, count) // example.har has one entry
	})
}

// 测试宽松模式
func TestLenientParsing(t *testing.T) {
	t.Run("ParseWithLenient", func(t *testing.T) {
		data, err := os.ReadFile("testdata/invalid.har")
		require.NoError(t, err)

		// 使用宽松解析选项
		har, err := Parse(data, WithLenient())
		// 应该解析成功，但会有警告
		assert.NoError(t, err)
		assert.NotNil(t, har)
	})
}

// 测试增强的错误处理
func TestEnhancedErrorHandling(t *testing.T) {
	t.Run("ParseWithWarnings", func(t *testing.T) {
		data, err := os.ReadFile("testdata/invalid_url.har")
		require.NoError(t, err)

		// 使用警告收集
		result, err := ParseHarWithWarnings(data)
		assert.NoError(t, err)
		assert.NotNil(t, result.Har)

		// 由于invalid_url.har包含无效的URL，应该会产生警告
		// 但在宽松模式下这些警告不会导致解析失败
		assert.NotEmpty(t, result.Warnings)

		// 验证警告中包含URL相关的错误
		found := false
		for _, warning := range result.Warnings {
			if strings.Contains(warning.Field, "url") || strings.Contains(warning.Message, "URL") {
				found = true
				break
			}
		}
		assert.True(t, found, "警告中应包含URL相关的错误")
	})
}

// 测试HAR转换
func TestHarConversion(t *testing.T) {
	t.Run("StandardToOptimized", func(t *testing.T) {
		data, err := os.ReadFile("testdata/full.har")
		require.NoError(t, err)

		standard, err := ParseHar(data)
		require.NoError(t, err)

		optimized := ToOptimizedHar(standard)
		assert.Equal(t, standard.Log.Version, optimized.Log.Version)
		assert.Equal(t, standard.Log.Creator.Name, optimized.Log.Creator.Name)
		assert.Len(t, optimized.Log.Entries, len(standard.Log.Entries))
	})

	t.Run("OptimizedToStandard", func(t *testing.T) {
		data, err := os.ReadFile("testdata/full.har")
		require.NoError(t, err)

		optimized, err := Parse(data, WithMemoryOptimized())
		require.NoError(t, err)

		standard := optimized.ToStandard()
		assert.Equal(t, "1.2", standard.Log.Version)
		assert.Equal(t, "Go-HAR Test", standard.Log.Creator.Name)
		assert.Len(t, standard.Log.Entries, 1)
	})
}
