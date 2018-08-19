package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("mugserver, listening at :8080...")

	http.HandleFunc("/scan", handleRequests)
	http.ListenAndServe(":8080", nil)
}

func handleRequests(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)

	setupResponse(&w, r)
	if r.Method == "OPTIONS" {
		return
	}

	j, err := json.MarshalIndent("", "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, fmt.Sprintf("%s", j))
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}
