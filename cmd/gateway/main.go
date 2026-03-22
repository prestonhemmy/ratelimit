package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prestonhemmy/ratelimit/internal/proxy"
)

func main() {
	revProxy := proxy.NewProxy("http://httpbin.org")

	http.Handle("/", revProxy)

	// initialize server
	fmt.Println("Starting HTTP server on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
