package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
)

type LoadBalancer struct {
	backends []string
	counter  int
}

func (lb *LoadBalancer) getNextBackend() string {
	nextBackend := lb.backends[lb.counter%len(lb.backends)]
	lb.counter++
	return nextBackend
}

func (lb *LoadBalancer) handleRequest(w http.ResponseWriter, r *http.Request) {
	backend := lb.getNextBackend()
	fmt.Println(backend + r.URL.Path)
	req, err := http.NewRequest(r.Method, backend+r.URL.Path, r.Body)
	if err != nil {
		fmt.Println("[NEWREQUEST]Something happend")
		w.WriteHeader(http.StatusInternalServerError)
		return

	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("[CLIENTREQCALL]Something happend")
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("[CONVERSION]Something happend")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println("Came back frm backend")
	for name, values := range res.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	w.WriteHeader(res.StatusCode)
	fmt.Fprint(w, string(resBody))
}

func main() {
	// port := flag.Int("p1", 9091, "First Backend Port")
	flag.Parse()
	lb := LoadBalancer{
		backends: []string{"http://localhost:9091", "http://localhost:9092"},
		counter:  0,
	}
	http.HandleFunc("/", lb.handleRequest)
	fmt.Println("Server listening on port: 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
}
