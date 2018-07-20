package main

import (
	"os"
	"fmt"
)

// How to reduce the line of codes for this program
/*
conn, err := os.Open("/tmp/file")
if err != nil {
      panic(err)
} else {
      conn.Read()
}
array := []string{}
x := array[0]
y := array[1]
z := array[2]
fmt.Printf("%s %s %s\n", x, y, z)
*/

func main() {
	if conn, err := os.Open("/tmp/file"); err != nil {
		panic(err)
	} else {
		//conn.Read()
	}
	array := []string{}
	x, y, z := array[0], array[1], array[2]
	fmt.Printf("%s %s %s\n", x, y, z)
}
