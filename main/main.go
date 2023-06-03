package main

import (
	"GeeCache/GeeCache"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"zxc": "666",
	"zzz": "555",
	"xxx": "777",
}

func createGroup() *GeeCache.Group {
	return GeeCache.NewGroup("scores", 2<<10, GeeCache.GetterFunc(func(key string) ([]byte, error) {
		log.Println("[SlowDB] search key", key)
		if v, ok := db[key]; ok {
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	}))
}

func startCacheServer(addr string, addrs []string, gee *GeeCache.Group) {
	peers := GeeCache.NewHTTPServer(addr)
	peers.Set(addrs...)
	gee.RegisterPeers(peers)
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, gee *GeeCache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := gee.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var (
		port int
	)
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.Parse()
	addrMap := map[int][]string{
		8001: {"http://localhost:8001", "http://localhost:9001"},
		8002: {"http://localhost:8002", "http://localhost:9002"},
		8003: {"http://localhost:8003", "http://localhost:9003"},
	}

	addrs := make([]string, len(addrMap))
	for _, v := range addrMap {
		addrs = append(addrs, v[0])
	}
	addr, ok := addrMap[port]
	if ok == false {
		return
	}
	gee := createGroup()
	go startAPIServer(addr[1], gee)
	startCacheServer(addr[0], addrs, gee)
}
