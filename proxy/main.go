package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func main() {
	targetUrl := "https://api.openai.com"
	target, err := url.Parse(targetUrl)
	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.Director = func(req *http.Request) {
		req.Host = target.Host
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
	}

	proxy.ModifyResponse = func(resp *http.Response) error {
		responseDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Printf("Failed to dump response: %v\n", err)
		} else {
			log.Printf("Response: \n%s\n", string(responseDump))
		}

		return nil
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		requestDump, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Printf("Failed to dump request: \n%v\n", err)
		} else {
			log.Printf("%s Request: %s\n", time.Now().Format("2006-01-02 15:04:05"), string(requestDump))
		}
		proxy.ServeHTTP(w, r)
	})

	log.Printf("Starting server on port 9000...\n")
	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatal(err)
	}
}
