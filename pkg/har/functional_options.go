package har

import (
	"time"
)

// FilterOption 定义过滤选项的函数式选项类型
type FilterOption func(*FilterOptions)

// WithFilterURL 设置URL过滤条件
func WithFilterURL(url string) FilterOption {
	return func(o *FilterOptions) {
		o.URL = url
	}
}

// WithFilterMethod 设置请求方法过滤条件
func WithFilterMethod(method string) FilterOption {
	return func(o *FilterOptions) {
		o.Method = method
	}
}

// WithFilterStatusCode 设置状态码过滤条件
func WithFilterStatusCode(code int) FilterOption {
	return func(o *FilterOptions) {
		o.StatusCode = code
	}
}

// WithFilterStatusCodeRange 设置状态码范围过滤条件
func WithFilterStatusCodeRange(min, max int) FilterOption {
	return func(o *FilterOptions) {
		o.StatusCodeMin = min
		o.StatusCodeMax = max
	}
}

// WithFilterContentType 设置内容类型过滤条件
func WithFilterContentType(contentType string) FilterOption {
	return func(o *FilterOptions) {
		o.ContentType = contentType
	}
}

// WithFilterTimeRange 设置时间范围过滤条件
func WithFilterTimeRange(start, end time.Time) FilterOption {
	return func(o *FilterOptions) {
		o.StartTime = start
		o.EndTime = end
	}
}

// WithFilterDuration 设置持续时间过滤条件
func WithFilterDuration(min, max float64) FilterOption {
	return func(o *FilterOptions) {
		o.MinDuration = min
		o.MaxDuration = max
	}
}

// WithFilterResourceType 设置资源类型过滤条件
func WithFilterResourceType(resourceType string) FilterOption {
	return func(o *FilterOptions) {
		o.ResourceType = resourceType
	}
}

// WithFilterHasError 设置只过滤有错误的请求
func WithFilterHasError() FilterOption {
	return func(o *FilterOptions) {
		o.HasError = true
	}
}

// WithFilterHeader 设置请求头过滤条件
func WithFilterHeader(name, value string) FilterOption {
	return func(o *FilterOptions) {
		o.HeaderName = name
		o.HeaderValue = value
	}
}

// WithFilterResponseHeader 设置响应头过滤条件
func WithFilterResponseHeader(name, value string) FilterOption {
	return func(o *FilterOptions) {
		o.RespHeaderName = name
		o.RespHeaderValue = value
	}
}

// WithFilterRegex 启用正则表达式匹配
func WithFilterRegex() FilterOption {
	return func(o *FilterOptions) {
		o.UseRegex = true
	}
}

// NewFilterOptions 从函数式选项创建过滤选项
func NewFilterOptions(opts ...FilterOption) FilterOptions {
	options := FilterOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// FilterWith 使用函数式选项过滤HAR条目
func (h *Har) FilterWith(opts ...FilterOption) *FilterResult {
	return h.Filter(NewFilterOptions(opts...))
}

// ReplayOption 定义重放选项的函数式选项类型
type ReplayOption func(*ReplayOptions)

// WithReplayTimeout 设置请求超时时间
func WithReplayTimeout(timeout time.Duration) ReplayOption {
	return func(o *ReplayOptions) {
		o.Timeout = timeout
	}
}

// WithReplayFollowRedirects 设置是否跟随重定向
func WithReplayFollowRedirects(follow bool) ReplayOption {
	return func(o *ReplayOptions) {
		o.FollowRedirects = follow
	}
}

// WithReplayMaxRedirects 设置最大重定向次数
func WithReplayMaxRedirects(max int) ReplayOption {
	return func(o *ReplayOptions) {
		o.MaxRedirects = max
	}
}

// WithReplaySkipSSLVerify 设置是否跳过SSL证书验证
func WithReplaySkipSSLVerify(skip bool) ReplayOption {
	return func(o *ReplayOptions) {
		o.SkipSSLVerify = skip
	}
}

// WithReplayOverrideHeader 设置覆盖的请求头
func WithReplayOverrideHeader(name, value string) ReplayOption {
	return func(o *ReplayOptions) {
		if o.OverrideHeaders == nil {
			o.OverrideHeaders = make(map[string]string)
		}
		o.OverrideHeaders[name] = value
	}
}

// WithReplayTransport 设置自定义Transport
func WithReplayTransport(transport interface{}) ReplayOption {
	return func(o *ReplayOptions) {
		if t, ok := transport.(interface {
			RoundTrip(*Request)
		}); ok {
			_ = t // 仅为类型断言示例
		}
	}
}

// NewReplayOptions 从函数式选项创建重放选项
func NewReplayOptions(opts ...ReplayOption) ReplayOptions {
	options := DefaultReplayOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// ReplayAllWith 使用函数式选项重放所有请求
func (h *Har) ReplayAllWith(opts ...ReplayOption) ([]*ReplayResult, error) {
	return h.ReplayAll(NewReplayOptions(opts...))
}

// ConvertOption 定义转换选项的函数式选项类型
type ConvertOption func(*ConvertOptions)

// WithConvertIncludeHeaders 设置是否包含头部
func WithConvertIncludeHeaders(include bool) ConvertOption {
	return func(o *ConvertOptions) {
		o.IncludeHeaders = include
	}
}

// WithConvertIncludeTimings 设置是否包含时间
func WithConvertIncludeTimings(include bool) ConvertOption {
	return func(o *ConvertOptions) {
		o.IncludeTimings = include
	}
}

// WithConvertIncludeBodies 设置是否包含请求体
func WithConvertIncludeBodies(include bool) ConvertOption {
	return func(o *ConvertOptions) {
		o.IncludePostData = include
	}
}

// WithConvertIncludeCookies 设置是否包含Cookie
func WithConvertIncludeCookies(include bool) ConvertOption {
	return func(o *ConvertOptions) {
		// Cookies are included via headers; no separate field in ConvertOptions
	}
}

// WithConvertIncludeQueryStrings 设置是否包含查询参数
func WithConvertIncludeQueryStrings(include bool) ConvertOption {
	return func(o *ConvertOptions) {
		o.IncludeQueryString = include
	}
}

// WithConvertIncludeStatus 设置是否包含状态码
func WithConvertIncludeStatus(include bool) ConvertOption {
	return func(o *ConvertOptions) {
		o.IncludeStatus = include
	}
}

// WithConvertIncludeSize 设置是否包含大小
func WithConvertIncludeSize(include bool) ConvertOption {
	return func(o *ConvertOptions) {
		o.IncludeSize = include
	}
}

// WithConvertIncludeURL 设置是否包含URL
func WithConvertIncludeURL(include bool) ConvertOption {
	return func(o *ConvertOptions) {
		o.IncludeURL = include
	}
}

// WithConvertIncludeMethod 设置是否包含方法
func WithConvertIncludeMethod(include bool) ConvertOption {
	return func(o *ConvertOptions) {
		o.IncludeMethod = include
	}
}

// WithConvertIncludeTime 设置是否包含时间
func WithConvertIncludeTime(include bool) ConvertOption {
	return func(o *ConvertOptions) {
		o.IncludeTime = include
	}
}

// WithConvertIncludeMimeType 设置是否包含MIME类型
func WithConvertIncludeMimeType(include bool) ConvertOption {
	return func(o *ConvertOptions) {
		o.IncludeContentType = include
	}
}

// WithConvertHeaders 设置自定义头部列表
func WithConvertHeaders(headers []string) ConvertOption {
	return func(o *ConvertOptions) {
		o.Headers = headers
	}
}

// WithConvertFilter 设置过滤选项
func WithConvertFilter(filter FilterOptions) ConvertOption {
	return func(o *ConvertOptions) {
		o.Filter = &filter
	}
}

// NewConvertOptions 从函数式选项创建转换选项
func NewConvertOptions(opts ...ConvertOption) ConvertOptions {
	options := DefaultConvertOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// ConvertWith 使用函数式选项转换HAR格式
func (h *Har) ConvertWith(format ConvertFormat, opts ...ConvertOption) (string, error) {
	return h.Convert(format, NewConvertOptions(opts...))
}

// DiffOption 定义差异比较选项的函数式选项类型
type DiffOption func(*DiffOptions)

// WithDiffIgnoreHeaders 设置忽略的头部字段
func WithDiffIgnoreHeaders(headers ...string) DiffOption {
	return func(o *DiffOptions) {
		o.IgnoreHeaders = append(o.IgnoreHeaders, headers...)
	}
}

// WithDiffIgnoreTimings 设置忽略时间差异
func WithDiffIgnoreTimings(ignore bool) DiffOption {
	return func(o *DiffOptions) {
		o.IgnoreTimings = ignore
	}
}

// WithDiffIgnoreDates 设置忽略日期差异
func WithDiffIgnoreDates(ignore bool) DiffOption {
	return func(o *DiffOptions) {
		o.IgnoreDates = ignore
	}
}

// WithDiffIgnoreCache 设置忽略缓存差异
func WithDiffIgnoreCache(ignore bool) DiffOption {
	return func(o *DiffOptions) {
		o.IgnoreCache = ignore
	}
}

// WithDiffIgnoreComment 设置忽略注释差异
func WithDiffIgnoreComment(ignore bool) DiffOption {
	return func(o *DiffOptions) {
		o.IgnoreComment = ignore
	}
}

// WithDiffNormalizeURL 设置URL归一化
func WithDiffNormalizeURL(normalize bool) DiffOption {
	return func(o *DiffOptions) {
		o.NormalizeURL = normalize
	}
}

// WithDiffCompareByURL 设置按URL匹配
func WithDiffCompareByURL(compare bool) DiffOption {
	return func(o *DiffOptions) {
		o.CompareByURL = compare
	}
}

// WithDiffIncludeBody 设置比较响应体
func WithDiffIncludeBody(include bool) DiffOption {
	return func(o *DiffOptions) {
		o.IncludeBody = include
	}
}

// NewDiffOptions 从函数式选项创建差异比较选项
func NewDiffOptions(opts ...DiffOption) DiffOptions {
	options := DefaultDiffOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// DiffWith 使用函数式选项比较两个HAR文件的差异
func DiffWith(har1, har2 *Har, opts ...DiffOption) *HarDiff {
	return Diff(har1, har2, NewDiffOptions(opts...))
}

// MergeOption 定义合并选项的函数式选项类型
type MergeOption func(*MergeOptions)

// WithMergeSortByTime 设置按时间排序
func WithMergeSortByTime(sort bool) MergeOption {
	return func(o *MergeOptions) {
		o.SortByTime = sort
	}
}

// WithMergeDeduplicate 设置去重
func WithMergeDeduplicate(dedup bool) MergeOption {
	return func(o *MergeOptions) {
		o.Deduplicate = dedup
	}
}

// NewMergeOptions 从函数式选项创建合并选项
func NewMergeOptions(opts ...MergeOption) MergeOptions {
	options := DefaultMergeOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// MergeWith 使用函数式选项合并多个HAR文件
func MergeWith(opts ...MergeOption) func(hars ...*Har) *Har {
	options := NewMergeOptions(opts...)
	return func(hars ...*Har) *Har {
		return MergeWithOptions(options, hars...)
	}
}

// HarBuilderOption 定义HAR构建器的函数式选项类型
type HarBuilderOption func(*HarBuilder)

// WithBuilderVersion 设置HAR版本
func WithBuilderVersion(version string) HarBuilderOption {
	return func(b *HarBuilder) {
		b.SetVersion(version)
	}
}

// WithBuilderCreator 设置创建者信息
func WithBuilderCreator(name, version string) HarBuilderOption {
	return func(b *HarBuilder) {
		b.SetCreator(name, version)
	}
}

// WithBuilderBrowser 设置浏览器信息
func WithBuilderBrowser(name, version string) HarBuilderOption {
	return func(b *HarBuilder) {
		b.SetBrowser(name, version)
	}
}

// WithBuilderComment 设置注释
func WithBuilderComment(comment string) HarBuilderOption {
	return func(b *HarBuilder) {
		b.SetComment(comment)
	}
}

// NewHarBuilderWithOptions 使用函数式选项创建HAR构建器
func NewHarBuilderWithOptions(opts ...HarBuilderOption) *HarBuilder {
	builder := NewHarBuilder()
	for _, opt := range opts {
		opt(builder)
	}
	return builder
}
