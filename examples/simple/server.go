package main

import "net/http"

func startServer() error {

	http.HandleFunc("/", pingHandler)
	return http.ListenAndServe(":8080", nil)
}
