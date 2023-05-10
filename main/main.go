package main

import (
	"GeeCache/GeeCache"
	"fmt"
	"log"
	"net/http"
)

func main() {
	db := map[string]string{
		"abc": "111",
		"def": "222",
		"ghi": "333",
	}
	GeeCache.NewGroup("zxc", 2<<10, GeeCache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, err := db[key]; err {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		},
	))
	pool := GeeCache.NewHTTPPool("localhost:8080")
	log.Println("geecache is running at localhost:8080")
	log.Fatal(http.ListenAndServe("localhost:8080", pool))
}
