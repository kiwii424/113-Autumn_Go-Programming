package main

import (
	"fmt"
	"math/big"
	// "strconv"
	"syscall/js"
)

func CheckPrime(this js.Value, args []js.Value) interface{} {
	// Get the value from the input field
    str := js.Global().Get("document").Call("getElementById", "value").Get("value").String()

    // Convert the value to a big integer
    num, ok := new(big.Int).SetString(str, 10)
    if !ok {
        js.Global().Get("document").Call("getElementById", "answer").Set("innerText", "Invalid input")
        return nil
    }

    // Check if the number is prime
    isPrime := num.ProbablyPrime(0)

    // Update the answer element based on the result
    if isPrime {
        js.Global().Get("document").Call("getElementById", "answer").Set("innerText", "It's prime")
    } else {
        js.Global().Get("document").Call("getElementById", "answer").Set("innerText", "It's not prime")
    }

    return nil
}

func registerCallbacks() {
	// Register the CheckPrime function
    js.Global().Set("CheckPrime", js.FuncOf(CheckPrime))
}

func main() {
	fmt.Println("Golang main function executed")
	registerCallbacks()

	//need block the main thread forever
	select {}
}