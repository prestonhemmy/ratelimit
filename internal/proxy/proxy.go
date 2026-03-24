package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Configures a reverse proxy that forwards incoming requests to the configured
// backend URL.

func NewProxy(targetURL string) http.Handler {
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	fmt.Println("API Gateway created")

	revProxy := httputil.NewSingleHostReverseProxy(target)

	return revProxy
}
