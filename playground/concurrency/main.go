package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Result struct {
	UserID int
	Err    error
}

func main() {

	// results := make(chan string)
	// go worker(1, results)
	// go worker(2, results)
	// go worker(3, results)

	// for i := 0; i < 3; i++ {
	// 	fmt.Println(<-results)
	// }

	// runWithWaitGroup()
	// fmt.Println("all workers finished")

	raceExample()

	// runWorkerPool()

}

func runWithWaitGroup() {
	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()
			fmt.Println("worker started", id+1)
			time.Sleep(time.Duration(id+1) * time.Second)
			fmt.Println("worker ended", id+1)
		}(i)
	}
	wg.Wait()

}

func raceExample() {

	var (
		count int
		wg    sync.WaitGroup
		mu    sync.Mutex
	)
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			count++
			mu.Unlock()
		}()
	}

	wg.Wait()
	fmt.Println(count)
}

func worker(ctx context.Context, id int, jobs <-chan int, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		results <- Result{
			UserID: job,
			Err:    sendNotification(ctx, job),
		}
	}
}

func runWorkerPool() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()
	jobs := make(chan int, 8)
	results := make(chan Result)

	var wg sync.WaitGroup

	wg.Add(3)

	go worker(ctx, 1, jobs, results, &wg)
	go worker(ctx, 2, jobs, results, &wg)
	go worker(ctx, 3, jobs, results, &wg)

	for job := 1; job <= 8; job++ {
		jobs <- job
	}

	close(jobs)

	// wg.Wait()
	go func() {
		wg.Wait()
		close(results)
	}()

	succeeded := 0
	failed := 0
	for result := range results {
		if result.Err != nil {
			failed++
			fmt.Println("notification failed for user", result.UserID, result.Err)
			continue
		}
		succeeded++
		fmt.Println("notification succeeded for user", result.UserID)

	}

	fmt.Println("jobs compleated failed", failed, "success jobs", succeeded)
}

func sendNotification(ctx context.Context, userID int) error {
	select {
	case <-ctx.Done():
		fmt.Println("notification cancelled for user", userID)
		return ctx.Err()
	case <-time.After(3 * time.Second):
		fmt.Println("notification send to user", userID)
		return nil

	}
}
