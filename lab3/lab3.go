package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func Calculator(w http.ResponseWriter, r *http.Request) {
	// TODO: implement a calculator
	parts := strings.Split(r.URL.Path[1:], "/")
	// fmt.Println(parts) // Print parts for debugging
	if len(parts) != 3 {
		fmt.Fprintf(w, "Error!")
		// http.Error(w, "Error!", http.StatusBadRequest)
		return
	}

	operation := parts[0]
	num1, err1 := strconv.Atoi(parts[1])
	num2, err2 := strconv.Atoi(parts[2])

	if err1 != nil || err2 != nil {
		fmt.Fprintf(w, "Error!")
		// http.Error(w, "Error!", http.StatusBadRequest)
		return
	}

	var result int
	var remainder int
	var output string

	switch operation {
	case "add":
		result = num1 + num2
		output = fmt.Sprintf("%d + %d = %d", num1, num2, result)
	case "sub":
		result = num1 - num2
		output = fmt.Sprintf("%d - %d = %d", num1, num2, result)
	case "mul":
		result = num1 * num2
		output = fmt.Sprintf("%d * %d = %d", num1, num2, result)
	case "div":
		if num2 == 0 {
			fmt.Fprintf(w, "Error!")
			// http.Error(w, "Error!", http.StatusBadRequest)
			return
		}
		result = num1 / num2
		remainder = num1 % num2
		output = fmt.Sprintf("%d / %d = %d, remainder %d", num1, num2, result, remainder)
	default:
		fmt.Fprintf(w, "Error!")
		// http.Error(w, "Error!", http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "%s", output)
}

func main() {
	http.HandleFunc("/", Calculator)
	log.Fatal(http.ListenAndServe(":8083", nil))
}
