package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	results := make(chan string, 1)
	done := make(chan struct{})


	go worker(ctx, results,done)

	select {
	case value := <-results:
		fmt.Println("RESULTS:", value)
	case <-ctx.Done():
		fmt.Println("timeout")
	}

	<-done
}

func worker(ctx context.Context, results chan<- string,done chan <- struct{}) {
	defer close(done)
	select {
	case <-ctx.Done():
		fmt.Println("worker canceled:", ctx.Err())
		return
	case <-time.After(1 * time.Second):
		results <- "done"
	}
}
