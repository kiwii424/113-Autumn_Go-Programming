package main

import (
    "flag"
    "fmt"
    "os"

    "github.com/gocolly/colly"
)

func main() {
    max := flag.Int("max", 10, "Max number of comments to show")
    flag.Parse()

    // 檢查 flag 是否正確
    if *max <= 0 {
        fmt.Println("Usage of", os.Args[0])
        flag.PrintDefaults()
        return
    }

    // 初始化 Colly 爬蟲
    c := colly.NewCollector(
        colly.AllowedDomains("www.ptt.cc"),
    )

    comments := []string{}
    var count int

    // 查找並抓取留言
    c.OnHTML(".push", func(e *colly.HTMLElement) {
        if count >= *max {
            return
        }
        id := e.ChildText(".push-userid")
        content := e.ChildText(".push-content")
        time := e.ChildText(".push-ipdatetime")

        comment := fmt.Sprintf("名字：%s，留言%s，時間： %s", id, content, time)
        comments = append(comments, comment)
        count++
    })

    // 爬取指定頁面
    c.Visit("https://www.ptt.cc/bbs/joke/M.1481217639.A.4DF.html")

    // 輸出抓取到的留言
    for i, comment := range comments {
        fmt.Printf("%d. %s\n", i+1, comment)
    }
}
