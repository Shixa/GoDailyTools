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

// UploadOptions 上传选项
type UploadOptions struct {
	ExcludePatterns []string // 排除的文件或目录模式
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
func (c *OSSClient) UploadFile(localPath, ossPath string, options *UploadOptions) error {
	// 检查是否为目录
	fileInfo, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("读取文件信息失败: %v", err)
	}

	// 如果是目录，则递归上传目录中的文件
	if fileInfo.IsDir() {
		return c.UploadDirectory(localPath, ossPath, options)
	}

	// 如果ossPath为空，使用本地文件名
	if ossPath == "" {
		ossPath = filepath.Base(localPath)
	}

	// 标准化OSS路径，去除前导斜杠
	ossPath = strings.TrimPrefix(ossPath, "/")

	// 使用中文名时需要指定Content-Disposition
	ossOptions := []oss.Option{
		oss.ContentDisposition(fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(localPath))),
	}

	err = c.bucket.PutObjectFromFile(ossPath, localPath, ossOptions...)
	if err != nil {
		return fmt.Errorf("上传文件失败: %v", err)
	}

	return nil
}

// UploadDirectory 上传目录及其所有文件到OSS
func (c *OSSClient) UploadDirectory(localDirPath, ossDirPath string, options *UploadOptions) error {
	// 确保OSS路径以斜杠结尾
	if ossDirPath != "" && !strings.HasSuffix(ossDirPath, "/") {
		ossDirPath += "/"
	}

	// 标准化OSS路径，去除前导斜杠
	ossDirPath = strings.TrimPrefix(ossDirPath, "/")

	// 获取本地目录的绝对路径
	absLocalDirPath, err := filepath.Abs(localDirPath)
	if err != nil {
		return fmt.Errorf("获取绝对路径失败: %v", err)
	}

	fmt.Printf("开始上传目录: %s 到 %s\n", absLocalDirPath, ossDirPath)
	// 仅显示排除模式的数量，而不是详细列出每个模式
	if options != nil && len(options.ExcludePatterns) > 0 {
		fmt.Printf("使用 %d 个排除模式\n", len(options.ExcludePatterns))
	}

	// 统计上传文件数量和排除文件数量
	var uploadCount, excludeCount int

	err = filepath.Walk(localDirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录本身
		if info.IsDir() {
			return nil
		}

		// 获取文件的绝对路径
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("获取绝对路径失败: %v", err)
		}

		// 计算相对路径
		relPath, err := filepath.Rel(localDirPath, path)
		if err != nil {
			return fmt.Errorf("计算相对路径失败: %v", err)
		}

		// 检查文件是否被排除
		if shouldExclude(relPath, options) || shouldExclude(absPath, options) {
			excludeCount++
			return nil
		}

		// 在Windows系统上将反斜杠转换为正斜杠
		relPath = filepath.ToSlash(relPath)

		// 构建OSS上的完整路径
		ossObjectPath := ossDirPath + relPath

		// 上传文件
		ossOptions := []oss.Option{
			oss.ContentDisposition(fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(path))),
		}

		err = c.bucket.PutObjectFromFile(ossObjectPath, path, ossOptions...)
		if err != nil {
			return fmt.Errorf("上传文件 %s 失败: %v", path, err)
		}

		uploadCount++
		fmt.Printf("已上传: %s\n", ossObjectPath)

		return nil
	})

	if err != nil {
		return fmt.Errorf("上传目录失败: %v", err)
	}

	fmt.Printf("成功上传 %d 个文件到 %s", uploadCount, ossDirPath)
	if excludeCount > 0 {
		fmt.Printf("，已排除 %d 个文件", excludeCount)
	}
	fmt.Println()
	return nil
}

// shouldExclude 检查文件是否应该被排除
func shouldExclude(path string, options *UploadOptions) bool {
	if options == nil || len(options.ExcludePatterns) == 0 {
		return false
	}

	// 规范化路径分隔符为正斜杠
	normalizedPath := filepath.ToSlash(path)

	// 获取绝对路径
	absPath, err := filepath.Abs(filepath.FromSlash(path))
	if err == nil {
		absPath = filepath.ToSlash(absPath)
	}

	for _, pattern := range options.ExcludePatterns {
		// 处理带通配符的绝对路径
		if strings.HasPrefix(pattern, "/") && strings.Contains(pattern, "*") {
			// 去掉通配符，检查前缀匹配
			patternPrefix := strings.Split(pattern, "*")[0]
			if strings.HasPrefix(absPath, patternPrefix) {
				// 如果是 /* 结尾，则只匹配目录下的顶级文件/目录
				if strings.HasSuffix(pattern, "/*") {
					// 检查剩余路径是否包含更多的斜杠
					remainingPath := absPath[len(patternPrefix):]
					if !strings.Contains(remainingPath, "/") {
						return true
					}
				} else {
					// 尝试用glob模式匹配
					baseName := filepath.Base(absPath)
					patternBaseName := filepath.Base(pattern)
					if matched, _ := filepath.Match(patternBaseName, baseName); matched {
						return true
					}

					// 尝试完整路径模式匹配
					filePattern := filepath.Base(pattern)
					fileToCheck := absPath
					if strings.HasPrefix(fileToCheck, filepath.ToSlash(filepath.Dir(pattern))) {
						fileNameToCheck := filepath.Base(fileToCheck)
						matched, _ := filepath.Match(filePattern, fileNameToCheck)
						if matched {
							return true
						}
					}
				}
			}
		}

		// 处理其他绝对路径匹配
		if strings.HasPrefix(pattern, "/") {
			// 如果排除模式是绝对路径，我们需要检查完整路径
			if absPath != "" {
				if strings.HasPrefix(absPath, pattern) || absPath == pattern {
					return true
				}
			}
			continue
		}

		// 将文件夹模式标准化（确保以 / 结尾）
		if strings.HasSuffix(pattern, "/") && strings.HasPrefix(normalizedPath, pattern) {
			return true
		}

		// 处理通配符
		if strings.Contains(pattern, "*") {
			// 处理目录通配符模式 (例如: dir/*)
			if strings.HasSuffix(pattern, "/*") {
				prefix := strings.TrimSuffix(pattern, "/*")
				if strings.HasPrefix(normalizedPath, prefix+"/") && !strings.Contains(normalizedPath[len(prefix)+1:], "/") {
					return true
				}
			} else {
				matched, err := filepath.Match(pattern, normalizedPath)
				if err == nil && matched {
					return true
				}

				// 检查文件名是否匹配
				if strings.HasPrefix(pattern, "*") {
					extension := pattern[1:]
					if strings.HasSuffix(normalizedPath, extension) {
						return true
					}
				}
			}
			continue
		}

		// 精确匹配
		if normalizedPath == pattern {
			return true
		}
	}

	return false
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

	// 检查路径是否以斜杠结尾，如果是，则可能是要删除文件夹
	if strings.HasSuffix(ossPath, "/") || strings.Contains(ossPath, "*") {
		return c.DeleteDirectory(ossPath)
	}

	err := c.bucket.DeleteObject(ossPath)
	if err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}

	return nil
}

// DeleteDirectory 删除OSS上的目录（删除指定前缀的所有文件）
func (c *OSSClient) DeleteDirectory(prefix string) error {
	// 标准化OSS路径，去除前导斜杠
	prefix = strings.TrimPrefix(prefix, "/")

	// 确保前缀以斜杠结尾，表示是一个目录
	if !strings.HasSuffix(prefix, "/") && !strings.Contains(prefix, "*") {
		prefix += "/"
	}

	// 获取所有匹配前缀的文件
	files, err := c.ListFiles(prefix)
	if err != nil {
		return fmt.Errorf("获取文件列表失败: %v", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("未找到匹配的文件")
	}

	// 批量删除文件
	deleteCount := 0
	for _, file := range files {
		err := c.bucket.DeleteObject(file)
		if err != nil {
			return fmt.Errorf("删除文件 %s 失败: %v", file, err)
		}
		fmt.Printf("已删除: %s\n", file)
		deleteCount++
	}

	fmt.Printf("成功删除 %d 个文件\n", deleteCount)
	return nil
}

func printUsage() {
	fmt.Println("阿里云OSS工具使用方法:")
	fmt.Println("  上传文件/文件夹: alioss upload <本地文件或文件夹路径> [OSS路径] [--exclude 模式1,模式2,...]")
	fmt.Println("  下载文件: alioss download <OSS路径> <本地保存路径>")
	fmt.Println("  列出文件: alioss list [前缀]")
	fmt.Println("  删除文件/文件夹: alioss delete <OSS路径或前缀>")
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
		if len(os.Args) > 3 && !strings.HasPrefix(os.Args[3], "--") {
			ossPath = os.Args[3]
		}

		// 处理排除选项
		uploadOptions := &UploadOptions{}
		for i := 3; i < len(os.Args); i++ {
			if os.Args[i] == "--exclude" && i+1 < len(os.Args) {
				excludePatterns := strings.Split(os.Args[i+1], ",")
				for j, pattern := range excludePatterns {
					excludePatterns[j] = strings.TrimSpace(pattern)
				}
				uploadOptions.ExcludePatterns = excludePatterns
				i++
			}
		}

		if err := client.UploadFile(localPath, ossPath, uploadOptions); err != nil {
			fmt.Fprintf(os.Stderr, "上传失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("上传完成!")

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
			fmt.Println("错误: 请提供OSS文件路径或前缀")
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
