package main

import (
    "bufio"
    "log"
    "net/http"
    "os"
    "strings"

	"github.com/gorilla/websocket"
	"github.com/reactivex/rxgo/v2"
)

type client chan<- string // an outgoing message channel

var (
	entering      = make(chan client)
	leaving       = make(chan client)
	messages      = make(chan rxgo.Item) // all incoming client messages
	ObservableMsg = rxgo.FromChannel(messages)
)

func broadcaster() {
	clients := make(map[client]bool) // all connected clients
	MessageBroadcast := ObservableMsg.Observe()
	for {
		select {
		case msg := <-MessageBroadcast:
			// Broadcast incoming message to all
			// clients' outgoing message channels.
			for cli := range clients {
				cli <- msg.V.(string)
			}

		case cli := <-entering:
			clients[cli] = true

		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}

func clientWriter(conn *websocket.Conn, ch <-chan string) {
	for msg := range ch {
		conn.WriteMessage(1, []byte(msg))
	}
}

func wshandle(w http.ResponseWriter, r *http.Request) {
	upgrader := &websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	ch := make(chan string) // outgoing client messages
	go clientWriter(conn, ch)

	who := conn.RemoteAddr().String()
	ch <- "你是 " + who + "\n"
	messages <- rxgo.Of(who + " 來到了現場" + "\n")
	entering <- ch

	defer func() {
		log.Println("disconnect !!")
		leaving <- ch
		messages <- rxgo.Of(who + " 離開了" + "\n")
		conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		messages <- rxgo.Of(who + " 表示: " + string(msg))
	}
}

func readFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var result []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func InitObservable() {
	// 讀取髒話和敏感人物名單
	swearWords, err := readFile("swear_word.txt")
	if err != nil {
		log.Println("Error reading swear_word.txt:", err)
		return
	}

	sensitiveNames, err := readFile("sensitive_name.txt")
	if err != nil {
		log.Println("Error reading sensitive_name.txt:", err)
		return
	}

	// 處理 ObservableMsg
	go func() {
		for msg := range ObservableMsg.Observe() {
			// 檢查訊息是否含有髒話
			if containsSwearWord(msg.V.(string), swearWords) {
				continue // 若含有髒話，跳過這條訊息
			}

			// 替換敏感人物名稱
			modifiedMsg := replaceSensitiveNames(msg.V.(string), sensitiveNames)

			// 將過濾後的訊息送到 messages channel
			messages <- rxgo.Of(modifiedMsg)
		}
	}()
}

// 檢查訊息中是否包含髒話
func containsSwearWord(msg string, swearWords []string) bool {
	for _, swearWord := range swearWords {
		if strings.Contains(strings.ToLower(msg), strings.ToLower(swearWord)) {
			return true
		}
	}
	return false
}

// 替換訊息中的敏感人物名稱
func replaceSensitiveNames(msg string, sensitiveNames []string) string {
	for _, sensitiveName := range sensitiveNames {
		// 若訊息中有敏感人物名稱，則將第二個字替換為 '*'
		if strings.Contains(msg, sensitiveName) {
			nameParts := []rune(sensitiveName)
			if len(nameParts) > 1 {
				nameParts[1] = '*'
				msg = strings.ReplaceAll(msg, sensitiveName, string(nameParts))
			}
		}
	}
	return msg
}


func main() {
	InitObservable()
	go broadcaster()
	http.HandleFunc("/wschatroom", wshandle)

	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Println("server start at :8090")
	log.Fatal(http.ListenAndServe(":8090", nil))
}