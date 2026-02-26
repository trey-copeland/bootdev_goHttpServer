package main

import (
	// "fmt"
	"fmt"
	"net/http"
	"os"
)

func main() {

	mux := &http.ServeMux{}
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
	}

}
