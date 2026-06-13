package main

import (
	"fmt"
	"os"

	"github.com/cyberspacesec/har-skills"
)

func main() {
	// 示例1: 解析现有HAR文件并使用过滤功能
	fmt.Println("====== 示例1: 解析和过滤 ======")
	harFilePath := "../../data/www.google.com.har"
	harFile, err := har.ParseHarFile(harFilePath)
	if err != nil {
		fmt.Println("解析HAR文件失败:", err)
		return
	}

	// 过滤所有POST请求
	postRequests := harFile.FindByMethod("POST")
	fmt.Printf("找到 %d 个POST请求\n", postRequests.Count())

	// 过滤所有图片请求
	imageRequests := harFile.Filter(har.FilterOptions{
		ContentType: "image/",
	})
	fmt.Printf("找到 %d 个图片请求\n", imageRequests.Count())

	// 过滤所有慢请求（超过500ms）
	slowRequests := harFile.FindSlowRequests(500)
	fmt.Printf("找到 %d 个慢请求（>500ms）\n", slowRequests.Count())

	// 查找错误请求
	errorRequests := harFile.FindErrors()
	fmt.Printf("找到 %d 个错误请求\n", errorRequests.Count())

	// 示例2: 将过滤结果转换为CSV格式
	fmt.Println("\n====== 示例2: 转换功能 ======")
	if slowRequests.Count() > 0 {
		options := har.DefaultConvertOptions()
		options.IncludeTimings = true

		csvData, err := slowRequests.ToHar().Convert(har.FormatCSV, options)
		if err != nil {
			fmt.Println("转换为CSV失败:", err)
		} else {
			fmt.Println("慢请求的CSV格式:")
			fmt.Println(csvData)
		}

		// 转换为Markdown
		mdData, _ := slowRequests.ToHar().Convert(har.FormatMarkdown, options)
		fmt.Println("慢请求的Markdown格式(部分显示):")
		lines := splitLines(mdData)
		if len(lines) > 5 {
			fmt.Println(lines[0])
			fmt.Println(lines[1])
			fmt.Println(lines[2])
			fmt.Println("... [更多行被省略] ...")
		} else {
			fmt.Println(mdData)
		}
	}

	// 示例3: 创建新的HAR文件
	fmt.Println("\n====== 示例3: 创建HAR文件 ======")
	newHar := har.NewHar()
	newHar.SetCreator("go-har-example", "1.0")

	// 添加页面
	page := newHar.AddPage("page1", "示例页面")
	page.SetPageTimings(100, 300)

	// 添加请求/响应条目
	entry := newHar.AddEntry("GET", "https://example.com/api/data", "HTTP/1.1", "page1")
	entry.AddRequestHeader("Accept", "application/json")
	entry.AddRequestHeader("User-Agent", "go-har/1.0")

	entry.SetResponseStatus(200, "OK")
	entry.AddResponseHeader("Content-Type", "application/json")
	entry.SetResponseContent(1024, "application/json")
	entry.SetTimings(10, 20, 30, 5, 50, 30, 25)

	// 保存为文件
	newHarPath := "./generated.har"
	err = newHar.SaveToFile(newHarPath, true)
	if err != nil {
		fmt.Println("保存HAR文件失败:", err)
	} else {
		fmt.Printf("成功创建并保存HAR文件到 %s\n", newHarPath)
	}

	// 读取保存的文件并验证
	generatedHar, err := har.ParseHarFile(newHarPath)
	if err != nil {
		fmt.Println("无法解析生成的HAR文件:", err)
	} else {
		fmt.Println("成功读取生成的HAR文件")
		fmt.Printf("页面数量: %d, 请求条目数量: %d\n",
			len(generatedHar.Log.Pages), len(generatedHar.Log.Entries))
	}

	// 清理测试文件
	os.Remove(newHarPath)
}

// 辅助函数：按行分割字符串
func splitLines(s string) []string {
	var lines []string
	var line string
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, line)
			line = ""
		} else {
			line += string(r)
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}
