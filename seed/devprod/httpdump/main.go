package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"
)

var flagListen = flag.String("listen", "0.0.0.0:8000", "listen server port")
var flagTarget = flag.String("target", "127.0.0.1:9000", "target server port")

func dump(w http.ResponseWriter, r *http.Request) {
	log.Printf("%#v %#v", r.Method, r.RequestURI)
	log.Printf("From: %#v", r.RemoteAddr)
	for key, values := range r.Header {
		for _, value := range values {
			log.Printf("Header [%#v]: %#v", key, value)
		}
	}
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Read Request Error: "+err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Body (bytes): %#v", requestBody)
	log.Printf("Body (string): %#v", string(requestBody))
	log.Println()
	r.URL.Scheme = "http"
	r.URL.Host = *flagTarget
	log.Printf("Forward to %#v", (r.URL.String()))
	proxyRequest, err := http.NewRequest(r.Method, r.URL.String(), bytes.NewReader(requestBody))
	if err != nil {
		http.Error(w, "Proxy Request Error: "+err.Error(), http.StatusBadRequest)
		return
	}
	proxyRequest.Header = r.Header
	proxyResponse, err := http.DefaultClient.Do(proxyRequest)
	if err != nil {
		http.Error(w, "Proxy Do Error: "+err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("%#v %#v", proxyResponse.StatusCode, proxyResponse.Status)
	for key, values := range proxyResponse.Header {
		for _, value := range values {
			log.Printf("Header [%#v]: %#v", key, value)
		}
	}
	responseBody, err := io.ReadAll(proxyResponse.Body)
	if err != nil {
		http.Error(w, "Read Response Error: "+err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Body (bytes): %#v", responseBody)
	log.Printf("Body (string): %#v", string(responseBody))
	log.Println()
	for key, values := range proxyResponse.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(proxyResponse.StatusCode)
	w.Write(responseBody)
}

func main() {
	http.HandleFunc("/", dump)
	log.Printf("Listening to %#v", *flagListen)
	err := http.ListenAndServe(*flagListen, nil)
	if err != nil {
		log.Fatalf("Listen & serve error: %#v", err)
	}
}
