package main

import (
	"flag"
	"fmt"
	"github.com/gocolly/colly"
	"os"
)

func main() {
	// Define the max flag with a default value of 10
	max := flag.Int("max", 10, "Max number of comments to show")

	flag.Parse()

	// Create a new Colly collector with custom User-Agent and allowed cookies
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.AllowURLRevisit(),
		colly.MaxDepth(1),
	)

	// Enable cookies support
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Referer", "https://www.ptt.cc/")
		// fmt.Println("Visiting:", r.URL)
	})

	// Counter to track the number of comments collected
	count := 0

	// Find comments on the page
	c.OnHTML(".push", func(e *colly.HTMLElement) {
		if count < *max {
			username := e.ChildText(".push-userid")
			content := e.ChildText(".push-content")
			time := e.ChildText(".push-ipdatetime")

			fmt.Printf("%d. 名字：%s，留言%s，時間： %s\n", count+1, username, content, time)
			count++
		}
	})

	// Error handling
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Error:", err)
		os.Exit(1)
	})

	// Visit the PTT post
	err := c.Visit("https://www.ptt.cc/bbs/joke/M.1481217639.A.4DF.html")
	if err != nil {
		fmt.Println("Visit failed:", err)
	}
}
