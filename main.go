package main

import (
	"log"
	"net/http"
)

func main() {
	hr := NewHTTPResources()
	//http.Handle("/", HTTPLogger(hr.serverMux))
	http.Handle("/", hr.serverMux)
	log.Println("Starting server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
