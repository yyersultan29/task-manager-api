package main

import "fmt"

func main() {
	suspects := []int{1, 2, 3, 4, 5, 6, 7, 8, 9} // подозреваемые
	innocents := []int{2, 5, 8}                  //невинные

	// убрать невинных из suspects
	fmt.Println(filter(suspects,innocents))
}

func filter(suspects []int, innocents []int) []int {
	results := make([]int,0, len(suspects))
	set := make(map[int]struct{})

	for _,innocent := range innocents {
		set[innocent] = struct{}{}
	}

	for _, sussuspect := range suspects {
		_, exists := set[sussuspect]

		if !exists {
			results = append(results, sussuspect)
		}
	}
	return  results
}

