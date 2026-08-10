package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type User struct {
	ID   int
	Name string
}

type Result struct {
	UserID int
	User   User
	Err    error
}

func main() {

	const workersCount = 3
	userIDs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	ctx,cancel := context.WithTimeout(context.Background(),3 * time.Second)

	defer cancel()

	results := fetchAll(ctx,userIDs,workersCount);

	for _,result := range results {
		if result.Err != nil {
			fmt.Println("failed user", result.UserID, result.Err)
			continue
		}
		fmt.Println("fetched user", result.User)
	}
	
}

func fetchAll(ctx context.Context,userIDs []int,workersCount int) []Result {

	var wg sync.WaitGroup
	jobs := make(chan int, len(userIDs))
	results := make(chan Result)

	wg.Add(workersCount)

	for i := 0; i< workersCount;i ++ {
		go worker(ctx,i+1, jobs,results, &wg)
	}

	for _, userID := range userIDs {
		jobs <- userID
	}
	close(jobs)

	go func ()  {
		wg.Wait()	
		close(results)
			
	}()

	allResults := make([]Result, 0, len(userIDs))

	for result := range results {
		allResults = append(allResults, result)
	}



	return  allResults
}

func worker(ctx context.Context,id int, jobs <-chan int, results chan <- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for userID := range jobs {
		user, err := fetchUser(ctx,userID)
		results <- Result{
			UserID:  userID,
			User: user,
			Err: err,
		}		
	}
}

func fetchUser(ctx context.Context, userID int) (User, error) {
	select {
	case <-ctx.Done():
		return User{}, ctx.Err()

	case <-time.After(500 * time.Millisecond):
		if userID == 4 || userID == 9 {
			return User{}, errors.New("not allowed user")
		}
		return User{
			ID:   userID,
			Name: fmt.Sprintf("User id %d", userID),
		}, nil

	}
}

