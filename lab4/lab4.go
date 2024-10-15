package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"path/filepath"
)

// TODO: Create a struct to hold the data sent to the template
type Calculation struct {
    Expression string
    Result     string
}

func gcd(a, b int) int {
    for b != 0 {
        a, b = b, a % b
    }
    return a
}

func lcm(a, b int) int {
    return a * b / gcd(a, b)
}

func Calculator(w http.ResponseWriter, r *http.Request) {
    op := r.URL.Query().Get("op")
    num1, err1 := strconv.Atoi(r.URL.Query().Get("num1"))
    num2, err2 := strconv.Atoi(r.URL.Query().Get("num2"))

    if err1 != nil || err2 != nil {
        renderErrorPage(w, "Invalid input")
        return
    }

    var result int
    var expression string
    switch op {
    case "add":
        result = num1 + num2
        expression = fmt.Sprintf("%d + %d", num1, num2)
    case "sub":
        result = num1 - num2
        expression = fmt.Sprintf("%d - %d", num1, num2)
    case "mul":
        result = num1 * num2
        expression = fmt.Sprintf("%d * %d", num1, num2)
    case "div":
        if num2 == 0 {
            renderErrorPage(w, "Cannot divide by zero")
            return
        }
        result = num1 / num2
        expression = fmt.Sprintf("%d / %d", num1, num2)
    case "gcd":
        result = gcd(num1, num2)
        expression = fmt.Sprintf("GCD(%d, %d)", num1, num2)
    case "lcm":
        result = lcm(num1, num2)
        expression = fmt.Sprintf("LCM(%d, %d)", num1, num2)
    default:
        renderErrorPage(w, "Invalid operation")
        return
    }

    calc := Calculation{
        Expression: expression,
        Result:     strconv.Itoa(result),
    }

    tmplPath, _ := filepath.Abs("index.html")
    tmpl, err := template.ParseFiles(tmplPath)
    if err != nil {
        renderErrorPage(w, err.Error())
        return
    }

    err = tmpl.Execute(w, calc)
    if err != nil {
        renderErrorPage(w, err.Error())
    }
}

func renderErrorPage(w http.ResponseWriter, errorMessage string) {
    tmplPath, _ := filepath.Abs("error.html")
    tmpl, err := template.ParseFiles(tmplPath)
    if err != nil {
        http.Error(w, "An unexpected error occurred", http.StatusInternalServerError)
        return
    }

    err = tmpl.Execute(w, errorMessage)
    if err != nil {
        http.Error(w, "An unexpected error occurred", http.StatusInternalServerError)
    }
}

func main() {
	http.HandleFunc("/", Calculator)
	log.Fatal(http.ListenAndServe(":8084", nil))
}

