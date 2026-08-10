package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup

	delays := []time.Duration{
		time.Second,
		2 * time.Second,
		3 * time.Second,
	}

	for index,delay := range delays {
		wg.Add(1)
		go worker(index + 1,delay,&wg)
		wg.Wait()

	}
	fmt.Println("all workers finished")

}

func worker(id int, timer time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Println("worker", id, "started")
	time.Sleep(timer)
	fmt.Println("worker", id, "finished")
}
