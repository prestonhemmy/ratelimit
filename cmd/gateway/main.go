package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prestonhemmy/ratelimit/internal/config"
	"github.com/prestonhemmy/ratelimit/internal/proxy"
)

func main() {
	// load config
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// TEMP (DEBUGGING)
	fmt.Printf("%#v\n", cfg)

	// create reverse proxy
	revProxy := proxy.NewProxy(cfg.Backend.Url)

	http.Handle("/", revProxy)

	// initialize server
	_, err = fmt.Printf("Starting HTTP server on port %d\n", cfg.Server.Port)
	if err != nil {
		log.Fatal(err)
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	if err = http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
