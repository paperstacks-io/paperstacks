package main

import (
	"fmt"
	"net/http"
)

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
}

func main() {
	fmt.Println("Starting server at port 8090")
	http.HandleFunc("/hello", hello)
	http.ListenAndServe(":8090", nil)
}
