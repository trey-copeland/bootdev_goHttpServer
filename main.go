package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	mux := &http.ServeMux{}
	mux.Handle("/", http.FileServer(http.Dir(".")))
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
	}

}
