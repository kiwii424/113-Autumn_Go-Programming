package main

import (
    "fmt"
    "sync"
    "time"
)

var doorStatus string
var handStatus string
var mu sync.Mutex  // 定義一個互斥鎖
var wg sync.WaitGroup

func hand() {
    mu.Lock()  // 進入臨界區
    handStatus = "in"
    time.Sleep(time.Millisecond * 200)
    handStatus = "out"
    mu.Unlock()  // 離開臨界區
    wg.Done()
}

func door() {
    mu.Lock()  // 進入臨界區
    doorStatus = "close"
    time.Sleep(time.Millisecond * 200)
    if handStatus == "in" {
        fmt.Println("夾到手了啦！")
    } else {
        fmt.Println("沒夾到喔！")
    }
    doorStatus = "open"
    mu.Unlock()  // 離開臨界區
    wg.Done()
}

func main() {
    for i := 0; i < 50; i++ {
        wg.Add(2)
        go door()
        go hand()
        wg.Wait()
        time.Sleep(time.Millisecond * 200)
    }
}



