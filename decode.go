package har

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
)

// DecodeContent 解码响应内容
//
// 该方法会自动检测编码方式（base64）并解码，
// 同时检测 Content-Encoding（gzip/deflate）并解压。
// 返回解码后的原始字节数据。
func (c *Content) DecodeContent() ([]byte, error) {
	if c == nil {
		return nil, NewInvalidFormatError("内容为空")
	}

	var data []byte

	// 步骤1：处理base64编码
	if strings.EqualFold(c.Encoding, "base64") && c.Text != "" {
		decoded, err := base64.StdEncoding.DecodeString(c.Text)
		if err != nil {
			// 尝试URL安全的base64
			decoded, err = base64.URLEncoding.DecodeString(c.Text)
			if err != nil {
				return nil, NewHarError(ErrCodeInvalidFormat,
					fmt.Sprintf("base64解码失败: %v", err), err)
			}
		}
		data = decoded
	} else if c.Text != "" {
		data = []byte(c.Text)
	} else {
		return nil, nil
	}

	// 步骤2：检测并解压内容
	data, err := decompressIfNeeded(data, c.MimeType)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// DecodeContent 解码指定条目的响应内容
func (e *Entries) DecodeContent() ([]byte, error) {
	if e == nil {
		return nil, NewInvalidFormatError("条目为空")
	}
	return e.Response.Content.DecodeContent()
}

// DecodeAllContent 解码HAR中所有条目的响应内容
// 返回每个条目的解码结果，索引与HAR条目一一对应
func (h *Har) DecodeAllContent() ([][]byte, error) {
	if h == nil {
		return nil, NewInvalidFormatError("HAR对象为空")
	}

	results := make([][]byte, len(h.Log.Entries))
	var partialErrors []*HarError

	for i, entry := range h.Log.Entries {
		data, err := entry.DecodeContent()
		if err != nil {
			partialErrors = append(partialErrors, decodePartialError(i, err))
			results[i] = nil
			continue
		}
		results[i] = data
	}

	if len(partialErrors) > 0 {
		rootErr := NewHarError(ErrCodeInvalidFormat,
			fmt.Sprintf("解码过程中有%d个错误", len(partialErrors)), nil).
			WithMetadata("error_count", len(partialErrors))
		for _, err := range partialErrors {
			rootErr = rootErr.AddPartialError(err)
		}
		return results, rootErr
	}

	return results, nil
}

func decodePartialError(index int, err error) *HarError {
	field := fmt.Sprintf("log.entries[%d].response.content", index)
	if harErr, ok := err.(*HarError); ok {
		return NewHarError(harErr.Code, harErr.Message, harErr.Err).
			WithField(field).
			WithMetadata("entry_index", index)
	}

	return NewHarError(ErrCodeInvalidFormat, "内容解码失败", err).
		WithField(field).
		WithMetadata("entry_index", index)
}

// IsBase64Encoded 检查内容是否为base64编码
func (c *Content) IsBase64Encoded() bool {
	if c == nil {
		return false
	}
	return strings.EqualFold(c.Encoding, "base64")
}

// IsCompressed 检查内容是否被压缩（根据Content-Type头部或MimeType判断）
func (e *Entries) IsCompressed() bool {
	if e == nil {
		return false
	}

	// 检查响应头中的Content-Encoding
	for _, header := range e.Response.Headers {
		if strings.EqualFold(header.Name, "Content-Encoding") {
			encoding := strings.ToLower(strings.TrimSpace(header.Value))
			if encoding == "gzip" || encoding == "deflate" || encoding == "br" || encoding == "zstd" {
				return true
			}
		}
	}

	return false
}

// GetContentEncoding 获取内容编码方式
func (e *Entries) GetContentEncoding() string {
	if e == nil {
		return ""
	}

	for _, header := range e.Response.Headers {
		if strings.EqualFold(header.Name, "Content-Encoding") {
			return strings.TrimSpace(header.Value)
		}
	}

	return ""
}

// DecodeEntryText 解码条目的响应文本（便捷方法）
func (e *Entries) DecodeEntryText() (string, error) {
	data, err := e.DecodeContent()
	if err != nil {
		return "", err
	}
	if data == nil {
		return "", nil
	}
	return string(data), nil
}

// decompressIfNeeded 根据MIME类型和内容特征尝试解压
func decompressIfNeeded(data []byte, mimeType string) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	// 尝试gzip解压
	if isGzipData(data) {
		reader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, NewHarError(ErrCodeInvalidFormat,
				fmt.Sprintf("gzip解压失败: %v", err), err)
		}
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return nil, NewHarError(ErrCodeInvalidFormat,
				fmt.Sprintf("gzip解压失败: %v", err), err)
		}
		return decompressed, nil
	}

	// 尝试deflate解压
	if isDeflateData(data) {
		// isDeflateData 仅接受通过 zlib 头部校验的 FLG 值，因此
		// zlib.NewReader 在此处必定成功。
		reader, _ := zlib.NewReader(bytes.NewReader(data))
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return nil, NewHarError(ErrCodeInvalidFormat,
				fmt.Sprintf("deflate解压失败: %v", err), err)
		}
		return decompressed, nil
	}

	// 尝试brotli解压
	if isBrotliData(data) {
		reader := brotli.NewReader(bytes.NewReader(data))
		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return nil, NewHarError(ErrCodeInvalidFormat,
				fmt.Sprintf("brotli解压失败: %v", err), err)
		}
		return decompressed, nil
	}

	// 尝试zstd解压
	if isZstdData(data) {
		decoder, err := zstd.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, NewHarError(ErrCodeInvalidFormat,
				fmt.Sprintf("zstd初始化失败: %v", err), err)
		}
		defer decoder.Close()

		decompressed, err := decoder.DecodeAll(data, nil)
		if err != nil {
			return nil, NewHarError(ErrCodeInvalidFormat,
				fmt.Sprintf("zstd解压失败: %v", err), err)
		}
		return decompressed, nil
	}

	return data, nil
}

// DecompressByEncoding 根据Content-Encoding值解压数据
//
// 支持的编码: "gzip", "deflate", "br" (brotli), "zstd", "identity"
// 多重编码（如 "gzip, deflate"）按 HTTP 语义逐层解压：
// Content-Encoding 头按声明顺序表示包裹的层次——列表中的第一个编码
// 是最外层（最后应用的），应最先解压。例如 "gzip, deflate" 表示数据
// 是 gzip(deflate(original))，解压时先解 gzip 再解 deflate。
func DecompressByEncoding(data []byte, encoding string) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	enc := strings.ToLower(strings.TrimSpace(encoding))
	if enc == "" || enc == "identity" {
		return data, nil
	}

	// 多重编码：按逗号拆分，按声明顺序正序逐层解压
	// （列表第一个是最外层，最先解）
	if strings.Contains(enc, ",") {
		encodings := splitEncodings(enc)
		current := data
		for i, e := range encodings {
			result, err := DecompressByEncoding(current, e)
			if err != nil {
				return nil, fmt.Errorf("第%d层 %q 解压失败: %w", i, e, err)
			}
			current = result
		}
		return current, nil
	}

	switch enc {
	case "", "identity":
		return data, nil
	case "gzip":
		reader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, NewHarError(ErrCodeInvalidFormat,
				fmt.Sprintf("gzip解压失败: %v", err), err)
		}
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return nil, NewHarError(ErrCodeInvalidFormat,
				fmt.Sprintf("gzip解压失败: %v", err), err)
		}
		return decompressed, nil
	case "deflate":
		reader, err := zlib.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, NewHarError(ErrCodeInvalidFormat,
				fmt.Sprintf("deflate解压失败: %v", err), err)
		}
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return nil, NewHarError(ErrCodeInvalidFormat,
				fmt.Sprintf("deflate解压失败: %v", err), err)
		}
		return decompressed, nil
	case "br":
		reader := brotli.NewReader(bytes.NewReader(data))
		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return nil, NewHarError(ErrCodeInvalidFormat,
				fmt.Sprintf("brotli解压失败: %v", err), err)
		}
		return decompressed, nil
	case "zstd":
		decoder, err := zstd.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, NewHarError(ErrCodeInvalidFormat,
				fmt.Sprintf("zstd初始化失败: %v", err), err)
		}
		defer decoder.Close()

		decompressed, err := decoder.DecodeAll(data, nil)
		if err != nil {
			return nil, NewHarError(ErrCodeInvalidFormat,
				fmt.Sprintf("zstd解压失败: %v", err), err)
		}
		return decompressed, nil
	default:
		return nil, NewUnsupportedError(
			fmt.Sprintf("不支持的Content-Encoding: %q", encoding))
	}
}

// splitEncodings 把 "gzip, deflate, br" 拆成 ["gzip", "deflate", "br"]（去空白、忽略空段）
func splitEncodings(enc string) []string {
	parts := strings.Split(enc, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// CompressContent 使用指定的编码方式压缩数据
//
// 支持的编码: "gzip", "deflate", "br" (brotli), "zstd", "identity"
func CompressContent(data []byte, encoding string) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	enc := strings.ToLower(strings.TrimSpace(encoding))

	switch enc {
	case "gzip":
		var buf bytes.Buffer
		writer := gzip.NewWriter(&buf)
		// bytes.Buffer.Write 永不返回错误，故 Write/Close 不会失败。
		_, _ = writer.Write(data)
		_ = writer.Close()
		return buf.Bytes(), nil
	case "deflate":
		var buf bytes.Buffer
		writer := zlib.NewWriter(&buf)
		_, _ = writer.Write(data)
		_ = writer.Close()
		return buf.Bytes(), nil
	case "br":
		var buf bytes.Buffer
		writer := brotli.NewWriter(&buf)
		if _, err := writer.Write(data); err != nil {
			return nil, NewHarError(ErrCodeInvalidFormat,
				fmt.Sprintf("brotli压缩失败: %v", err), err)
		}
		if err := writer.Close(); err != nil {
			return nil, NewHarError(ErrCodeInvalidFormat,
				fmt.Sprintf("brotli压缩关闭失败: %v", err), err)
		}
		return buf.Bytes(), nil
	case "zstd":
		encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedDefault))
		if err != nil {
			return nil, NewHarError(ErrCodeInvalidFormat,
				fmt.Sprintf("zstd初始化失败: %v", err), err)
		}
		defer encoder.Close()
		return encoder.EncodeAll(data, nil), nil
	default:
		return nil, NewUnsupportedError(
			fmt.Sprintf("不支持的压缩编码: %q", encoding))
	}
}

// DecompressWithEncoding 使用Content-Encoding头部值解压数据
//
// 该函数根据HTTP Content-Encoding头部值来决定如何解压数据，
// 与 decompressIfNeeded 不同，后者仅依赖magic bytes检测。
func DecompressWithEncoding(data []byte, contentEncoding string) ([]byte, error) {
	return DecompressByEncoding(data, contentEncoding)
}

// isGzipData 检查数据是否为gzip格式
func isGzipData(data []byte) bool {
	if len(data) < 2 {
		return false
	}
	// gzip magic number: 0x1f 0x8b
	return data[0] == 0x1f && data[1] == 0x8b
}

// isDeflateData 检查数据是否为deflate格式
func isDeflateData(data []byte) bool {
	if len(data) < 2 {
		return false
	}
	// zlib header: 通常以 0x78 开头
	// 0x78 0x01 = no compression
	// 0x78 0x5E = best speed
	// 0x78 0x9C = default compression
	// 0x78 0xDA = best compression
	return data[0] == 0x78 && (data[1] == 0x01 || data[1] == 0x5e || data[1] == 0x9c || data[1] == 0xda)
}

// isZstdData 检查数据是否为 zstd 格式（通过 4 字节魔数 0x28 0xB5 0x2F 0xFD）
func isZstdData(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	return data[0] == 0x28 && data[1] == 0xb5 && data[2] == 0x2f && data[3] == 0xfd
}

// isBrotliData 启发式判断数据是否为 brotli 压缩。
// brotli 没有像 gzip/zstd 那样的固定魔数，本函数靠尝试解码首段来判断：
// 用 brotli.Reader 读取前若干字节，若能成功读出非空结果且无错误，则视为 brotli。
// 误判概率很低（普通文本/JSON 几乎不会被 brotli 解码出有效字节）。
func isBrotliData(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	reader := brotli.NewReader(bytes.NewReader(data))
	// 只探测前 64 字节，避免对大文件全量解码
	probe := make([]byte, 64)
	n, err := reader.Read(probe)
	// 成功读出非零字节且无错误 → 视为 brotli
	return err == nil && n > 0
}
