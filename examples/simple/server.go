package main

import "net/http"

func startServer() error {

	http.HandleFunc("/", getPingHandler)
	return http.ListenAndServe(":8080", nil)
}
