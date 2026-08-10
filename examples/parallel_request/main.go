package main

import (
	"fmt"
	"net/http"
	"sync"
)

// таска сделать их параллельным
func main() {
	var wg sync.WaitGroup


	urls := []string{
		"https://google.com",
		"https://dzen.ru/",
	}
	results := make(chan http.Response,len(urls));


	wg.Add(len(urls))
	for _,url := range urls {
		go fetch(url,results,&wg)

	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		fmt.Println(res.Status)
		res.Body.Close()
	}

	
}

func fetch(url string, results chan <- http.Response,wg *sync.WaitGroup)  {
	defer wg.Done()
	response, err := http.Get(url);

	if err != nil {
		fmt.Println("Request err", err.Error()); 
		return
	}

	results <- *response

}

