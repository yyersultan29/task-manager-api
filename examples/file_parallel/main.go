package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const timeoutLimit = 90

type Result struct {
	msg string
	err error
}

// имитация скачка документов в разное вермя 
func fakeDownload(url string, wg *sync.WaitGroup,results chan <- Result)  {
	defer wg.Done()
	r := rand.Intn(100)
	time.Sleep(time.Duration(r) * time.Millisecond)

	if r > timeoutLimit{
		results<- Result{
			err: fmt.Errorf("failed to download data from %s: timeout", url),
		}
		return
	}
	results<- Result{
		msg: fmt.Sprintf("dowmload data from %s\n", url),
	}
}

// download - параллельно скачивает данные из urls
func download(urls []string) ([]string, error) {
	// here goes 

	var wg sync.WaitGroup
	results := make(chan Result, len(urls))

	wg.Add(len(urls))

	for _,url := range urls {
		go fakeDownload(url, &wg,results)

	}

	go func() {
		wg.Wait()
		close(results)
	}()

	resultArr := make([]string,0,len(urls))
	var firstErr error

	for result := range results {
		if result.err != nil {
			if firstErr == nil {
				firstErr = result.err
			}
			continue
		}
		resultArr = append(resultArr, result.msg)
	}

	if firstErr != nil {
		return nil, firstErr
	}
  
	return resultArr, nil
  
}

func main() {
	msgs, err := download([]string{
		"https://example.com/e25e26d3-6aa3-4d79-9ab4-fc9b71103a8c.xml",
		"https://example.com/a601590e-31c1-424a-8ccc-decf5b35c0f6.xml",
		"https://example.com/1cf0dd69-a3e5-4682-84e3-dfe22ca771f4.xml",
		"https://example.com/ceb566f2-a234-4cb8-9466-4a26f1363aa8.xml",
		"https://example.com/b6ed16d7-cb3d-4cba-b81a-01a789d3a914.xml",
	})

	if err != nil {
		panic(err)
	}

	fmt.Println(msgs)
}