package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"io/ioutil"
	"strings"
	"time"
	"strconv"
	"sort"
	"os"


	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/joho/godotenv"
)

func init() {
    err := godotenv.Load()
    if err != nil {
        log.Println(".env not loaded")
    }
}

// 城市名稱中英對照表
var cityMapping = map[string]string{
	"台北":   "Taipei",
	"新北":   "New Taipei",
	"基隆":   "Keelung",
	"台中":   "Taichung",
	"台南":   "Tainan",
	"高雄":   "Kaohsiung",
	"桃園":   "Taoyuan",
	"新竹":   "Hsinchu",
	"嘉義":   "Chiayi",
	"台東":   "Taitung",
	"花蓮":   "Hualien",
	"屏東":   "Pingtung",
	"宜蘭":   "Yilan",
	"澎湖":   "Penghu",
	"金門":   "Kinmen",
	"馬祖":   "Matsu",
	"南投":   "Nantou",
	"彰化":   "Changhua",
	"雲林":   "Yunlin",
	"高雄市":  "Kaohsiung",
	"台中市":  "Taichung",
	"台南市":  "Tainan",
	"新竹市":  "Hsinchu",
	"嘉義市":  "Chiayi City",
	"新北市":  "New Taipei City",
	"基隆市":  "Keelung City",
	"桃園市":  "Taoyuan City",
	"台東縣":  "Taitung County",
	"屏東縣":  "Pingtung County",
	"花蓮縣":  "Hualien County",
	"宜蘭縣":  "Yilan County",
	"澎湖縣":  "Penghu County",
	"金門縣":  "Kinmen County",
	"馬祖縣":  "Matsu County",
	"南投縣":  "Nantou County",
	"彰化縣":  "Changhua County",
	"雲林縣":  "Yunlin County",
	"新竹縣": "Hsinchu County",
}

func translateCityToEnglish(city string) string {
	// 查找城市名稱是否存在於對應表中
	if translatedCity, found := cityMapping[city]; found {
		return translatedCity
	}

	// 如果找不到對應城市，返回原始城市名稱
	return city
}

// // WeatherResponse 定義天氣 API 的回應結構
type WeatherResponse struct {
	List []struct {
		Main struct {
			Temp      float64 `json:"temp"`
			FeelsLike float64 `json:"feels_like"`
			Humidity  float64 `json:"humidity"`
		} `json:"main"`
		Weather []struct {
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
		Pop  float64 `json:"pop"` // 降雨機率
		Rain struct {
			ThreeHour float64 `json:"3h"`
		} `json:"rain"`
	} `json:"list"`
}

// 定義匯率查詢結構
type ExchangeRateResponse struct {
    Rates map[string]float64 `json:"rates"`
    Base  string             `json:"base"`
}

type NewsResponse struct {
	Articles []struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		URL         string `json:"url"`
		Source      struct {
			Name string `json:"name"`
		} `json:"source"`
	} `json:"articles"`
}

type Task struct {
	Name        string
	Deadline    time.Time
	Reminder    time.Time
	IsCompleted bool
}

type WeatherReminder struct {
	City    string
	Time    string // 格式 HH:MM
	Enabled bool
}

var(
	bot *linebot.Client
	tasks = make(map[string][]Task)
	userIDs = make(map[string]bool)  // 可紀錄多使用者，使用 map 儲存避免重複
	api_weather = os.Getenv("API_WEATHER")
	api_rate = os.Getenv("API_RATE")
	api_news = os.Getenv("API_NEWS")
	api_translate = os.Getenv("TRANSLATE_KEY")
)


func main() {
	// 初始化 Line Bot
	// channel secret // channel token
	var err error
	channelSecret := os.Getenv("CHANNEL_SECRET")
	channelAccessToken := os.Getenv("CHANNEL_ACCESS_TOKEN")
	bot, err = linebot.New(channelSecret, channelAccessToken)
	if err != nil {
		log.Fatal(err)
	}

	// 設置 HTTP Handler
	http.HandleFunc("/callback", callbackHandler)

	// 啟動定時功能，城市預設為新竹
	go func() {
		for range time.Tick(1 * time.Hour) { // 若要測一分鐘發送一次改為time.Minute
			sendWeatherNotification("Hsinchu")
		}
	}()

	// 啟動伺服器
	log.Println("Line bot server started at :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}


// Line 回調處理函數，用於處理用戶的回應
func callbackHandler(w http.ResponseWriter, req *http.Request) {
	events, err := bot.ParseRequest(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, event := range events {
		// 獲取 userID
		userID := event.Source.UserID
		if _, exists := userIDs[userID]; !exists {
			userIDs[userID] = true
			log.Printf("新增 userID: %s", userID) // 打印新用戶
		}

		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if message.Text == "天氣" {
					if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("請輸入您想查詢的城市名稱，例如：「城市：新竹」")).Do(); err != nil {
						log.Print(err)
					}
				} else if message.Text == "匯率" {
					if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("請輸入您想查詢的貨幣 (例如：USD to TWD)")).Do(); err != nil {
						log.Print(err)
					}
				} else if message.Text == "開啟" {

				} else if message.Text == "新聞" {
					if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("請輸入您想查詢的新聞關鍵字，例如：「新聞：關鍵字」 查詢新聞。")).Do(); err != nil {
						log.Print(err)
					}
				} else if message.Text == "翻譯" {
					if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("請輸入您想翻譯的字句，例如：「翻譯：hello」")).Do(); err != nil {
						log.Print(err)
					}
				} else if message.Text == "任務" {
					if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("請輸入您想添加的任務，例如：「任務：2024/12/16 15:30 VLIS期末考 12小時前」")).Do(); err != nil {
						log.Print(err)
					}
				} else if message.Text == "完成" {
					if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("請輸入完成的任務，例如：「完成：VLSI期末考」")).Do(); err != nil {
						log.Print(err)
					}
				} else if strings.HasPrefix(message.Text, "城市：") {
					city := strings.TrimSpace(strings.TrimPrefix(message.Text, "城市："))
					cityEnglish := translateCityToEnglish(city)
					weatherResult := getWeather(api_weather, cityEnglish)
					if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(weatherResult)).Do(); err != nil {
						log.Print(err)
					}
				} else if strings.Contains(message.Text, "to") {
					parts := strings.Fields(message.Text)
					if len(parts) != 3 {
						if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("請輸入正確的格式，例如 'USD to TWD'。")).Do(); err != nil {
							log.Print(err)
						}
						return
					}

					fromCurrency := strings.ToUpper(parts[0])
					toCurrency := strings.ToUpper(parts[2])

					rate, err := getExchangeRate(fromCurrency, toCurrency, api_rate)
					if err != nil {
						if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(err.Error())).Do(); err != nil {
							log.Print(err)
						}
					} else {
						response := fmt.Sprintf("1 %s = %.4f %s\n", fromCurrency, rate, toCurrency)
						if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(response)).Do(); err != nil {
							log.Print(err)
						}
					}
				} else if strings.HasPrefix(message.Text, "新聞：") {
					keyword := strings.TrimSpace(strings.TrimPrefix(message.Text, "新聞："))
					flexMsg := getNews(api_news, keyword)
					if flexMsg != nil {
						if _, err := bot.ReplyMessage(event.ReplyToken, flexMsg).Do(); err != nil {
							log.Print(err)
						}
					} else {
						if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("無法獲取新聞資料，請稍後再試。")).Do(); err != nil {
							log.Print(err)
						}
					}
				} else if strings.HasPrefix(message.Text, "翻譯：") {
					translationResult := handleTranslation(message.Text)
					if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(translationResult)).Do(); err != nil {
						log.Print(err)
					}
				} else if strings.HasPrefix(message.Text, "任務：") {
					// 設定作業，格式："作業：2024/12/20 12:00 統計作業 1天前"
					response := handleTaskCreation(userID, message.Text)
					if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(response)).Do(); err != nil {
						log.Print(err)
					}
				} else if message.Text == "任務清單" {
					// 列出未完成作業
					response := listTasks(userID)
					if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(response)).Do(); err != nil {
						log.Print(err)
					}
				} else if strings.HasPrefix(message.Text, "完成：") {
					// 標記作業完成，格式："完成：統計作業"
					taskName := strings.TrimSpace(strings.TrimPrefix(message.Text, "完成："))
					response := completeTask(userID, taskName)
					if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(response)).Do(); err != nil {
						log.Print(err)
					}
				} else {
					if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(
							"請輸入\n" +
							"「天氣」查詢天氣\n" +
							"「匯率」查詢匯率\n" +
							"「新聞」查詢新聞\n" +
							"「翻譯」翻譯文字\n" +
							"「任務」添加任務\n" +
							"「完成」刪除已完成任務\n" +
							"「任務清單」查詢未完成任務",
						)).Do(); err != nil {
						log.Print(err)
					}
				}
			}
		}
	}
}

func handleTaskCreation(userID, text string) string {
	parts := strings.Fields(strings.TrimPrefix(text, "任務："))
	if len(parts) < 4 {
		return "格式錯誤，請輸入「任務：YYYY/MM/DD HH:MM 作業名稱 X小時前」。"
	}

	deadline, err := time.Parse("2006/01/02 15:04", parts[0]+" "+parts[1])
	if err != nil {
		return "日期格式錯誤，請輸入正確的格式，例如：2024/12/20 12:00。"
	}

	taskName := strings.Join(parts[2:len(parts)-1], " ")
	reminderDuration := strings.TrimSuffix(parts[len(parts)-1], "小時前")
	hoursBefore, err := strconv.Atoi(reminderDuration)
	if err != nil || hoursBefore < 0 {
		return "提醒時間格式錯誤，請輸入正確的格式，例如：12小時前。"
	}

	// 計算提醒時間
	reminderTime := deadline.Add(-time.Duration(hoursBefore) * time.Hour)
	if reminderTime.Before(time.Now()) {
		return "提醒時間已過，請設定未來的提醒時間。"
	}

	// 儲存作業
	task := Task{
		Name:         taskName,
		Deadline:     deadline,
		Reminder: 	  reminderTime,
		IsCompleted:  false,
	}
	tasks[userID] = append(tasks[userID], task)

	// Goroutine
	go scheduleReminder(userID, task)

	return fmt.Sprintf("成功新增任務：%s\n截止時間：%s\n提醒時間：%s",
		taskName, deadline.Format("2006/01/02 15:04"), reminderTime.Format("2006/01/02 15:04"))
}

func scheduleReminder(userID string, task Task) {
	duration := time.Until(task.Reminder)
	if duration <= 0 {
		log.Printf("提醒時間已過，無法設置提醒：%s", task.Name)
		return
	}

	time.Sleep(duration)

	// 發送提醒
	if !task.IsCompleted {
		message := fmt.Sprintf("提醒：任務「%s」的截止時間是：%s。請記得完成！",
			task.Name, task.Deadline.Format("2006/01/02 15:04"))
		if _, err := bot.PushMessage(userID, linebot.NewTextMessage(message)).Do(); err != nil {
			log.Print(err)
		}
	}
}

func listTasks(userID string) string {
	tasks, exists := tasks[userID]
	if !exists || len(tasks) == 0 {
		return "目前沒有任何未完成的任務。"
	}

	var pendingTasks []Task
	for _, task := range tasks {
		if !task.IsCompleted && task.Deadline.After(time.Now()) {
			pendingTasks = append(pendingTasks, task)
		}
	}

	if len(pendingTasks) == 0 {
		return "目前沒有任何未完成的任務。"
	}

	sort.Slice(pendingTasks, func(i, j int) bool {
		return pendingTasks[i].Deadline.Before(pendingTasks[j].Deadline)
	})

	// output
	var response strings.Builder
	response.WriteString("未完成任務清單：\n")
	for _, task := range pendingTasks {
		response.WriteString(fmt.Sprintf(
			"- %s\n截止時間：%s\n提醒時間：%s\n",
			task.Name, task.Deadline.Format("2006/01/02 15:04"),
			task.Reminder.Format("2006/01/02 15:04"),
		))
	}

	return response.String()
}

func completeTask(userID, taskName string) string {
	tasks, exists := tasks[userID]
	if !exists {
		return "目前沒有任何任務可完成。"
	}

	taskName = strings.TrimSpace(strings.ToLower(taskName))
	for i, task := range tasks {
		if strings.ToLower(task.Name) == taskName && !task.IsCompleted {
			tasks[i].IsCompleted = true
			return fmt.Sprintf("成功標記任務「%s」為已完成！", task.Name)
		}
	}

	return fmt.Sprintf("未找到名稱為「%s」的未完成任務。", taskName)
}


// 獲取即時天氣資料
func getWeather(apiKey, city string) string {
	// 設置天氣 API 請求 URL
	url := "https://api.openweathermap.org/data/2.5/forecast?q=" + city + "&appid=" + apiKey + "&units=metric"
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return "無法獲取天氣資料，請檢查城市名稱是否正確。"
	}
	defer resp.Body.Close()

	var weather WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		return "解析天氣資料失敗"
	}

	return fmt.Sprintf("城市: %s\n天氣狀況: %s\n當前氣溫: %.1f°C\n體感溫度: %.1f°C\n濕度: %v\n過去三小時降雨量: %.2f mm\n降雨機率: %.2f",
	city, weather.List[0].Weather[0].Description, weather.List[0].Main.Temp, weather.List[0].Main.FeelsLike, weather.List[0].Main.Humidity,
	weather.List[0].Rain.ThreeHour, weather.List[0].Pop)
}

// 查詢即時匯率的函數
func getExchangeRate(fromCurrency, toCurrency, apiKey string) (float64, error) {
	// 建立 API 請求的 URL
	url := fmt.Sprintf("https://v6.exchangerate-api.com/v6/%s/latest/%s", apiKey, fromCurrency)

	// 發送 GET 請求
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("無法連接到匯率 API: %v", err)
	}
	defer resp.Body.Close()

	// 檢查是否成功獲得回應
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API 回應錯誤: %s", resp.Status)
	}

	// 定義回應數據結構
    type ExchangeRateResponse struct {
        ConversionRates map[string]float64 `json:"conversion_rates"`
    }

	// 解析 JSON 回應資料
	var data ExchangeRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("解析匯率資料失敗: %v", err)
	}

	// 查找目標貨幣的匯率
	if rate, exists := data.ConversionRates[toCurrency]; exists {
        return rate, nil
    }

	// 如果沒有找到匯率資料，返回錯誤
	return 0, fmt.Errorf("無法找到 %s 到 %s 的匯率", fromCurrency, toCurrency)
}

//查詢新聞並回傳flex message
func getNews(apiKey, keyword string) *linebot.FlexMessage {
	url := fmt.Sprintf("https://newsapi.org/v2/everything?q=%s&sortBy=popularity&apiKey=%s", keyword, apiKey)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println("Unable to fetch news:", err)
		return nil
	}
	defer resp.Body.Close()

	var news NewsResponse
	if err := json.NewDecoder(resp.Body).Decode(&news); err != nil {
		log.Println("Failed to parse news data:", err)
		return nil
	}

	if len(news.Articles) == 0 {
		return nil
	}

	// Building the Flex message components
	var contents []linebot.FlexComponent

	for i, article := range news.Articles {
		if i >= 5 {
			break // Only show the first 5 articles
		}

		// Title Component
		contents = append(contents, &linebot.TextComponent{
			Type:  "text",
			Text:  article.Title,
			Size:  "lg", // Large size for the title
			Wrap:  true,
		})

		// Source Component
		contents = append(contents, &linebot.TextComponent{
			Type:  "text",
			Text:  fmt.Sprintf("Source: %s", article.Source.Name),
			Size:  "sm",  // Small size for the source text
			Wrap:  true,
		})

		// URL Link Component
		contents = append(contents, &linebot.TextComponent{
			Type:  "text",
			Text:  fmt.Sprintf("Read More: %s", article.URL),
			Size:  "sm",  // Small size for the link
			Wrap:  true,
			Color: "#0066CC",  // Link color
			Action: &linebot.URIAction{
				Label: "Click Here",
				URI:   article.URL,
			},
		})
	}

	// Bubble Header
	header := &linebot.TextComponent{
		Type:   "text",
		Text:   "Relative Trending News",
		Size:   "xl",  // Extra large size for the header
		Align:  "center",
	}

	// Bubble Body
	body := &linebot.BoxComponent{
		Type:   "box",
		Layout: "vertical",
		Contents: contents,
	}

	// Wrap header in a BoxComponent
	headerBox := &linebot.BoxComponent{
		Type:   "box",
		Layout: "horizontal",
		Contents: []linebot.FlexComponent{header},  // Wrap header in BoxComponent
	}

	// Flex Bubble
	bubble := &linebot.BubbleContainer{
		Type:   "bubble",
		Header: headerBox,
		Body:   body,
	}

	// Returning Flex Message with Bubble
	return linebot.NewFlexMessage("News Search Results", bubble)
}

func sendWeatherNotification(city string) {
	weather := getWeather(api_weather, city)
	for userID := range userIDs {
		if _, err := bot.PushMessage(userID, linebot.NewTextMessage(weather)).Do(); err != nil {
			log.Printf("發送訊息給 userID %s 失敗: %v", userID, err)
		} else {
			log.Printf("成功發送訊息給 userID: %s", userID)
		}
	}
}


////////////////////////////////
func handleTranslation(input string) string {
	// 檢查是否以 "翻譯：" 開頭
	if !strings.HasPrefix(input, "翻譯：") {
		return "未提供翻譯請求。"
	}

	// 提取需要翻譯的文字
	stringToTranslate := strings.TrimPrefix(input, "翻譯：")

	// 設定 Translator API 詳細資料
	transKey := api_translate
	transURL := "https://api.cognitive.microsofttranslator.com/translate"
	params := "?api-version=3.0&to=zh-Hant"
	fullURL := transURL + params
	region := os.Getenv("TRANSLATE_REGION")


	headers := map[string]string{
		"Ocp-Apim-Subscription-Key":     transKey,
		"Content-type":                  "application/json",
		"Ocp-Apim-Subscription-Region":  region,
	}

	// 準備請求體
	body := []map[string]string{{"text": stringToTranslate}}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return fmt.Sprintf("JSON 序列化錯誤：%v", err)
	}

	// 建立 HTTP 請求
	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return fmt.Sprintf("請求建立錯誤：%v", err)
	}

	// 設置請求頭
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 發送請求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("HTTP 請求錯誤：%v", err)
	}
	defer resp.Body.Close()

	// 讀取回應
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("回應讀取錯誤：%v", err)
	}

	// 定義回應結構
	var response []struct {
		DetectedLanguage struct {
			Language string `json:"language"`
		} `json:"detectedLanguage"`
		Translations []struct {
			Text string `json:"text"`
		} `json:"translations"`
	}

	// 解析 JSON 回應
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return fmt.Sprintf("JSON 解析錯誤：%v", err)
	}

	// 提取翻譯結果
	if len(response) > 0 {
		detectedLanguage := response[0].DetectedLanguage.Language
		var translations []string
		for _, t := range response[0].Translations {
			translations = append(translations, t.Text)
		}
		return fmt.Sprintf("檢測到的語言：%s\n翻譯結果：%s", detectedLanguage, strings.Join(translations, ", "))
	}

	return "未收到有效的翻譯回應。"
}


