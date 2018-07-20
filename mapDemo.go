package main

import "fmt"

func main(){
	m := make(map[string]bool)
	m["jack"] = true
	m["susan"] = false

	for k,v := range m {
		fmt.Println(k, v)
	}
}
