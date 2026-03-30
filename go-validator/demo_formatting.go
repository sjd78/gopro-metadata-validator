package main

import "fmt"

// DemoFormatting shows lefthook in action - this will be auto-formatted
func DemoFormatting(x int, y int) {
	result := x + y
	message := "Sum is"
	fmt.Println(message, result)
}
