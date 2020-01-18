package main

import (
	"fmt"
	"net/http"
)

func getPingHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Content-Type", "Application/json")

	respStr, err := getPingFromService()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error":"%s"}`, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, respStr)

}
