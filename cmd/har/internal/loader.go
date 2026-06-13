package internal

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	har "github.com/cyberspacesec/har-skills"
)

// LoadHar 从 --file 标志或 stdin 加载 HAR 文件
// 如果 --file 为空且 stdin 有数据，则从 stdin 读取
func LoadHar(cmd *cobra.Command, args []string) *har.Har {
	filePath, _ := cmd.Flags().GetString("file")

	if filePath == "-" || (filePath == "" && hasStdinData()) {
		return LoadHarFromStdin()
	}

	if filePath == "" {
		fmt.Fprintln(os.Stderr, "错误: 未指定HAR文件。请使用 -f <文件路径> 或通过管道传入数据")
		os.Exit(1)
	}

	h, err := LoadHarFromPath(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法加载HAR文件 '%s': %v\n", filePath, err)
		os.Exit(1)
	}
	return h
}

// LoadHarFromPath 从指定路径加载 HAR 文件（支持 gzip 自动检测）
func LoadHarFromPath(path string) (*har.Har, error) {
	return har.ParseHarFileAuto(path)
}

// LoadHarFromStdin 从标准输入加载 HAR 数据
func LoadHarFromStdin() *har.Har {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法从stdin读取数据: %v\n", err)
		os.Exit(1)
	}

	h, err := har.ParseHar(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法解析HAR数据: %v\n", err)
		os.Exit(1)
	}
	return h
}

// LoadHarFromArg 从命令行参数加载 HAR 文件（用于 diff、merge 等多文件命令）
func LoadHarFromArg(path string) *har.Har {
	h, err := LoadHarFromPath(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法加载HAR文件 '%s': %v\n", path, err)
		os.Exit(1)
	}
	return h
}

// hasStdinData 检查 stdin 是否有可用数据
func hasStdinData() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}
