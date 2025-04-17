package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

// RecommendApp 结构体表示推荐应用的数据
type RecommendApp struct {
	Languages map[string][]App `json:"recommendApp"`
}

// App 定义应用信息
type App struct {
	Icon   string `placeholder:"Enter icon URL..."`
	Name   string `placeholder:"Enter name..."`
	URL    string `placeholder:"Enter URL..."`
	PicURL string `placeholder:"Enter picture URL..."`
}

func createEntryWithPlaceholder(placeholder string) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(placeholder)
	return entry
}

func guiStart() {
	// 创建一个新的 Fyne 应用
	a := app.New()
	w := a.NewWindow("Fyne Input Example")

	appInfo := &App{}

	// 创建输入框
	iconEntry := createEntryWithPlaceholder(appInfo.Icon)
	nameEntry := createEntryWithPlaceholder(appInfo.Name)
	urlEntry := createEntryWithPlaceholder(appInfo.URL)
	picURLEntry := createEntryWithPlaceholder(appInfo.PicURL)

	// 创建一个显示应用信息的文本对象
	appContent := canvas.NewText(fmt.Sprintf("%+v", appInfo), color.Black)
	appContent.TextStyle = fyne.TextStyle{Bold: true, Monospace: true} // 可选：设置样式

	// 创建提交按钮
	button := widget.NewButton("Submit", func() {
		// 获取输入框中的文本并更新 appInfo
		appInfo.Icon = iconEntry.Text
		appInfo.Name = nameEntry.Text
		appInfo.URL = urlEntry.Text
		appInfo.PicURL = picURLEntry.Text

		// 更新显示的应用信息
		appContent.Text = fmt.Sprintf("%+v", appInfo)

		// 后台输出文本
		fmt.Println("Input Text:", appInfo)
	})

	// 将输入框和按钮放入容器中
	content := container.NewVBox(
		widget.NewLabel("Icon:"),
		iconEntry,
		widget.NewLabel("Name:"),
		nameEntry,
		widget.NewLabel("URL:"),
		urlEntry,
		widget.NewLabel("Picture URL:"),
		picURLEntry,
		button,
		widget.NewLabel("Application Info:"),
		appContent,
	)

	// 设置窗口的内容
	w.SetContent(content)

	// 设置窗口大小
	w.Resize(fyne.NewSize(800, 600))

	// 显示窗口并运行应用
	w.ShowAndRun()
}

func main() {
	guiStart()
}
