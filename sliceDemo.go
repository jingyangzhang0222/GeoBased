package main

import "fmt"

func main() {
	array1 := []string{"a", "b", "c"}
	array2 := []int{1, 2, 3}
	for _,v := range array1 {
		fmt.Println(v)
	}
	for _,v := range array2 {
		fmt.Println(v)
	}
}
