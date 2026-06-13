package har

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidationRule 定义自定义验证规则
type ValidationRule struct {
	Name        string                                           // 规则名称
	Description string                                          // 规则描述
	Validate    func(har *Har) []*ValidationError               // 验证函数
}

// ValidationError 表示验证错误
type ValidationError struct {
	Field   string // 字段路径
	Message string // 错误消息
	Rule    string // 触发规则的名称
}

// Error 实现error接口
func (e *ValidationError) Error() string {
	if e.Rule != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Rule, e.Field, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// customRules 存储自定义验证规则
var customRules []ValidationRule

// RegisterValidator 注册自定义验证规则
//
// 注册的规则将在调用ValidateWithRules时执行。
// 规则名称必须唯一，重复注册会覆盖之前的规则。
//
// 示例:
//
//	har.RegisterValidator("no-internal-ips", har.ValidationRule{
//	    Name:        "no-internal-ips",
//	    Description: "禁止访问内网IP地址",
//	    Validate: func(h *har.Har) []*har.ValidationError {
//	        var errors []*har.ValidationError
//	        for _, entry := range h.Log.Entries {
//	            if isInternalIP(entry.ServerIPAddress) {
//	                errors = append(errors, &har.ValidationError{
//	                    Field:   "serverIPAddress",
//	                    Message: fmt.Sprintf("内网IP地址: %s", entry.ServerIPAddress),
//	                    Rule:    "no-internal-ips",
//	                })
//	            }
//	        }
//	        return errors
//	    },
//	})
func RegisterValidator(name string, rule ValidationRule) {
	// 检查是否已存在同名规则，如果存在则替换
	for i, r := range customRules {
		if r.Name == name {
			customRules[i] = rule
			return
		}
	}
	customRules = append(customRules, rule)
}

// UnregisterValidator 移除自定义验证规则
func UnregisterValidator(name string) {
	for i, r := range customRules {
		if r.Name == name {
			customRules = append(customRules[:i], customRules[i+1:]...)
			return
		}
	}
}

// ListValidators 列出所有已注册的自定义验证规则
func ListValidators() []ValidationRule {
	result := make([]ValidationRule, len(customRules))
	copy(result, customRules)
	return result
}

// ValidateWithRules 使用自定义验证规则验证HAR对象
//
// 该方法先执行标准的HAR规范验证，再执行所有已注册的自定义验证规则。
// 返回标准验证错误和自定义规则验证错误。
func ValidateWithRules(har *Har) error {
	if har == nil {
		return NewInvalidFormatError("HAR对象为空")
	}

	// 执行标准验证
	stdErr := ValidateHarFile(har)

	// 执行自定义规则验证
	var customErrors []*ValidationError
	for _, rule := range customRules {
		errs := rule.Validate(har)
		customErrors = append(customErrors, errs...)
	}

	if stdErr != nil && len(customErrors) > 0 {
		// 合并标准错误和自定义错误
		if harErr, ok := stdErr.(*HarError); ok {
			for _, ce := range customErrors {
				harErr.AddPartialError(NewValidationError(
					fmt.Sprintf("[%s] %s", ce.Rule, ce.Message),
					ce.Field,
				))
			}
			return harErr
		}
	}

	if len(customErrors) > 0 {
		rootError := &HarError{
			Code:    ErrCodeValidation,
			Message: "HAR验证失败（包含自定义规则）",
		}
		for _, ce := range customErrors {
			rootError.AddPartialError(NewValidationError(
				fmt.Sprintf("[%s] %s", ce.Rule, ce.Message),
				ce.Field,
			))
		}
		return rootError
	}

	return stdErr
}

// ValidateStrict 严格验证HAR对象
//
// 比标准验证更严格，包括：
// - 验证pageref交叉引用
// - 验证页面ID唯一性
// - 验证HTTP方法值
// - 验证状态码范围
// - 验证Cookie.SameSite值
// - 验证Timings与Time一致性
// - 验证Cache的必填字段
func ValidateStrict(har *Har) error {
	if har == nil {
		return NewInvalidFormatError("HAR对象为空")
	}

	// 先执行标准验证
	rootError := &HarError{
		Code:    ErrCodeValidation,
		Message: "HAR严格验证失败",
	}

	// 标准验证
	if err := ValidateHarFile(har); err != nil {
		if harErr, ok := err.(*HarError); ok {
			for _, pe := range harErr.GetPartialErrors() {
				rootError.AddPartialError(pe)
			}
		}
	}

	// 严格验证：pageref交叉引用
	validatePagerefReferences(har, rootError)

	// 严格验证：页面ID唯一性
	validatePageIDUniqueness(har, rootError)

	// 严格验证：HTTP方法值
	validateHTTPMethods(har, rootError)

	// 严格验证：状态码范围
	validateStatusCodeRange(har, rootError)

	// 严格验证：Cookie.SameSite值
	validateCookieSameSite(har, rootError)

	// 严格验证：Cache必填字段
	validateCacheFields(har, rootError)

	if rootError.HasPartialErrors() {
		return rootError
	}

	return nil
}

// validatePagerefReferences 验证pageref引用的页面是否存在
func validatePagerefReferences(har *Har, rootError *HarError) {
	// 构建页面ID集合
	pageIDs := make(map[string]bool)
	for _, page := range har.Log.Pages {
		pageIDs[page.ID] = true
	}

	// 检查每个条目的pageref
	for i, entry := range har.Log.Entries {
		if entry.Pageref != "" {
			if !pageIDs[entry.Pageref] {
				rootError.AddPartialError(NewValidationError(
					fmt.Sprintf("条目引用的页面ID '%s' 不存在", entry.Pageref),
					fmt.Sprintf("log.entries[%d].pageref", i),
				))
			}
		}
	}
}

// validatePageIDUniqueness 验证页面ID的唯一性
func validatePageIDUniqueness(har *Har, rootError *HarError) {
	seen := make(map[string]int) // ID -> first occurrence index
	for i, page := range har.Log.Pages {
		if prevIdx, exists := seen[page.ID]; exists {
			rootError.AddPartialError(NewValidationError(
				fmt.Sprintf("页面ID '%s' 重复（首次出现在索引%d）", page.ID, prevIdx),
				fmt.Sprintf("log.pages[%d].id", i),
			))
		} else {
			seen[page.ID] = i
		}
	}
}

// validateHTTPMethods 验证HTTP方法的合法性
func validateHTTPMethods(har *Har, rootError *HarError) {
	validMethods := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"HEAD":    true,
		"OPTIONS": true,
		"PATCH":   true,
		"CONNECT": true,
		"TRACE":   true,
	}

	for i, entry := range har.Log.Entries {
		method := strings.ToUpper(entry.Request.Method)
		if !validMethods[method] {
			rootError.AddPartialError(NewValidationError(
				fmt.Sprintf("不常见的HTTP方法: %s（标准方法: GET, POST, PUT, DELETE, HEAD, OPTIONS, PATCH, CONNECT, TRACE）", entry.Request.Method),
				fmt.Sprintf("log.entries[%d].request.method", i),
			))
		}
	}
}

// validateStatusCodeRange 验证状态码范围
func validateStatusCodeRange(har *Har, rootError *HarError) {
	for i, entry := range har.Log.Entries {
		status := entry.Response.Status
		if status < 100 || status > 599 {
			rootError.AddPartialError(NewValidationError(
				fmt.Sprintf("无效的HTTP状态码: %d（有效范围: 100-599）", status),
				fmt.Sprintf("log.entries[%d].response.status", i),
			))
		}
	}
}

// validateCookieSameSite 验证Cookie.SameSite值
func validateCookieSameSite(har *Har, rootError *HarError) {
	validSameSite := map[string]bool{
		"Strict": true,
		"Lax":    true,
		"None":   true,
		"":       true, // 空值是合法的（未设置）
	}

	for i, entry := range har.Log.Entries {
		// 检查请求Cookie
		for j, cookie := range entry.Request.Cookies {
			if !validSameSite[cookie.SameSite] {
				rootError.AddPartialError(NewValidationError(
					fmt.Sprintf("Cookie.SameSite值无效: '%s'（有效值: Strict, Lax, None）", cookie.SameSite),
					fmt.Sprintf("log.entries[%d].request.cookies[%d].sameSite", i, j),
				))
			}
		}

		// 检查响应Cookie
		for j, cookie := range entry.Response.Cookies {
			if !validSameSite[cookie.SameSite] {
				rootError.AddPartialError(NewValidationError(
					fmt.Sprintf("Cookie.SameSite值无效: '%s'（有效值: Strict, Lax, None）", cookie.SameSite),
					fmt.Sprintf("log.entries[%d].response.cookies[%d].sameSite", i, j),
				))
			}
		}
	}
}

// validateCacheFields 验证Cache的必填字段
func validateCacheFields(har *Har, rootError *HarError) {
	for i, entry := range har.Log.Entries {
		entryPrefix := fmt.Sprintf("log.entries[%d]", i)

		// 验证BeforeRequest
		if entry.Cache.BeforeRequest != nil {
			br := entry.Cache.BeforeRequest
			brPrefix := fmt.Sprintf("%s.cache.beforeRequest", entryPrefix)

			if br.LastAccess.IsZero() {
				rootError.AddPartialError(NewValidationError(
					"BeforeRequest必须有lastAccess字段",
					fmt.Sprintf("%s.lastAccess", brPrefix),
				))
			}
			if br.ETag == "" {
				rootError.AddPartialError(NewValidationError(
					"BeforeRequest必须有eTag字段",
					fmt.Sprintf("%s.eTag", brPrefix),
				))
			}
			if br.HitCount < 0 {
				rootError.AddPartialError(NewValidationError(
					"BeforeRequest.hitCount不能为负",
					fmt.Sprintf("%s.hitCount", brPrefix),
				))
			}
		}

		// 验证AfterRequest
		if entry.Cache.AfterRequest != nil {
			ar := entry.Cache.AfterRequest
			arPrefix := fmt.Sprintf("%s.cache.afterRequest", entryPrefix)

			if ar.LastAccess.IsZero() {
				rootError.AddPartialError(NewValidationError(
					"AfterRequest必须有lastAccess字段",
					fmt.Sprintf("%s.lastAccess", arPrefix),
				))
			}
			if ar.ETag == "" {
				rootError.AddPartialError(NewValidationError(
					"AfterRequest必须有eTag字段",
					fmt.Sprintf("%s.eTag", arPrefix),
				))
			}
			if ar.HitCount < 0 {
				rootError.AddPartialError(NewValidationError(
					"AfterRequest.hitCount不能为负",
					fmt.Sprintf("%s.hitCount", arPrefix),
				))
			}
		}
	}
}

// ValidateURL 严格验证URL格式
//
// 检查URL是否包含有效的scheme和host
func ValidateURL(rawURL string) *ValidationError {
	if rawURL == "" {
		return &ValidationError{
			Field:   "url",
			Message: "URL不能为空",
		}
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return &ValidationError{
			Field:   "url",
			Message: fmt.Sprintf("URL解析失败: %v", err),
		}
	}

	if parsed.Scheme == "" {
		return &ValidationError{
			Field:   "url.scheme",
			Message: "URL缺少scheme（如http://或https://）",
		}
	}

	if parsed.Host == "" {
		return &ValidationError{
			Field:   "url.host",
			Message: "URL缺少host",
		}
	}

	return nil
}

// ValidateTimingsConsistency 验证Timings与Entries.Time的一致性
//
// HAR规范要求Time字段等于各timing字段之和（减去SSL与Connect的重叠）。
// 此方法检查差异是否超过给定的容差（毫秒）。
func ValidateTimingsConsistency(har *Har, tolerance float64) []*ValidationError {
	var errors []*ValidationError

	for i, entry := range har.Log.Entries {
		// 计算timing总和
		var sum float64
		if entry.Timings.Blocked > 0 {
			sum += entry.Timings.Blocked
		}
		if entry.Timings.DNS > 0 {
			sum += entry.Timings.DNS
		}
		if entry.Timings.Connect > 0 {
			sum += entry.Timings.Connect
		}
		// SSL时间包含在Connect中，不重复计算（如果Connect > 0且SSL > 0）
		if entry.Timings.Ssl > 0 && entry.Timings.Connect <= 0 {
			sum += entry.Timings.Ssl
		}
		if entry.Timings.Send > 0 {
			sum += entry.Timings.Send
		}
		if entry.Timings.Wait > 0 {
			sum += entry.Timings.Wait
		}
		if entry.Timings.Receive > 0 {
			sum += entry.Timings.Receive
		}

		// 检查差异
		diff := entry.Time - sum
		if diff < 0 {
			diff = -diff
		}
		if diff > tolerance {
			errors = append(errors, &ValidationError{
				Field:   fmt.Sprintf("log.entries[%d].time", i),
				Message: fmt.Sprintf("Time字段(%.2fms)与Timings总和(%.2fms)差异超过容差(%.2fms)", entry.Time, sum, tolerance),
				Rule:    "timings-consistency",
			})
		}
	}

	return errors
}
