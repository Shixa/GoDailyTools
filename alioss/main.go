package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// OSSConfig 存储OSS配置信息
type OSSConfig struct {
	Bucket   string `json:"bucket"`
	ID       string `json:"id"`
	Secret   string `json:"secret"`
	EndPoint string `json:"endPoint"`
}

// OSSClient 封装OSS客户端
type OSSClient struct {
	client *oss.Client
	bucket *oss.Bucket
	config OSSConfig
}

// NewOSSClient 创建一个新的OSS客户端
func NewOSSClient() (*OSSClient, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %v", err)
	}

	client, err := oss.New(config.EndPoint, config.ID, config.Secret)
	if err != nil {
		return nil, fmt.Errorf("创建OSS客户端失败: %v", err)
	}

	bucket, err := client.Bucket(config.Bucket)
	if err != nil {
		return nil, fmt.Errorf("获取Bucket失败: %v", err)
	}

	return &OSSClient{
		client: client,
		bucket: bucket,
		config: config,
	}, nil
}

// 从用户目录下的.oss-config文件加载配置
func loadConfig() (OSSConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return OSSConfig{}, fmt.Errorf("获取用户目录失败: %v", err)
	}

	configPath := filepath.Join(homeDir, ".oss-config")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return OSSConfig{}, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config OSSConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return OSSConfig{}, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return config, nil
}

// UploadFile 上传本地文件到OSS
func (c *OSSClient) UploadFile(localPath, ossPath string) error {
	// 如果ossPath为空，使用本地文件名
	if ossPath == "" {
		ossPath = filepath.Base(localPath)
	}

	// 标准化OSS路径，去除前导斜杠
	ossPath = strings.TrimPrefix(ossPath, "/")

	// 使用中文名时需要指定Content-Disposition
	options := []oss.Option{
		oss.ContentDisposition(fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(localPath))),
	}

	err := c.bucket.PutObjectFromFile(ossPath, localPath, options...)
	if err != nil {
		return fmt.Errorf("上传文件失败: %v", err)
	}

	return nil
}

// DownloadFile 从OSS下载文件到本地
func (c *OSSClient) DownloadFile(ossPath, localPath string) error {
	// 标准化OSS路径，去除前导斜杠
	ossPath = strings.TrimPrefix(ossPath, "/")

	// 如果本地路径是目录，则使用OSS文件名
	fileInfo, err := os.Stat(localPath)
	if err == nil && fileInfo.IsDir() {
		localPath = filepath.Join(localPath, filepath.Base(ossPath))
	}

	err = c.bucket.GetObjectToFile(ossPath, localPath)
	if err != nil {
		return fmt.Errorf("下载文件失败: %v", err)
	}

	return nil
}

// ListFiles 列出指定前缀的文件
func (c *OSSClient) ListFiles(prefix string) ([]string, error) {
	// 标准化前缀，去除前导斜杠
	prefix = strings.TrimPrefix(prefix, "/")

	marker := ""
	var files []string

	for {
		lsRes, err := c.bucket.ListObjects(oss.Marker(marker), oss.Prefix(prefix))
		if err != nil {
			return nil, fmt.Errorf("列举文件失败: %v", err)
		}

		for _, object := range lsRes.Objects {
			files = append(files, object.Key)
		}

		if lsRes.IsTruncated {
			marker = lsRes.NextMarker
		} else {
			break
		}
	}

	return files, nil
}

// GetSignedURL 获取文件的临时访问URL
func (c *OSSClient) GetSignedURL(ossPath string, expireTime time.Duration) (string, error) {
	// 标准化OSS路径，去除前导斜杠
	ossPath = strings.TrimPrefix(ossPath, "/")

	// 生成签名URL
	signedURL, err := c.bucket.SignURL(ossPath, oss.HTTPGet, int64(expireTime.Seconds()))
	if err != nil {
		return "", fmt.Errorf("生成签名URL失败: %v", err)
	}

	return signedURL, nil
}

// DeleteFile 删除OSS上的文件
func (c *OSSClient) DeleteFile(ossPath string) error {
	// 标准化OSS路径，去除前导斜杠
	ossPath = strings.TrimPrefix(ossPath, "/")

	err := c.bucket.DeleteObject(ossPath)
	if err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}

	return nil
}

func printUsage() {
	fmt.Println("阿里云OSS工具使用方法:")
	fmt.Println("  上传文件: alioss upload <本地文件路径> [OSS路径]")
	fmt.Println("  下载文件: alioss download <OSS路径> <本地保存路径>")
	fmt.Println("  列出文件: alioss list [前缀]")
	fmt.Println("  删除文件: alioss delete <OSS路径>")
	fmt.Println("  获取临时URL: alioss url <OSS路径> [过期时间(秒)，默认3600]")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	client, err := NewOSSClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "upload":
		if len(os.Args) < 3 {
			fmt.Println("错误: 缺少本地文件路径")
			printUsage()
			os.Exit(1)
		}
		localPath := os.Args[2]
		ossPath := ""
		if len(os.Args) > 3 {
			ossPath = os.Args[3]
		}
		if err := client.UploadFile(localPath, ossPath); err != nil {
			fmt.Fprintf(os.Stderr, "上传失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("文件上传成功!")

	case "download":
		if len(os.Args) < 4 {
			fmt.Println("错误: 请提供OSS文件路径和本地保存路径")
			printUsage()
			os.Exit(1)
		}
		ossPath := os.Args[2]
		localPath := os.Args[3]
		if err := client.DownloadFile(ossPath, localPath); err != nil {
			fmt.Fprintf(os.Stderr, "下载失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("文件下载成功!")

	case "list":
		prefix := ""
		if len(os.Args) > 2 {
			prefix = os.Args[2]
		}
		files, err := client.ListFiles(prefix)
		if err != nil {
			fmt.Fprintf(os.Stderr, "列举文件失败: %v\n", err)
			os.Exit(1)
		}
		if len(files) == 0 {
			fmt.Println("未找到文件")
		} else {
			fmt.Println("文件列表:")
			for _, file := range files {
				fmt.Println("  " + file)
			}
		}

	case "delete":
		if len(os.Args) < 3 {
			fmt.Println("错误: 请提供OSS文件路径")
			printUsage()
			os.Exit(1)
		}
		ossPath := os.Args[2]
		if err := client.DeleteFile(ossPath); err != nil {
			fmt.Fprintf(os.Stderr, "删除失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("文件删除成功!")

	case "url":
		if len(os.Args) < 3 {
			fmt.Println("错误: 请提供OSS文件路径")
			printUsage()
			os.Exit(1)
		}
		ossPath := os.Args[2]
		expireTime := 3600 * time.Second // 默认1小时
		if len(os.Args) > 3 {
			expireSeconds := 0
			if _, err := fmt.Sscanf(os.Args[3], "%d", &expireSeconds); err == nil && expireSeconds > 0 {
				expireTime = time.Duration(expireSeconds) * time.Second
			}
		}
		url, err := client.GetSignedURL(ossPath, expireTime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取URL失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("临时访问URL:")
		fmt.Println(url)

	default:
		fmt.Printf("未知命令: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}
