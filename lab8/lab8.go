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

	// Check for invalid flags
	if len(os.Args) > 1 && os.Args[1][0] == '-' && os.Args[1] != "-max" {
		fmt.Printf("flag provided but not defined: %s\n", os.Args[1])
		fmt.Println("Usage of:", os.Args[0])
		flag.PrintDefaults()
		os.Exit(2)
	}

	// Create a new Colly collector
	c := colly.NewCollector()

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
