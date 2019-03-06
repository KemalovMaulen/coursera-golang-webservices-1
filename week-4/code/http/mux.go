package main

import (
	"fmt"
	"net/http"
	"time"
)

func handlerMux(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "URL:", r.URL.String())
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlerMux)

	server := http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("starting server at :8080")
	server.ListenAndServe()
}
