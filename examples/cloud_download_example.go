package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// 云盘下载示例程序
// 演示如何配置和使用 RSS 自动下载的云盘功能

type Config struct {
	BaseURL string
	Token   string
}

type DownloadToolsResponse struct {
	Code int `json:"code"`
	Data struct {
		Tools           []DownloadTool `json:"tools"`
		RecommendedTool string         `json:"recommended_tool"`
	} `json:"data"`
}

type DownloadTool struct {
	Name         string   `json:"name"`
	DisplayName  string   `json:"display_name"`
	Type         string   `json:"type"`
	IsConfigured bool     `json:"is_configured"`
	IsAvailable  bool     `json:"is_available"`
	Categories   []string `json:"categories"`
	Description  string   `json:"description"`
}

type AutoDownloadRule struct {
	Name            string   `json:"name"`
	MustContain     string   `json:"must_contain"`
	MustNotContain  string   `json:"must_not_contain"`
	UseRegex        bool     `json:"use_regex"`
	AffectedFeeds   []string `json:"affected_feeds"`
	DestinationPath string   `json:"destination_path"`
	DownloadTool    string   `json:"download_tool"`
	DeletePolicy    string   `json:"delete_policy"`
	TorrentTempPath string   `json:"torrent_temp_path,omitempty"`
}

func main() {
	config := Config{
		BaseURL: "http://localhost:5244",
		Token:   os.Getenv("ALIST_TOKEN"),
	}

	if config.Token == "" {
		fmt.Println("请设置环境变量 ALIST_TOKEN")
		os.Exit(1)
	}

	client := &http.Client{}

	// 1. 获取可用的下载工具
	fmt.Println("=== 获取可用的下载工具 ===")
	tools, err := getDownloadTools(client, config)
	if err != nil {
		fmt.Printf("获取下载工具失败: %v\n", err)
		return
	}

	fmt.Printf("推荐工具: %s\n", tools.Data.RecommendedTool)
	fmt.Println("可用工具:")
	for _, tool := range tools.Data.Tools {
		status := "❌ 未配置"
		if tool.IsAvailable {
			status = "✅ 可用"
		}
		fmt.Printf("  - %s (%s) - %s [%s]\n",
			tool.DisplayName, tool.Type, tool.Description, status)
	}

	// 2. 找到可用的云盘工具
	var cloudTool *DownloadTool
	for _, tool := range tools.Data.Tools {
		if tool.Type == "cloud" && tool.IsAvailable {
			cloudTool = &tool
			break
		}
	}

	if cloudTool == nil {
		fmt.Println("\n⚠️  没有找到可用的云盘下载工具")
		fmt.Println("请先配置云盘存储 (PikPak, 115云盘, 或迅雷网盘)")
		return
	}

	fmt.Printf("\n✅ 选择云盘工具: %s\n", cloudTool.DisplayName)

	// 3. 创建云盘自动下载规则
	fmt.Println("\n=== 创建云盘自动下载规则 ===")
	rule := AutoDownloadRule{
		Name:            "云盘自动下载示例",
		MustContain:     "1080p",
		MustNotContain:  "预告|PV|CM",
		UseRegex:        false,
		AffectedFeeds:   []string{"feed-uuid-example"},
		DestinationPath: "/downloads/auto",
		DownloadTool:    cloudTool.Name,
		DeletePolicy:    "delete_on_upload_succeed",
		TorrentTempPath: "/temp/torrents",
	}

	err = createAutoDownloadRule(client, config, rule)
	if err != nil {
		fmt.Printf("创建规则失败: %v\n", err)
		return
	}

	fmt.Println("✅ 云盘自动下载规则创建成功!")
	fmt.Printf("规则名称: %s\n", rule.Name)
	fmt.Printf("下载工具: %s\n", rule.DownloadTool)
	fmt.Printf("目标路径: %s\n", rule.DestinationPath)
	fmt.Printf("临时路径: %s\n", rule.TorrentTempPath)

	// 4. 输出使用说明
	fmt.Println("\n=== 使用说明 ===")
	fmt.Println("1. 确保已配置相应的云盘存储驱动")
	fmt.Println("2. RSS 订阅将自动使用云盘下载匹配的资源")
	fmt.Println("3. 下载完成后文件会自动转存到目标目录")
	fmt.Println("4. 可以在任务管理界面查看下载进度")

	fmt.Println("\n🎉 云盘下载配置完成!")
}

func getDownloadTools(client *http.Client, config Config) (*DownloadToolsResponse, error) {
	req, err := http.NewRequest("GET", config.BaseURL+"/api/admin/rss/download-tools", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+config.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result DownloadToolsResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func createAutoDownloadRule(client *http.Client, config Config, rule AutoDownloadRule) error {
	data, err := json.Marshal(rule)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", config.BaseURL+"/api/admin/rss/rules", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+config.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}