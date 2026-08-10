package main

import (
	"fmt"
)

func main() {

	numbers := make(chan int)
	squares := make(chan int)
	evens := make(chan int)
	
	// заполняем наш numbers
	go producer(numbers)

	// filter even numerbers
	go filterEven(numbers,evens)

	// numbers есть и сдеалем их квадрать и записываем на squares
	go squarer(evens, squares)

	for num := range squares {
		fmt.Println(num)
	}
}

func producer(numbers chan<- int) {

	// закрыть канал после выполнение функций
	defer close(numbers)

	for i := 1;i < 6;i ++ {
		numbers <- i
	}
}

func filterEven(numbers <- chan int,evens chan <- int) {
	defer close(evens)
	for num := range numbers {
		if num % 2 == 0 {
			evens <- num
		}
	}
}

func squarer(numbers <-chan int, squares chan<- int) {
	defer close(squares)

	for num := range numbers {
		squares <- num * num
	}
}