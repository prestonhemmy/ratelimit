package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewProxy(targetURL string) http.Handler {
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	fmt.Println("Gateway created")

	revProxy := httputil.NewSingleHostReverseProxy(target)

	return revProxy
}
