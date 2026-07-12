package har

// 本测试文件为 decode.go 补充覆盖率，使用独立的 Cov 前缀函数名，
// 不与 decode_test.go 重复或冲突。
//
// 说明：decode.go 中以下分支为“结构性不可达”的防御性代码，无法在不修改
// 源码的前提下覆盖（已通过穷举验证，详见各 TestCov*Unreachable* 用例）：
//
//   - decompressIfNeeded 行 186-189（zlib.NewReader 返回错误的分支）：
//     isDeflateData 仅接受 {0x78, 0x01/0x5e/0x9c/0xda} 这种“完全合法且
//     不带 FDICT 预设字典标志”的 zlib 头；而 zlib.NewReader 对这些头永远
//     返回 nil。两者无交集，故该 err != nil 分支不可达。
//
//   - CompressContent 行 279-282 / 283-286 / 291-294 / 295-298
//     （gzip/zlib Writer.Write 与 Writer.Close 返回错误的分支）：
//     CompressContent 内部固定使用 bytes.Buffer 作为底层 Writer，
//     bytes.Buffer.Write 永不返回错误，因此 gzip/zlib 的 Write 与 Close
//     在此处也永不返回错误，这些 err != nil 分支不可达。
//
// 本文件对可达行为做最大化覆盖，并对上述不可达分支以实验方式再次确认其
// “实际行为”（即对应调用确实不会返回错误），以满足“匹配源码实际行为”的约束。

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- CompressContent 成功路径（gzip） ---

func TestCovCompressContentGzipRoundTrip(t *testing.T) {
	original := []byte("coverage round trip gzip payload 1234567890")

	compressed, err := CompressContent(original, "gzip")
	require.NoError(t, err)
	require.NotEmpty(t, compressed)
	assert.True(t, isGzipData(compressed), "压缩结果应具备 gzip magic bytes")

	// 解压回去验证内容一致
	out, err := DecompressByEncoding(compressed, "gzip")
	require.NoError(t, err)
	assert.Equal(t, original, out)
}

func TestCovCompressContentGzipBinaryData(t *testing.T) {
	// 二进制数据（含 0x00 字节）也应正常压缩
	original := bytes.Repeat([]byte{0x00, 0xFF, 0x7F, 0x80, 0x01}, 1000)

	compressed, err := CompressContent(original, "gzip")
	require.NoError(t, err)

	out, err := DecompressByEncoding(compressed, "gzip")
	require.NoError(t, err)
	assert.Equal(t, original, out)
}

func TestCovCompressContentGzipLargeData(t *testing.T) {
	// 超过内部缓冲的大数据量
	original := bytes.Repeat([]byte("A"), 1<<20) // 1 MiB

	compressed, err := CompressContent(original, "gzip")
	require.NoError(t, err)
	assert.Less(t, len(compressed), len(original), "gzip 应能压缩重复数据")

	out, err := DecompressByEncoding(compressed, "gzip")
	require.NoError(t, err)
	assert.Equal(t, original, out)
}

// --- CompressContent 成功路径（deflate） ---

func TestCovCompressContentDeflateRoundTrip(t *testing.T) {
	original := []byte("coverage round trip deflate payload abcdef")

	compressed, err := CompressContent(original, "deflate")
	require.NoError(t, err)
	require.NotEmpty(t, compressed)
	assert.True(t, isDeflateData(compressed), "压缩结果应具备 zlib 头字节")

	out, err := DecompressByEncoding(compressed, "deflate")
	require.NoError(t, err)
	assert.Equal(t, original, out)
}

func TestCovCompressContentDeflateBinaryData(t *testing.T) {
	original := bytes.Repeat([]byte{0x10, 0x20, 0x30, 0x40}, 500)

	compressed, err := CompressContent(original, "deflate")
	require.NoError(t, err)

	out, err := DecompressByEncoding(compressed, "deflate")
	require.NoError(t, err)
	assert.Equal(t, original, out)
}

// --- CompressContent 大小写 / 空白 / 空数据 ---

func TestCovCompressContentCaseInsensitiveGzip(t *testing.T) {
	for _, enc := range []string{"GZIP", "Gzip", "  gzip  ", "\tgzip\n"} {
		compressed, err := CompressContent([]byte("x"), enc)
		require.NoError(t, err, "编码 %q 应能识别为 gzip", enc)
		assert.True(t, isGzipData(compressed))
	}
}

func TestCovCompressContentCaseInsensitiveDeflate(t *testing.T) {
	for _, enc := range []string{"DEFLATE", "Deflate", "  deflate  "} {
		compressed, err := CompressContent([]byte("x"), enc)
		require.NoError(t, err, "编码 %q 应能识别为 deflate", enc)
		assert.True(t, isDeflateData(compressed))
	}
}

func TestCovCompressContentEmptyGzip(t *testing.T) {
	out, err := CompressContent([]byte{}, "gzip")
	require.NoError(t, err)
	assert.Empty(t, out)
}

func TestCovCompressContentEmptyDeflate(t *testing.T) {
	out, err := CompressContent([]byte{}, "deflate")
	require.NoError(t, err)
	assert.Empty(t, out)
}

func TestCovCompressContentEmptyBrZstd(t *testing.T) {
	// 空数据优先返回，不会进入 br/zstd 的不支持分支
	for _, enc := range []string{"br", "zstd", "unknown"} {
		out, err := CompressContent([]byte{}, enc)
		require.NoError(t, err, "空数据 + 编码 %q 应直接返回空", enc)
		assert.Empty(t, out)
	}
}

// --- CompressContent brotli/zstd 正向往返（现已支持） ---

func TestCovCompressContentBrUnsupported(t *testing.T) {
	original := []byte("brotli coverage round-trip 数据")
	compressed, err := CompressContent(original, "br")
	require.NoError(t, err)
	decompressed, err := DecompressByEncoding(compressed, "br")
	require.NoError(t, err)
	assert.Equal(t, original, decompressed)
}

func TestCovCompressContentZstdUnsupported(t *testing.T) {
	original := []byte("zstd coverage round-trip 数据")
	compressed, err := CompressContent(original, "zstd")
	require.NoError(t, err)
	decompressed, err := DecompressByEncoding(compressed, "zstd")
	require.NoError(t, err)
	assert.Equal(t, original, decompressed)
}

func TestCovCompressContentUnknownEncoding(t *testing.T) {
	for _, enc := range []string{"snappy", "lz4", "identity", ""} {
		_, err := CompressContent([]byte("data"), enc)
		require.Error(t, err, "编码 %q 应不支持", enc)
		he, ok := err.(*HarError)
		require.True(t, ok)
		assert.Equal(t, ErrCodeUnsupported, he.Code)
	}
}

func TestCovCompressContentUnknownEncodingMessage(t *testing.T) {
	_, err := CompressContent([]byte("data"), "snappy")
	require.Error(t, err)
	he, ok := err.(*HarError)
	require.True(t, ok)
	assert.Contains(t, he.Message, "不支持")
	assert.Contains(t, he.Message, "snappy")
}

// --- decompressIfNeeded 各路径 ---

func TestCovDecompressIfNeededEmpty(t *testing.T) {
	out, err := decompressIfNeeded(nil, "text/plain")
	require.NoError(t, err)
	assert.Nil(t, out)

	out2, err2 := decompressIfNeeded([]byte{}, "text/plain")
	require.NoError(t, err2)
	assert.Empty(t, out2)
}

func TestCovDecompressIfNeededGzipSuccess(t *testing.T) {
	original := []byte("cov gzip success payload")
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, err := w.Write(original)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	out, err := decompressIfNeeded(buf.Bytes(), "application/gzip")
	require.NoError(t, err)
	assert.Equal(t, original, out)
}

func TestCovDecompressIfNeededDeflateSuccess(t *testing.T) {
	original := []byte("cov deflate success payload")
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	_, err := w.Write(original)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	out, err := decompressIfNeeded(buf.Bytes(), "application/deflate")
	require.NoError(t, err)
	assert.Equal(t, original, out)
}

func TestCovDecompressIfNeededPlainPassthrough(t *testing.T) {
	original := []byte("just some plain text, no compression at all")
	out, err := decompressIfNeeded(original, "text/plain")
	require.NoError(t, err)
	assert.Equal(t, original, out)
}

func TestCovDecompressIfNeededShortData(t *testing.T) {
	// 长度 < 2，既不是 gzip 也不是 deflate，直接返回
	short := []byte{0x1f}
	out, err := decompressIfNeeded(short, "text/plain")
	require.NoError(t, err)
	assert.Equal(t, short, out)
}

func TestCovDecompressIfNeededGzipReadAllError(t *testing.T) {
	// 有效 gzip 头但被截断 -> io.ReadAll 失败（覆盖 176-179）
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, err := w.Write([]byte("payload"))
	require.NoError(t, err)
	require.NoError(t, w.Close())
	truncated := buf.Bytes()[:buf.Len()-4]

	_, err = decompressIfNeeded(truncated, "text/plain")
	require.Error(t, err)
	he, ok := err.(*HarError)
	require.True(t, ok)
	assert.Equal(t, ErrCodeInvalidFormat, he.Code)
	assert.Contains(t, he.Message, "gzip")
}

func TestCovDecompressIfNeededDeflateReadAllError(t *testing.T) {
	// 有效 zlib 头但被截断 -> io.ReadAll 失败（覆盖 193-196）
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	_, err := w.Write([]byte("payload"))
	require.NoError(t, err)
	require.NoError(t, w.Close())
	truncated := buf.Bytes()[:buf.Len()-4]

	_, err = decompressIfNeeded(truncated, "text/plain")
	require.Error(t, err)
	he, ok := err.(*HarError)
	require.True(t, ok)
	assert.Equal(t, ErrCodeInvalidFormat, he.Code)
	assert.Contains(t, he.Message, "deflate")
}

// --- 通过 Content.DecodeContent 触发 decompressIfNeeded 的 deflate 路径 ---

func TestCovDecodeContentDeflateViaBase64(t *testing.T) {
	original := []byte("deflate via base64 decode path")
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	_, err := w.Write(original)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	content := &Content{
		Size:     buf.Len(),
		MimeType: "text/plain",
		Text:     encoded,
		Encoding: "base64",
	}

	data, err := content.DecodeContent()
	require.NoError(t, err)
	assert.Equal(t, original, data)
}

// --- 不可达分支的实际行为验证（匹配源码实际行为） ---
//
// 以下用例通过实验再次确认：在当前源码（固定使用 bytes.Buffer / bytes.NewReader）
// 的前提下，这些 err != nil 分支确实不会被触发。这不是“覆盖”该分支，而是
// “证明该分支不可达”，作为对任务中“匹配源码实际行为”约束的交代。

// TestCovUnreachableGzipWriteNeverFails 证明：对 bytes.Buffer 调用
// gzip.Writer.Write 永不返回错误，因此 CompressContent 行 279-282 不可达。
func TestCovUnreachableGzipWriteNeverFails(t *testing.T) {
	for _, payload := range [][]byte{
		[]byte("small"),
		bytes.Repeat([]byte("A"), 1<<20),
		bytes.Repeat([]byte{0x00, 0xFF}, 4096),
	} {
		var buf bytes.Buffer
		w := gzip.NewWriter(&buf)
		_, err := w.Write(payload)
		assert.NoError(t, err, "gzip.Writer.Write 对 bytes.Buffer 不应失败; payload len=%d", len(payload))
		assert.NoError(t, w.Close(), "gzip.Writer.Close 对 bytes.Buffer 不应失败; payload len=%d", len(payload))
	}
}

// TestCovUnreachableDeflateWriteNeverFails 证明：对 bytes.Buffer 调用
// zlib.Writer.Write / Close 永不返回错误，因此 CompressContent 行 291-298 不可达。
func TestCovUnreachableDeflateWriteNeverFails(t *testing.T) {
	for _, payload := range [][]byte{
		[]byte("small"),
		bytes.Repeat([]byte("B"), 1<<20),
		bytes.Repeat([]byte{0x00, 0xFF}, 4096),
	} {
		var buf bytes.Buffer
		w := zlib.NewWriter(&buf)
		_, err := w.Write(payload)
		assert.NoError(t, err, "zlib.Writer.Write 对 bytes.Buffer 不应失败; payload len=%d", len(payload))
		assert.NoError(t, w.Close(), "zlib.Writer.Close 对 bytes.Buffer 不应失败; payload len=%d", len(payload))
	}
}

// TestCovUnreachableZlibNewReaderNeverFailsForIsDeflateDataInputs 证明：
// 对所有 isDeflateData 接受的 zlib 头字节组合，zlib.NewReader 永不返回错误，
// 因此 decompressIfNeeded 行 186-189 不可达。
func TestCovUnreachableZlibNewReaderNeverFailsForIsDeflateDataInputs(t *testing.T) {
	acceptedFlgs := []byte{0x01, 0x5e, 0x9c, 0xda}
	for _, flg := range acceptedFlgs {
		for _, suffix := range [][]byte{
			{},                               // 仅 2 字节头
			{0xFF, 0xFF, 0xFF, 0xFF},         // 头 + 垃圾
			{0x00, 0x00, 0x00, 0x00},         // 头 + 零
			bytes.Repeat([]byte{0xAB}, 1024), // 头 + 大量垃圾
		} {
			data := append([]byte{0x78, flg}, suffix...)
			require.True(t, isDeflateData(data), "前置条件: isDeflateData 应为 true (flg=0x%02x)", flg)

			r, err := zlib.NewReader(bytes.NewReader(data))
			assert.NoError(t, err, "zlib.NewReader 对 isDeflateData 接受的输入 (0x78,0x%02x) 不应失败", flg)
			if err == nil {
				// 消耗 reader 以释放资源；忽略 ReadAll 的错误（本用例只验证 NewReader）
				_, _ = readAll(r)
				_ = r.Close()
			}
		}
	}
}

// TestCovUnreachableBruteForceConfirms 证明：对所有 data[0]==0x78 的 2 字节头，
// 唯一能让 zlib.NewReader 失败的第二字节都不在 isDeflateData 的接受集合内，
// 即两者无交集 -> 行 186-189 不可达。
func TestCovUnreachableBruteForceConfirms(t *testing.T) {
	intersection := 0
	for b := 0; b < 256; b++ {
		data := []byte{0x78, byte(b)}
		_, err := zlib.NewReader(bytes.NewReader(data))
		if err != nil {
			// NewReader 失败
			if isDeflateData(data) {
				// 同时被 isDeflateData 接受 -> 交集
				intersection++
			}
		}
	}
	assert.Equal(t, 0, intersection,
		"isDeflateData 接受集合与 zlib.NewReader 失败集合的交集应为空（实际=%d），"+
			"因此 decompressIfNeeded 行 186-189 不可达", intersection)
}

// --- 借助 readAll 辅助（已在 decode_test.go 中定义）避免重复导入 io ---
// readAll 定义于 decode_test.go，此处直接复用，无需重新声明。
