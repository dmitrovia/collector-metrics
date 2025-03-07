package main

import (
	"fmt"
	"os"
)

//nolint:lll
func main() {
	fmt.Println("message")

	temp := 5
	temp1 := 10

	if temp > temp1 {
		os.Exit(0) // want "calling os.Exit in main package is not allowed"
	}

	call()

	os.Exit(0) // want "calling os.Exit in main package is not allowed"
}

func call() {
	os.Exit(0)
}
