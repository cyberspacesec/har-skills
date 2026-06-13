package har

import (
	"encoding/json"
	"os"
	"time"
)

// NewHar 创建一个新的HAR对象
func NewHar() *Har {
	return &Har{
		Log: Log{
			Version: "1.2",
			Creator: Creator{
				Name:    "go-har",
				Version: "1.0",
			},
			Pages:   []Pages{},
			Entries: []Entries{},
		},
	}
}

// SetBrowser 设置HAR文件的浏览器信息
func (h *Har) SetBrowser(name, version string) *Har {
	h.Log.Browser = Browser{
		Name:    name,
		Version: version,
	}
	return h
}

// SetVersion 设置HAR规范版本
func (h *Har) SetVersion(version string) *Har {
	h.Log.Version = version
	return h
}

// SetCreator 设置HAR文件的创建者信息
func (h *Har) SetCreator(name, version string) *Har {
	h.Log.Creator.Name = name
	h.Log.Creator.Version = version
	return h
}

// AddPage 添加页面信息
func (h *Har) AddPage(id, title string) *Pages {
	page := Pages{
		StartedDateTime: time.Now(),
		ID:              id,
		Title:           title,
		PageTimings: PageTimings{
			OnContentLoad: -1,
			OnLoad:        -1,
		},
	}
	h.Log.Pages = append(h.Log.Pages, page)
	return &h.Log.Pages[len(h.Log.Pages)-1]
}

// SetPageTimings 设置页面加载时间
func (p *Pages) SetPageTimings(onContentLoad, onLoad float64) *Pages {
	p.PageTimings.OnContentLoad = onContentLoad
	p.PageTimings.OnLoad = onLoad
	return p
}

// AddEntry 添加一个请求/响应条目
func (h *Har) AddEntry(method, url, httpVersion string, pageref string) *Entries {
	entry := Entries{
		StartedDateTime: time.Now(),
		Time:            0,
		Request: Request{
			Method:      method,
			URL:         url,
			HTTPVersion: httpVersion,
			Headers:     []Headers{},
			Cookies:     []Cookie{},
			QueryString: []QueryString{},
			HeadersSize: -1,
			BodySize:    -1,
		},
		Response: Response{
			Status:      0,
			StatusText:  "",
			HTTPVersion: httpVersion,
			Headers:     []Headers{},
			Cookies:     []Cookie{},
			Content: Content{
				Size:     0,
				MimeType: "",
			},
			RedirectURL:  "",
			HeadersSize:  -1,
			BodySize:     -1,
			TransferSize: -1,
		},
		Cache: Cache{},
		Timings: Timings{
			Blocked: -1,
			DNS:     -1,
			Connect: -1,
			Send:    -1,
			Wait:    -1,
			Receive: -1,
			Ssl:     -1,
		},
		Pageref: pageref,
	}
	h.Log.Entries = append(h.Log.Entries, entry)
	return &h.Log.Entries[len(h.Log.Entries)-1]
}

// AddRequestHeader 添加请求头
func (e *Entries) AddRequestHeader(name, value string) *Entries {
	e.Request.Headers = append(e.Request.Headers, Headers{
		Name:  name,
		Value: value,
	})
	return e
}

// AddResponseHeader 添加响应头
func (e *Entries) AddResponseHeader(name, value string) *Entries {
	e.Response.Headers = append(e.Response.Headers, Headers{
		Name:  name,
		Value: value,
	})
	return e
}

// SetResponseStatus 设置响应状态
func (e *Entries) SetResponseStatus(status int, statusText string) *Entries {
	e.Response.Status = status
	e.Response.StatusText = statusText
	return e
}

// SetResponseContent 设置响应内容
func (e *Entries) SetResponseContent(size int, mimeType string) *Entries {
	e.Response.Content.Size = size
	e.Response.Content.MimeType = mimeType
	return e
}

// SetTimings 设置时间数据
func (e *Entries) SetTimings(blocked, dns, connect, send, wait, receive, ssl float64) *Entries {
	e.Timings.Blocked = blocked
	e.Timings.DNS = dns
	e.Timings.Connect = connect
	e.Timings.Send = send
	e.Timings.Wait = wait
	e.Timings.Receive = receive
	e.Timings.Ssl = ssl

	// 计算总时间
	// 注意：根据HAR规范，SSL时间包含在connect时间内，不重复计算
	e.Time = blocked + dns + connect + send + wait + receive
	return e
}

// AddCookie 添加请求Cookie
func (e *Entries) AddCookie(name, value string) *Entries {
	e.Request.Cookies = append(e.Request.Cookies, Cookie{
		Name:  name,
		Value: value,
	})
	return e
}

// AddResponseCookie 添加响应Cookie
func (e *Entries) AddResponseCookie(name, value string) *Entries {
	e.Response.Cookies = append(e.Response.Cookies, Cookie{
		Name:  name,
		Value: value,
	})
	return e
}

// AddQueryParameter 添加查询参数
func (e *Entries) AddQueryParameter(name, value string) *Entries {
	e.Request.QueryString = append(e.Request.QueryString, QueryString{
		Name:  name,
		Value: value,
	})
	return e
}

// SetPostData 设置POST请求体
func (e *Entries) SetPostData(mimeType, text string) *Entries {
	e.Request.PostData = &PostData{
		MimeType: mimeType,
		Text:     text,
	}
	return e
}

// SetPostDataParams 设置POST表单参数
func (e *Entries) SetPostDataParams(mimeType string, params []Param) *Entries {
	e.Request.PostData = &PostData{
		MimeType: mimeType,
		Params:   params,
	}
	return e
}

// SetResponseContentText 设置响应内容文本
func (e *Entries) SetResponseContentText(text string) *Entries {
	e.Response.Content.Text = text
	e.Response.Content.Size = len(text)
	return e
}

// SetServerIP 设置服务器IP地址
func (e *Entries) SetServerIP(ip string) *Entries {
	e.ServerIPAddress = ip
	return e
}

// SetConnection 设置连接ID
func (e *Entries) SetConnection(id string) *Entries {
	e.Connection = id
	return e
}

// SetPageref 设置页面引用
func (e *Entries) SetPageref(ref string) *Entries {
	e.Pageref = ref
	return e
}

// ToJSON 将HAR对象转换为JSON字节
func (h *Har) ToJSON(indent bool) ([]byte, error) {
	if indent {
		return json.MarshalIndent(h, "", "  ")
	}
	return json.Marshal(h)
}

// SaveToFile 将HAR对象保存到文件
func (h *Har) SaveToFile(filePath string, indent bool) error {
	data, err := h.ToJSON(indent)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}
