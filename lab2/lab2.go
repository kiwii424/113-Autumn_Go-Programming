package main

import "fmt"

func main() {
	var n int64

	fmt.Print("Enter a number: ")
	fmt.Scanln(&n)

	result := Sum(n)
	fmt.Println(result)
}

func Sum(n int64) string {
	var sum int64 = 0
	var expression string

	for i := int64(1); i <= n; i++ {
		if i % 7 != 0 {
			sum += i
			if expression == "" {
				expression = fmt.Sprintf("%d", i)
			} else {
				expression = fmt.Sprintf("%s+%d", expression, i)
			}
		}
	}
	return fmt.Sprintf("%s=%d", expression, sum)
}