package main

import (
	"fmt"
	"strings"
)

func main() {
	s := "go-ms_1-2"
	arr := strings.Split(s, "go-ms_")
	fmt.Println(arr[1])
}
