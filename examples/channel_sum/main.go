package main

import "fmt"



func main() {

	nums := make(chan int)

	go worker(nums)

	var sum int
	for num := range nums {
		sum += num
	}

	fmt.Println("Sum is ", sum)

}


func worker(nums chan <- int) {
	defer close(nums)
	
	for i := 1;i < 6;i ++ {
		nums <- i
	}
}