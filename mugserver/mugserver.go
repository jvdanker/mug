package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Url struct {
	Id  int    `json:"id"`
	Url string `json:"url"`
}

func main() {
	fmt.Println("mugserver, listening at :8080...")

	http.HandleFunc("/list", handleGetRequests)
	http.HandleFunc("/scan/", handleScanRequests)
	http.ListenAndServe(":8080", nil)
}

func handleGetRequests(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	if r.Method == "OPTIONS" {
		return
	}

	data := []Url{
		{Id: 1, Url: "1"},
		{Id: 2, Url: "2"},
		{Id: 3, Url: "3"},
		{Id: 4, Url: "4"},
	}

	j, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, fmt.Sprintf("%s", j))
}

func handleScanRequests(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)

	id := r.URL.Path[len("/scan/"):]
	fmt.Println(id)

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
