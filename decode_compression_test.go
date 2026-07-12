package har

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrotliRoundtrip(t *testing.T) {
	original := []byte("Hello Brotli! This is a test payload with some 非ASCII 字符 for good measure." + repeatString("X", 200))

	// 压缩
	compressed, err := CompressContent(original, "br")
	require.NoError(t, err, "brotli 压缩不应失败")
	assert.LessOrEqual(t, len(compressed), len(original), "brotli 压缩后应小于或等于原文（短文本可能无压缩收益）")

	// 解压
	decompressed, err := DecompressByEncoding(compressed, "br")
	require.NoError(t, err, "brotli 解压不应失败")
	assert.Equal(t, original, decompressed, "往返应还原原文")

	// 验证 isBrotliData 能识别
	assert.True(t, isBrotliData(compressed), "brotli 压缩数据应被正确识别")
	assert.False(t, isBrotliData(original), "普通文本不应被误判为 brotli")
}

func TestZstdRoundtrip(t *testing.T) {
	original := []byte("Zstandard compression test. 长文本测试：" + repeatString("A", 1000))

	compressed, err := CompressContent(original, "zstd")
	require.NoError(t, err, "zstd 压缩不应失败")
	assert.Less(t, len(compressed), len(original), "zstd 压缩后应更小")

	decompressed, err := DecompressByEncoding(compressed, "zstd")
	require.NoError(t, err, "zstd 解压不应失败")
	assert.Equal(t, original, decompressed, "往返应还原")

	// 验证 isZstdData 魔数检测
	assert.True(t, isZstdData(compressed), "zstd 压缩数据应有魔数")
	assert.False(t, isZstdData(original), "普通数据无 zstd 魔数")
}

func TestMultiEncodingDecompress(t *testing.T) {
	// 场景：gzip(deflate(original)) —— 两层包裹
	original := []byte("Multi-layer encoding test 数据")

	// 先 deflate 再 gzip 包裹（模拟 Content-Encoding: gzip, deflate）
	deflated, err := CompressContent(original, "deflate")
	require.NoError(t, err)

	gzipWrapped, err := CompressContent(deflated, "gzip")
	require.NoError(t, err)

	// DecompressByEncoding 多重编码：先解 gzip 再解 deflate
	decompressed, err := DecompressByEncoding(gzipWrapped, "gzip, deflate")
	require.NoError(t, err, "多重编码解压应成功")
	assert.Equal(t, original, decompressed, "两层解压后还原")

	// 三层：br(gzip(deflate))
	brGzipDeflate, err := CompressContent(gzipWrapped, "br")
	require.NoError(t, err)

	decompressed3, err := DecompressByEncoding(brGzipDeflate, "br, gzip, deflate")
	require.NoError(t, err)
	assert.Equal(t, original, decompressed3)
}

func TestDecompressByEncodingErrors(t *testing.T) {
	// 空数据
	result, err := DecompressByEncoding(nil, "gzip")
	assert.NoError(t, err)
	assert.Nil(t, result)

	// 不支持的编码
	_, err = DecompressByEncoding([]byte("test"), "lz4")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不支持")

	// 损坏的压缩数据
	corruptedGzip := []byte{0x1f, 0x8b, 0x00, 0x00} // gzip magic 但截断
	_, err = DecompressByEncoding(corruptedGzip, "gzip")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gzip")

	corruptedZstd := []byte{0x28, 0xb5, 0x2f, 0xfd, 0x00} // zstd magic 但截断
	_, err = DecompressByEncoding(corruptedZstd, "zstd")
	assert.Error(t, err)
}

func TestDecompressIfNeededWithBrotli(t *testing.T) {
	original := []byte("Auto-detect brotli 测试")
	compressed, err := CompressContent(original, "br")
	require.NoError(t, err)

	// decompressIfNeeded 应靠魔数嗅探解压
	decompressed, err := decompressIfNeeded(compressed, "application/octet-stream")
	require.NoError(t, err)
	assert.Equal(t, original, decompressed)
}

func TestDecompressIfNeededWithZstd(t *testing.T) {
	original := []byte("Auto-detect zstd 测试")
	compressed, err := CompressContent(original, "zstd")
	require.NoError(t, err)

	decompressed, err := decompressIfNeeded(compressed, "")
	require.NoError(t, err)
	assert.Equal(t, original, decompressed)
}

func TestDecodeContentWithBrotliBase64(t *testing.T) {
	// 模拟 HAR 里 Content.Encoding="base64" + 实际被 brotli 压缩的 body
	original := []byte("HAR body content compressed with brotli then base64 encoded")
	brCompressed, err := CompressContent(original, "br")
	require.NoError(t, err)

	// 手造一个 Content 结构
	content := &Content{
		Text:     encodeBase64(brCompressed),
		Encoding: "base64",
		MimeType: "application/json",
		Size:     len(original),
	}

	// DecodeContent 应：先 base64 解码 → 再 brotli 解压
	decoded, err := content.DecodeContent()
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestDecodeContentWithZstdBase64(t *testing.T) {
	original := []byte("zstd then base64")
	zstdCompressed, err := CompressContent(original, "zstd")
	require.NoError(t, err)

	content := &Content{
		Text:     encodeBase64(zstdCompressed),
		Encoding: "base64",
		MimeType: "",
		Size:     len(original),
	}

	decoded, err := content.DecodeContent()
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

// 辅助函数
func repeatString(s string, n int) string {
	var b bytes.Buffer
	for i := 0; i < n; i++ {
		b.WriteString(s)
	}
	return b.String()
}

func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
