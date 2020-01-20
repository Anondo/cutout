package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func pingHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Content-Type", "Application/json")

	pr, err := getPingFromService()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error":"%s"}`, err.Error())
		return
	}

	byteRep, err := json.Marshal(pr)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error":"%s"}`, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(byteRep))

}
