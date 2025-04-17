package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// findFiles 查找指定目录下所有名为name的文件。
func findFiles(dir, name string) ([]string, error) {
	var matches []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == name {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

// printLinesWithContents 打开文件并打印包含任意一个内容项的行。
func printLinesWithContents(filePath string, contents []string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		for _, content := range contents {
			if strings.Contains(line, content) {
				fmt.Println(line)
				break // 一旦找到匹配的内容就跳出循环，避免重复打印
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func main() {
	dirToSearch := flag.String("dir", "", "Directory to search for files")
	fileNameToFind := flag.String("file", "product_extend.info", "File name to search for")
	contentToFind := flag.String("content", "", "Comma-separated list of contents to search within the files")

	flag.Parse()

	if *dirToSearch == "" || *contentToFind == "" {
		fmt.Fprintln(os.Stderr, "Error: -dir and -content flags are required.")
		flag.Usage()
		os.Exit(1)
	}

	// 将逗号分隔的内容转换为切片
	contents := strings.Split(*contentToFind, ",")

	files, err := findFiles(*dirToSearch, *fileNameToFind)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(files) > 0 {
		fmt.Println("Found files and lines with any of the specified contents:")
		for _, file := range files {
			fmt.Printf("In file: %s\n", file)
			if err := printLinesWithContents(file, contents); err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", file, err)
			}
		}
	} else {
		fmt.Println("No files found with the name:", *fileNameToFind)
	}
}
