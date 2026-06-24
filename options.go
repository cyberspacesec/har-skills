package har

// Option 定义解析器的选项
type Option func(*options)

// options 内部选项结构体，用于处理所有配置选项
type options struct {
	// 是否开启宽松解析模式，会尽量解析有效部分
	lenient bool
	// 是否跳过验证
	skipValidation bool
	// 是否记录所有解析警告（不会导致解析失败）
	collectWarnings bool
	// 最大允许的警告数量，超过则停止解析
	maxWarnings int
	// 是否使用内存优化结构
	useMemoryOptimized bool
	// 是否使用懒加载
	useLazyLoading bool
	// 是否使用流式处理
	useStreaming bool
	// 指定HAR版本
	harVersion string
	// 是否自动检测版本
	autoDetectVersion bool
}

// 默认选项
var defaultOptions = options{
	lenient:            false,
	skipValidation:     false,
	collectWarnings:    false,
	maxWarnings:        100,
	useMemoryOptimized: false,
	useLazyLoading:     false,
	useStreaming:       false,
	harVersion:         HarSpecVersion12, // 默认版本1.2
	autoDetectVersion:  true,             // 默认开启自动检测
}

// 将options转换为旧版ParseOptions结构
func (o *options) toParseOptions() ParseOptions {
	return ParseOptions{
		Lenient:         o.lenient,
		SkipValidation:  o.skipValidation,
		CollectWarnings: o.collectWarnings,
		MaxWarnings:     o.maxWarnings,
	}
}

// WithLenient 启用宽松解析模式
func WithLenient() Option {
	return func(o *options) {
		o.lenient = true
	}
}

// WithSkipValidation 跳过验证
func WithSkipValidation() Option {
	return func(o *options) {
		o.skipValidation = true
	}
}

// WithCollectWarnings 收集警告而不是失败
func WithCollectWarnings() Option {
	return func(o *options) {
		o.collectWarnings = true
	}
}

// WithMaxWarnings 设置最大警告数量
func WithMaxWarnings(max int) Option {
	return func(o *options) {
		o.maxWarnings = max
	}
}

// WithMemoryOptimized 使用内存优化结构
func WithMemoryOptimized() Option {
	return func(o *options) {
		o.useMemoryOptimized = true
	}
}

// WithLazyLoading 使用懒加载
func WithLazyLoading() Option {
	return func(o *options) {
		o.useLazyLoading = true
	}
}

// WithStreaming 使用流式处理
func WithStreaming() Option {
	return func(o *options) {
		o.useStreaming = true
	}
}

// WithHarVersion 指定HAR版本
func WithHarVersion(version string) Option {
	return func(o *options) {
		if IsValidHarVersion(version) {
			o.harVersion = version
			o.autoDetectVersion = false
		}
	}
}

// WithAutoDetectVersion 自动检测HAR版本
func WithAutoDetectVersion(enabled bool) Option {
	return func(o *options) {
		o.autoDetectVersion = enabled
	}
}

// applyOptions 应用选项到默认选项并返回结果
func applyOptions(opts ...Option) options {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// 定义常用的选项组合
var (
	// OptMemoryEfficient 内存高效配置
	OptMemoryEfficient = []Option{
		WithMemoryOptimized(),
		WithSkipValidation(),
	}

	// OptFast 快速解析配置
	OptFast = []Option{
		WithSkipValidation(),
	}

	// OptLenient 宽松解析配置
	OptLenient = []Option{
		WithLenient(),
		WithCollectWarnings(),
	}

	// OptPerformance 高性能配置
	OptPerformance = []Option{
		WithMemoryOptimized(),
		WithSkipValidation(),
		WithLazyLoading(),
	}
)
