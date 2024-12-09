package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gocolly/colly"
)

func main() {
	// 定義 flag
	maxComments := flag.Int("max", 10, "Max number of comments to show")
	flag.Parse()

	// 檢查是否有多餘或錯誤的 flag
	if len(flag.Args()) > 0 {
		fmt.Println("Usage of:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// 初始化 Colly Collector
	c := colly.NewCollector()

	// 用來存取留言資訊
	type Comment struct {
		Name    string
		Content string
		Time    string
	}
	comments := []Comment{}

	// 抓取留言的選擇器
	c.OnHTML("div.push", func(e *colly.HTMLElement) {
		name := e.ChildText("span.push-userid")
		content := e.ChildText("span.push-content")
		time := e.ChildText("span.push-ipdatetime")
		if name != "" && content != "" && time != "" {
			// 去掉留言前的冒號和空白
			trimmedContent := content[2:]
			comments = append(comments, Comment{
				Name:    name,
				Content: trimmedContent,
				Time:    time,
			})
		}
	})

	// 訪問指定的 PTT 網址
	url := "https://www.ptt.cc/bbs/joke/M.1481217639.A.4DF.html"
	err := c.Visit(url)
	if err != nil {
		fmt.Println("Error visiting URL:", err)
		os.Exit(1)
	}

	// 印出留言資訊，最多 *maxComments 筆
	for i, comment := range comments {
		if i >= *maxComments {
			break
		}
		fmt.Printf("%d. 名字：%s，留言: %s，時間： %s\n", i+1, comment.Name, comment.Content, comment.Time)
	}
}