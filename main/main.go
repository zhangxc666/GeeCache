package main

import (
	"GeeCache/GeeCache/consistenthash"
	"fmt"
	"strconv"
)

func main() {

	hash := consistenthash.New(3, func(data []byte) uint32 {
		i, _ := strconv.Atoi(string(data))
		return uint32(i)
	})
	hash.Add("6", "4", "2")
	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}
	for k, v := range testCases {
		if hash.Get(k) != v {
			fmt.Println(k, v)
		}
	}
	testCases["27"] = "2"
	for k, v := range testCases {
		if hash.Get(k) != v {
			fmt.Println(k, v)
		}
	}

}
