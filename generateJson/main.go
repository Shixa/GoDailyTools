package main

import (
	"encoding/json"
	"fmt"
	"github.com/tealeg/xlsx"
	"log"
	"os"
)

// App 结构体表示单个应用的信息
type App struct {
	Icon   string `json:"icon"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	PicURL string `json:"pic_url"`
}

// RecommendApp 结构体表示推荐应用的数据
type RecommendApp struct {
	Languages map[string][]App `json:"recommendApp"`
}

func findSheetByName(file *xlsx.File, name string) (*xlsx.Sheet, error) {
	for _, sheet := range file.Sheets {
		if sheet.Name == name {
			return sheet, nil
		}
	}
	return nil, fmt.Errorf("sheet with name '%s' not found", name)
}

func main() {
	// 指定要读取的工作表名称
	productSheetName := "产品信息"
	filePath := "./generate_json.xlsx"

	// 打开Excel文件
	xlFile, err := xlsx.OpenFile(filePath)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}

	// 查找指定名称的工作表
	productSheet, err := findSheetByName(xlFile, productSheetName)
	if err != nil {
		log.Fatal(err)
	}

	// 初始化 RecommendApp 结构体
	recommendApp := RecommendApp{
		Languages: make(map[string][]App),
	}
	app := App{
		Icon:   "",
		Name:   "",
		URL:    "",
		PicURL: "",
	}
	domain := "https://config.aiseesoft.com/android-unlocker/win/uninstall/recappicon/"
	for _, row := range productSheet.Rows[1:] {
		if len(row.Cells) >= 2 { // 确保有足够的单元格
			language := row.Cells[0].String()
			app.Icon = row.Cells[1].String()
			app.Name = row.Cells[2].String()
			app.URL = row.Cells[3].String()
			//app.PicURL = row.Cells[4].String()
			app.PicURL = domain + app.Name
			// 判断键是否存在
			if _, exists := recommendApp.Languages[language]; exists {
				recommendApp.Languages[language] = append(recommendApp.Languages[language], app)
			} else {
				recommendApp.Languages[language] = []App{app}
			}
		}
	}
	// 将数据编码为JSON格式
	jsonData, err := json.MarshalIndent(recommendApp, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	// 指定输出文件路径
	outputFilePath := "recommendapp.json"

	// 将 JSON 数据写入新的文件
	err = os.WriteFile(outputFilePath, jsonData, 0644)
	if err != nil {
		log.Fatalf("Error writing JSON to file: %v", err)
	}

	fmt.Printf("Parsed data written to %s\n", outputFilePath)
}
