package main

import (
	"flag"
	"fmt"
	"net/http"
)

func main() {
	port := flag.Int("p1", 9091, "First Backend Port")
	flag.Parse()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, fmt.Sprintf("HelloWorld from Port: %d", *port))
	})
	fmt.Println("Server listening on port: ", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		fmt.Println(err)
	}
}
