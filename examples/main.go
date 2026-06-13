package main

import (
	"fmt"
	"os"

	"github.com/cyberspacesec/har-skills"
)

func main() {

	harFilePath := "./data/www.google.com.har"

	// 用法一：解析HAR文件的字节
	harFileBytes, err := os.ReadFile(harFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	harFile, err := har.ParseHar(harFileBytes)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(harFile)

	// 用法二：给定HAR文件的路径解析它
	harFile002, err := har.ParseHarFile(harFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(harFile002)

}
