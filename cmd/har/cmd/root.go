package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string

	// 构建时注入的版本信息（由 GoReleaser 通过 -ldflags 注入）
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "har",
	Short: "HAR (HTTP Archive) 文件分析工具",
	Long: `HAR Skills CLI — AI Agent Skill for HAR File Analysis

支持 HAR 文件的解析、过滤、统计、安全审计、性能评分、
数据脱敏、请求转换、差异比较、合并拆分、导出等多种操作。

示例:
  har -f capture.har info              # 查看HAR文件概要
  har -f capture.har find "api/users"  # 搜索请求
  har -f capture.har security          # 安全审计
  har -f capture.har performance       # 性能评分
  har -f capture.har export curl       # 导出为curl命令`,
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// 自定义版本输出格式
	rootCmd.SetVersionTemplate(`HAR Skills {{.Version}}
commit:  {{.Annotations.commit}}
date:    {{.Annotations.date}}
`)
	rootCmd.Annotations = map[string]string{
		"commit": commit,
		"date":   date,
	}

	rootCmd.PersistentFlags().StringP("file", "f", "", "HAR文件路径 (使用 - 读取stdin)")
	rootCmd.PersistentFlags().String("format", "text", "输出格式 (text, json, csv, yaml)")
	rootCmd.PersistentFlags().StringP("output", "o", "", "输出文件路径")
	rootCmd.PersistentFlags().Bool("no-header", false, "在text/csv输出中隐藏表头")

	_ = viper.BindPFlag("file", rootCmd.PersistentFlags().Lookup("file"))
	_ = viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	_ = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	_ = viper.BindPFlag("no-header", rootCmd.PersistentFlags().Lookup("no-header"))

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认 $HOME/.har.yaml)")
}

func initConfig() {
	viper.SetEnvPrefix("HAR")
	viper.AutomaticEnv()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".har")
		viper.AddConfigPath("$HOME")
		viper.AddConfigPath(".")
	}

	_ = viper.ReadInConfig()
}
