package har

import (
	"fmt"
	"net/url"
	"strings"
)

// HAR规范版本常量
const (
	// HAR规范1.1版本
	HarSpecVersion11 = "1.1"
	// HAR规范1.2版本
	HarSpecVersion12 = "1.2"
	// HAR规范1.3版本 (非官方，但一些工具使用)
	HarSpecVersion13 = "1.3"
)

// ValidateHarFile 验证HAR对象内容的有效性
// 支持不同版本的HAR规范
func ValidateHarFile(har *Har) error {
	if har == nil {
		return NewInvalidFormatError("HAR对象为空")
	}

	// 创建根错误
	rootError := &HarError{
		Code:    ErrCodeValidation,
		Message: "HAR验证失败",
	}

	// 验证基本结构
	if err := validateBasicStructure(har, rootError); err != nil {
		return err
	}

	// 根据版本进行特定验证
	switch har.Log.Version {
	case HarSpecVersion11:
		validateHarV11(har, rootError)
	case HarSpecVersion12:
		validateHarV12(har, rootError)
	case HarSpecVersion13:
		validateHarV13(har, rootError)
	default:
		_ = rootError.AddPartialError(NewValidationError(
			fmt.Sprintf("不支持的HAR版本: %s", har.Log.Version),
			"log.version",
		))
	}

	// 验证条目的通用部分
	validateEntries(har.Log.Entries, rootError)

	// 验证页面
	validatePages(har.Log.Pages, rootError)

	// 有部分错误时返回
	if rootError.HasPartialErrors() {
		return rootError
	}

	return nil
}

// validateBasicStructure 验证HAR的基本结构
func validateBasicStructure(har *Har, rootError *HarError) error {
	// 验证Log字段
	if har.Log.Version == "" {
		_ = rootError.AddPartialError(NewMissingFieldError("log.version"))
	}

	// 验证Creator字段
	if har.Log.Creator.Name == "" {
		_ = rootError.AddPartialError(NewMissingFieldError("log.creator.name"))
	}

	if har.Log.Creator.Version == "" {
		_ = rootError.AddPartialError(NewMissingFieldError("log.creator.version"))
	}

	// 验证Browser字段（如果存在）
	if har.Log.Browser.Name != "" && har.Log.Browser.Version == "" {
		_ = rootError.AddPartialError(NewValidationError(
			"浏览器名称存在但版本为空",
			"log.browser.version",
		))
	}

	// 验证Entries数组
	// HAR文件可以没有条目，但必须有数组
	if har.Log.Entries == nil {
		_ = rootError.AddPartialError(NewMissingFieldError("log.entries"))
	}

	// 有部分错误时返回
	if rootError.HasPartialErrors() {
		return rootError
	}

	return nil
}

// validateHarV11 验证HAR 1.1版本的特定要求
func validateHarV11(har *Har, rootError *HarError) {
	// 1.1版本：PostData.params中每个param必须有name
	for i, entry := range har.Log.Entries {
		if entry.Request.PostData != nil && entry.Request.PostData.Params != nil {
			for j, param := range entry.Request.PostData.Params {
				if param.Name == "" {
					_ = rootError.AddPartialError(NewValidationError(
						"PostData参数必须有name字段",
						fmt.Sprintf("log.entries[%d].request.postData.params[%d].name", i, j),
					))
				}
			}
		}
	}
}

// validateHarV12 验证HAR 1.2版本的特定要求
func validateHarV12(har *Har, rootError *HarError) {
	// 1.2版本：验证QueryString必须有name
	for i, entry := range har.Log.Entries {
		for j, qs := range entry.Request.QueryString {
			if qs.Name == "" {
				_ = rootError.AddPartialError(NewValidationError(
					"QueryString参数必须有name字段",
					fmt.Sprintf("log.entries[%d].request.queryString[%d].name", i, j),
				))
			}
		}

		// 验证PostData（如果存在）
		if entry.Request.PostData != nil {
			if entry.Request.PostData.MimeType == "" {
				_ = rootError.AddPartialError(NewValidationError(
					"PostData必须有mimeType字段",
					fmt.Sprintf("log.entries[%d].request.postData.mimeType", i),
				))
			}
		}
	}

	// 验证Content.encoding（如果存在，必须是base64）
	for i, entry := range har.Log.Entries {
		if entry.Response.Content.Encoding != "" &&
			!strings.EqualFold(entry.Response.Content.Encoding, "base64") {
			_ = rootError.AddPartialError(NewValidationError(
				fmt.Sprintf("Content.encoding只支持base64，当前为: %s", entry.Response.Content.Encoding),
				fmt.Sprintf("log.entries[%d].response.content.encoding", i),
			))
		}
	}
}

// validateHarV13 验证HAR 1.3版本的特定要求
func validateHarV13(har *Har, rootError *HarError) {
	// 1.3版本特定验证
	// 非官方但有一些工具使用此版本
	// 包含1.2的所有验证
	validateHarV12(har, rootError)
}

// validateEntries 验证HAR条目
func validateEntries(entries []Entries, rootError *HarError) {
	for i, entry := range entries {
		entryPrefix := fmt.Sprintf("log.entries[%d]", i)

		// 验证必要的时间字段
		if entry.StartedDateTime.IsZero() {
			_ = rootError.AddPartialError(NewValidationError(
				"条目必须有开始时间",
				fmt.Sprintf("%s.startedDateTime", entryPrefix),
			))
		}

		// 验证时间值
		if entry.Time < 0 {
			_ = rootError.AddPartialError(NewValidationError(
				"条目时间不能为负",
				fmt.Sprintf("%s.time", entryPrefix),
			))
		}

		// 验证请求
		validateRequest(entry.Request, fmt.Sprintf("%s.request", entryPrefix), rootError)

		// 验证响应
		validateResponse(entry.Response, fmt.Sprintf("%s.response", entryPrefix), rootError)

		// 验证时间字段
		validateTimings(entry.Timings, fmt.Sprintf("%s.timings", entryPrefix), rootError)

	}
}

// validateRequest 验证HTTP请求
func validateRequest(req Request, fieldPath string, rootError *HarError) {
	// 验证方法
	if req.Method == "" {
		_ = rootError.AddPartialError(NewValidationError(
			"HTTP请求必须有方法",
			fmt.Sprintf("%s.method", fieldPath),
		))
	}

	// 验证URL
	if req.URL == "" {
		_ = rootError.AddPartialError(NewValidationError(
			"HTTP请求必须有URL",
			fmt.Sprintf("%s.url", fieldPath),
		))
	} else {
		// 验证URL格式
		_, err := url.Parse(req.URL)
		if err != nil {
			_ = rootError.AddPartialError(NewValidationError(
				fmt.Sprintf("无效的URL格式: %s", err.Error()),
				fmt.Sprintf("%s.url", fieldPath),
			))
		}
	}

	// 验证HTTP版本
	if req.HTTPVersion == "" {
		_ = rootError.AddPartialError(NewValidationError(
			"HTTP请求必须有版本",
			fmt.Sprintf("%s.httpVersion", fieldPath),
		))
	}

	// 验证headers
	validateHeaders(req.Headers, fmt.Sprintf("%s.headers", fieldPath), rootError)

	// 验证cookies
	validateCookies(req.Cookies, fmt.Sprintf("%s.cookies", fieldPath), rootError)

	// 验证QueryString
	validateQueryString(req.QueryString, fmt.Sprintf("%s.queryString", fieldPath), rootError)

	// 验证PostData（如果存在）
	if req.PostData != nil {
		validatePostData(req.PostData, fmt.Sprintf("%s.postData", fieldPath), rootError)
	}
}

// validateResponse 验证HTTP响应
func validateResponse(resp Response, fieldPath string, rootError *HarError) {
	// 验证状态码
	if resp.Status <= 0 {
		_ = rootError.AddPartialError(NewValidationError(
			"HTTP响应必须有有效的状态码",
			fmt.Sprintf("%s.status", fieldPath),
		))
	}

	// 验证HTTP版本
	if resp.HTTPVersion == "" {
		_ = rootError.AddPartialError(NewValidationError(
			"HTTP响应必须有版本",
			fmt.Sprintf("%s.httpVersion", fieldPath),
		))
	}

	// 验证content
	validateContent(resp.Content, fmt.Sprintf("%s.content", fieldPath), rootError)

	// 验证headers
	validateHeaders(resp.Headers, fmt.Sprintf("%s.headers", fieldPath), rootError)

	// 验证cookies
	validateCookies(resp.Cookies, fmt.Sprintf("%s.cookies", fieldPath), rootError)
}

// validateContent 验证内容
func validateContent(content Content, fieldPath string, rootError *HarError) {
	// 验证MIME类型
	if content.MimeType == "" {
		_ = rootError.AddPartialError(NewValidationError(
			"内容必须有MIME类型",
			fmt.Sprintf("%s.mimeType", fieldPath),
		))
	}

	// 验证size
	if content.Size < 0 {
		_ = rootError.AddPartialError(NewValidationError(
			"内容大小不能为负",
			fmt.Sprintf("%s.size", fieldPath),
		))
	}

	// 验证encoding（如果存在，必须是已知值）
	if content.Encoding != "" &&
		!strings.EqualFold(content.Encoding, "base64") {
		_ = rootError.AddPartialError(NewValidationError(
			fmt.Sprintf("不支持的Content.encoding: %s（仅支持base64）", content.Encoding),
			fmt.Sprintf("%s.encoding", fieldPath),
		))
	}
}

// validateHeaders 验证HTTP头
func validateHeaders(headers []Headers, fieldPath string, rootError *HarError) {
	for i, header := range headers {
		headerPath := fmt.Sprintf("%s[%d]", fieldPath, i)

		if header.Name == "" {
			_ = rootError.AddPartialError(NewValidationError(
				"HTTP头必须有名称",
				fmt.Sprintf("%s.name", headerPath),
			))
		}
	}
}

// validateCookies 验证Cookies
func validateCookies(cookies []Cookie, fieldPath string, rootError *HarError) {
	for i, cookie := range cookies {
		cookiePath := fmt.Sprintf("%s[%d]", fieldPath, i)

		if cookie.Name == "" {
			_ = rootError.AddPartialError(NewValidationError(
				"Cookie必须有名称",
				fmt.Sprintf("%s.name", cookiePath),
			))
		}
	}
}

// validateQueryString 验证查询参数
func validateQueryString(params []QueryString, fieldPath string, rootError *HarError) {
	for i, param := range params {
		paramPath := fmt.Sprintf("%s[%d]", fieldPath, i)

		if param.Name == "" {
			_ = rootError.AddPartialError(NewValidationError(
				"查询参数必须有名称",
				fmt.Sprintf("%s.name", paramPath),
			))
		}
	}
}

// validatePostData 验证POST数据
func validatePostData(postData *PostData, fieldPath string, rootError *HarError) {
	// mimeType是必需字段
	if postData.MimeType == "" {
		_ = rootError.AddPartialError(NewValidationError(
			"PostData必须有mimeType",
			fmt.Sprintf("%s.mimeType", fieldPath),
		))
	}

	// 验证params（如果存在）
	for i, param := range postData.Params {
		paramPath := fmt.Sprintf("%s.params[%d]", fieldPath, i)

		if param.Name == "" {
			_ = rootError.AddPartialError(NewValidationError(
				"PostData参数必须有名称",
				fmt.Sprintf("%s.name", paramPath),
			))
		}
	}
}

// validateTimings 验证时间
func validateTimings(timings Timings, fieldPath string, rootError *HarError) {
	// 验证必要的时间字段
	if timings.Wait < 0 {
		_ = rootError.AddPartialError(NewValidationError(
			"等待时间不能为负",
			fmt.Sprintf("%s.wait", fieldPath),
		))
	}

	if timings.Receive < 0 {
		_ = rootError.AddPartialError(NewValidationError(
			"接收时间不能为负",
			fmt.Sprintf("%s.receive", fieldPath),
		))
	}

	if timings.Send < 0 {
		_ = rootError.AddPartialError(NewValidationError(
			"发送时间不能为负",
			fmt.Sprintf("%s.send", fieldPath),
		))
	}
}

// validatePages 验证页面
func validatePages(pages []Pages, rootError *HarError) {
	for i, page := range pages {
		pagePath := fmt.Sprintf("log.pages[%d]", i)

		// 验证ID
		if page.ID == "" {
			_ = rootError.AddPartialError(NewValidationError(
				"页面必须有ID",
				fmt.Sprintf("%s.id", pagePath),
			))
		}

		// 验证开始时间
		if page.StartedDateTime.IsZero() {
			_ = rootError.AddPartialError(NewValidationError(
				"页面必须有开始时间",
				fmt.Sprintf("%s.startedDateTime", pagePath),
			))
		}

		// 验证页面加载时间
		validatePageTimings(page.PageTimings, fmt.Sprintf("%s.pageTimings", pagePath), rootError)

		// 验证页面标题
		if page.Title == "" {
			_ = rootError.AddPartialError(NewValidationError(
				"页面必须有标题",
				fmt.Sprintf("%s.title", pagePath),
			))
		}
	}
}

// validatePageTimings 验证页面加载时间
func validatePageTimings(timings PageTimings, fieldPath string, rootError *HarError) {
	// onContentLoad和onLoad可以为负值（表示不可用）
	// 但不应为极端值
	if timings.OnContentLoad < -1 {
		_ = rootError.AddPartialError(NewValidationError(
			fmt.Sprintf("页面内容加载时间异常: %f", timings.OnContentLoad),
			fmt.Sprintf("%s.onContentLoad", fieldPath),
		))
	}

	if timings.OnLoad < -1 {
		_ = rootError.AddPartialError(NewValidationError(
			fmt.Sprintf("页面加载时间异常: %f", timings.OnLoad),
			fmt.Sprintf("%s.onLoad", fieldPath),
		))
	}
}


// IsValidHarVersion 检查是否为支持的HAR版本
func IsValidHarVersion(version string) bool {
	return version == HarSpecVersion11 ||
		version == HarSpecVersion12 ||
		version == HarSpecVersion13
}

// DetectHarVersion 检测HAR版本
func DetectHarVersion(har *Har) string {
	if har == nil || har.Log.Version == "" {
		return HarSpecVersion12 // 默认使用1.2版本
	}

	version := strings.TrimSpace(har.Log.Version)
	if IsValidHarVersion(version) {
		return version
	}

	// 如果不是支持的版本，尝试规范化
	if strings.HasPrefix(version, "1.1") {
		return HarSpecVersion11
	} else if strings.HasPrefix(version, "1.2") {
		return HarSpecVersion12
	} else if strings.HasPrefix(version, "1.3") {
		return HarSpecVersion13
	}

	return HarSpecVersion12 // 默认
}
