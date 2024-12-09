package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	// Define the max flag with a default value of 10
	max := flag.Int("max", 10, "Max number of comments to show")
	flag.Parse()

	// Create a custom HTTP client
	client := &http.Client{}

	// Step 1: Confirm Age Verification
	req, err := http.NewRequest("POST", "https://www.ptt.cc/ask/over18", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Form = map[string][]string{
		"yes": {"yes"},
	}

	// Perform the age verification request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error in age verification:", err)
		return
	}
	defer resp.Body.Close()

	// Step 2: Fetch the PTT post content
	postURL := "https://www.ptt.cc/bbs/joke/M.1481217639.A.4DF.html"
	req, err = http.NewRequest("GET", postURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err = client.Do(req)
	if err != nil {
		fmt.Println("Error fetching post:", err)
		return
	}
	defer resp.Body.Close()

	// Load the response body into GoQuery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println("Error loading document:", err)
		return
	}

	// Counter to track the number of comments collected
	count := 0

	// Extract comments
	doc.Find(".push").Each(func(i int, s *goquery.Selection) {
		if count < *max {
			username := s.Find(".push-userid").Text()
			content := s.Find(".push-content").Text()
			time := s.Find(".push-ipdatetime").Text()
			fmt.Printf("%d. 名字：%s，留言%s，時間：%s", count+1, username, content, time)
			count++
		}
	})

	// Check if no comments were printed
	if count == 0 {
		fmt.Println("No comments found or failed to fetch comments.")
		os.Exit(1)
	}
}
