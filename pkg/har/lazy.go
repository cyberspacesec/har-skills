package har

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// LazyContent 延迟加载的内容
type LazyContent struct {
	// 基本信息总是加载
	Size        int    `json:"size"`
	MimeType    string `json:"mimeType"`
	Compression int    `json:"compression,omitempty"`

	// 实际内容延迟加载
	Text     *string `json:"text,omitempty"`
	Encoding *string `json:"encoding,omitempty"`
	Comment  string  `json:"comment,omitempty"`

	// 用于延迟加载的原始数据
	rawData   json.RawMessage `json:"-"`
	loaded    bool            `json:"-"`
	loadMutex sync.RWMutex    `json:"-"`
}

// LazyResponse 带有延迟加载内容的响应
type LazyResponse struct {
	Status       int          `json:"status"`
	StatusText   string       `json:"statusText"`
	HTTPVersion  string       `json:"httpVersion"`
	Cookies      []Cookie     `json:"cookies"`
	Headers      []Headers    `json:"headers"`
	RedirectURL  string       `json:"redirectURL"`
	HeadersSize  int          `json:"headersSize"`
	BodySize     int          `json:"bodySize"`
	Content      *LazyContent `json:"content"`
	TransferSize int          `json:"_transferSize"`
	Error        any          `json:"_error"`
	Comment      string       `json:"comment,omitempty"`
}

// LazyEntries 带有延迟加载内容的条目
type LazyEntries struct {
	StartedDateTime time.Time    `json:"startedDateTime"`
	Time            float64      `json:"time"`
	Request         Request      `json:"request"`
	Response        LazyResponse `json:"response"`
	Cache           Cache        `json:"cache"`
	Timings         Timings      `json:"timings"`
	Pageref         string       `json:"pageref"`
	Initiator       Initiator    `json:"_initiator"`
	Priority        string       `json:"_priority"`
	ResourceType    string       `json:"_resourceType"`
	Connection      string       `json:"connection"`
	ServerIPAddress string       `json:"serverIPAddress"`
	Comment         string       `json:"comment,omitempty"`
}

// LazyHar 带有延迟加载功能的HAR对象
type LazyHar struct {
	Log struct {
		Version string        `json:"version"`
		Creator Creator       `json:"creator"`
		Browser Browser       `json:"browser,omitempty"`
		Pages   []Pages       `json:"pages"`
		Entries []LazyEntries `json:"entries"`
	} `json:"log"`
}

// UnmarshalJSON 自定义JSON解析，初始时只解析基本信息
func (lc *LazyContent) UnmarshalJSON(data []byte) error {
	// 保存原始数据用于延迟加载
	lc.rawData = make(json.RawMessage, len(data))
	copy(lc.rawData, data)

	// 解析基本信息
	type BasicContent struct {
		Size        int    `json:"size"`
		MimeType    string `json:"mimeType"`
		Compression int    `json:"compression"`
		Comment     string `json:"comment"`
	}

	var basic BasicContent
	if err := json.Unmarshal(data, &basic); err != nil {
		return err
	}

	lc.Size = basic.Size
	lc.MimeType = basic.MimeType
	lc.Compression = basic.Compression
	lc.Comment = basic.Comment
	lc.loaded = false

	return nil
}

// Load 加载完整的内容数据
func (lc *LazyContent) Load() error {
	lc.loadMutex.Lock()
	defer lc.loadMutex.Unlock()

	if lc.loaded {
		return nil
	}

	// 临时结构体，用于解析完整内容
	type FullContent struct {
		Text     *string `json:"text,omitempty"`
		Encoding *string `json:"encoding,omitempty"`
	}

	var full FullContent
	if err := json.Unmarshal(lc.rawData, &full); err != nil {
		return err
	}

	lc.Text = full.Text
	lc.Encoding = full.Encoding
	lc.loaded = true

	return nil
}

// GetText 获取内容文本，如果尚未加载则先加载
func (lc *LazyContent) GetText() (*string, error) {
	lc.loadMutex.RLock()
	if lc.loaded {
		text := lc.Text
		lc.loadMutex.RUnlock()
		return text, nil
	}
	lc.loadMutex.RUnlock()

	if err := lc.Load(); err != nil {
		return nil, err
	}

	return lc.Text, nil
}

// ParseHarWithLazyLoading 解析HAR内容，对大型字段使用延迟加载
func ParseHarWithLazyLoading(harFileBytes []byte) (*LazyHar, error) {
	har := new(LazyHar)
	err := json.Unmarshal(harFileBytes, har)
	if err != nil {
		return nil, WrapJSONUnmarshalError(err)
	}
	return har, nil
}

// ParseHarFileWithLazyLoading 解析HAR文件，对大型字段使用延迟加载
func ParseHarFileWithLazyLoading(harFilePath string) (*LazyHar, error) {
	harFileBytes, err := os.ReadFile(harFilePath)
	if err != nil {
		return nil, NewFileSystemError("无法读取HAR文件", err)
	}
	return ParseHarWithLazyLoading(harFileBytes)
}

// ToStandardHar 将LazyHar转换为标准Har对象
func (lh *LazyHar) ToStandardHar() (*Har, error) {
	// 创建标准HAR对象
	result := &Har{
		Log: Log{
			Version: lh.Log.Version,
			Creator: lh.Log.Creator,
			Browser: lh.Log.Browser,
			Pages:   lh.Log.Pages,
			Entries: make([]Entries, len(lh.Log.Entries)),
		},
	}

	// 转换entries
	for i, lazyEntry := range lh.Log.Entries {
		// 复制基本字段
		entry := Entries{
			StartedDateTime: lazyEntry.StartedDateTime,
			Time:            lazyEntry.Time,
			Request:         lazyEntry.Request,
			Cache:           lazyEntry.Cache,
			Timings:         lazyEntry.Timings,
			Pageref:         lazyEntry.Pageref,
			Initiator:       lazyEntry.Initiator,
			Priority:        lazyEntry.Priority,
			ResourceType:    lazyEntry.ResourceType,
			Connection:      lazyEntry.Connection,
			ServerIPAddress: lazyEntry.ServerIPAddress,
				Comment:         lazyEntry.Comment,
			}

		// 复制响应字段
		entry.Response = Response{
			Status:       lazyEntry.Response.Status,
			StatusText:   lazyEntry.Response.StatusText,
			HTTPVersion:  lazyEntry.Response.HTTPVersion,
			Cookies:      lazyEntry.Response.Cookies,
			Headers:      lazyEntry.Response.Headers,
			RedirectURL:  lazyEntry.Response.RedirectURL,
			HeadersSize:  lazyEntry.Response.HeadersSize,
			BodySize:     lazyEntry.Response.BodySize,
			TransferSize: lazyEntry.Response.TransferSize,
			Error:        lazyEntry.Response.Error,
				Comment:      lazyEntry.Response.Comment,
		}

		// 复制内容
		if lazyEntry.Response.Content != nil {
			// 确保内容已加载
			if err := lazyEntry.Response.Content.Load(); err != nil {
				return nil, NewJSONParseError("无法加载延迟加载的内容", err)
			}

			entry.Response.Content = Content{
				Size:     lazyEntry.Response.Content.Size,
				MimeType: lazyEntry.Response.Content.MimeType,
			}
		}

		result.Log.Entries[i] = entry
	}

	return result, nil
}

// GetEntry 获取指定索引的条目
func (lh *LazyHar) GetEntry(index int) (*LazyEntries, error) {
	if index < 0 || index >= len(lh.Log.Entries) {
		return nil, NewInvalidValueError("index", index, "索引超出范围")
	}
	return &lh.Log.Entries[index], nil
}

// GetEntriesCount 获取条目数量
func (lh *LazyHar) GetEntriesCount() int {
	return len(lh.Log.Entries)
}

// GetResponseContent 获取指定索引条目的响应内容
func (lh *LazyHar) GetResponseContent(index int) (*LazyContent, error) {
	entry, err := lh.GetEntry(index)
	if err != nil {
		return nil, err
	}
	return entry.Response.Content, nil
}

// GetResponseText 获取指定索引条目的响应文本
func (lh *LazyHar) GetResponseText(index int) (*string, error) {
	content, err := lh.GetResponseContent(index)
	if err != nil {
		return nil, err
	}
	if content == nil {
		return nil, nil
	}
	return content.GetText()
}
