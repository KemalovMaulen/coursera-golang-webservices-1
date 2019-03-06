package main

import (
	"fmt"
	"net/http"
)

func handlerHttp(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Привет, мир!")
	w.Write([]byte("!!!"))
}

func main() {
	http.HandleFunc("/", handlerHttp)

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)
}
