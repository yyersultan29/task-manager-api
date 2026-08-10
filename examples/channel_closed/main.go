package main

import "fmt"

func main() {
	messages := make(chan string, 3)

	messages <- "hello"
	messages <- "go"
	messages <- "channels"
	close(messages)

	for message := range messages {
		fmt.Println(message)
	}	
}
