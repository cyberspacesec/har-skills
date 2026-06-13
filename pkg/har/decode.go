package har

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
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
	var errors []string

	for i, entry := range h.Log.Entries {
		data, err := entry.DecodeContent()
		if err != nil {
			errors = append(errors, fmt.Sprintf("条目[%d]: %v", i, err))
			results[i] = nil
			continue
		}
		results[i] = data
	}

	if len(errors) > 0 {
		return results, NewHarError(ErrCodeInvalidFormat,
			fmt.Sprintf("解码过程中有%d个错误", len(errors)), nil)
	}

	return results, nil
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
	}

	return data, nil
}

// DecompressByEncoding 根据Content-Encoding值解压数据
//
// 支持的编码: "gzip", "deflate"
// 不支持的编码（标准库不支持）: "br" (brotli), "zstd"
// 多重编码（如 "gzip, deflate"）不支持，会返回错误。
func DecompressByEncoding(data []byte, encoding string) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	enc := strings.ToLower(strings.TrimSpace(encoding))

	// 检查多重编码（包含逗号分隔的多个值）
	if strings.Contains(enc, ",") {
		return nil, NewUnsupportedError(
			fmt.Sprintf("不支持多重编码: %q，请逐层解压", encoding))
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
		return nil, NewUnsupportedError(
			"brotli (br) 解压不被Go标准库支持，请引入第三方库如 github.com/andybalholm/brotli")
	case "zstd":
		return nil, NewUnsupportedError(
			"zstd 解压不被Go标准库支持，请引入第三方库如 github.com/klauspost/compress/zstd")
	default:
		return nil, NewUnsupportedError(
			fmt.Sprintf("不支持的Content-Encoding: %q", encoding))
	}
}

// CompressContent 使用指定的编码方式压缩数据
//
// 支持的编码: "gzip", "deflate"
// 不支持的编码（标准库不支持）: "br" (brotli), "zstd"
func CompressContent(data []byte, encoding string) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	enc := strings.ToLower(strings.TrimSpace(encoding))

	switch enc {
	case "gzip":
		var buf bytes.Buffer
		writer := gzip.NewWriter(&buf)
		if _, err := writer.Write(data); err != nil {
			return nil, NewHarError(ErrCodeInvalidFormat,
				fmt.Sprintf("gzip压缩失败: %v", err), err)
		}
		if err := writer.Close(); err != nil {
			return nil, NewHarError(ErrCodeInvalidFormat,
				fmt.Sprintf("gzip压缩失败: %v", err), err)
		}
		return buf.Bytes(), nil
	case "deflate":
		var buf bytes.Buffer
		writer := zlib.NewWriter(&buf)
		if _, err := writer.Write(data); err != nil {
			return nil, NewHarError(ErrCodeInvalidFormat,
				fmt.Sprintf("deflate压缩失败: %v", err), err)
		}
		if err := writer.Close(); err != nil {
			return nil, NewHarError(ErrCodeInvalidFormat,
				fmt.Sprintf("deflate压缩失败: %v", err), err)
		}
		return buf.Bytes(), nil
	case "br":
		return nil, NewUnsupportedError(
			"brotli (br) 压缩不被Go标准库支持，请引入第三方库如 github.com/andybalholm/brotli")
	case "zstd":
		return nil, NewUnsupportedError(
			"zstd 压缩不被Go标准库支持，请引入第三方库如 github.com/klauspost/compress/zstd")
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
