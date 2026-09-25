package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type Backend struct {
	URL     string
	Healthy bool
}

type LoadBalancer struct {
	backends []Backend
	counter  int
	mu       sync.Mutex
}

func (lb *LoadBalancer) getNextBackend() (int, string, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	for attempt := 0; attempt < len(lb.backends); attempt++ {
		counter := lb.counter % len(lb.backends)
		if lb.backends[counter].Healthy == true {
			url := lb.backends[counter].URL
			lb.counter++
			return counter, url, nil
		}
		lb.counter++
	}

	return 0, "", errors.New("No healthy instance found")
}

func (lb *LoadBalancer) checkHealthy(i int) {
	req, err := http.NewRequest("GET", lb.backends[i].URL+"/health", nil)
	if err != nil {
		fmt.Printf("[HEALTH] %s -> unhealthy\n", lb.backends[i].URL)
		return
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("[HEALTH] %s -> unhealthy\n", lb.backends[i].URL)
		lb.mu.Lock()
		defer lb.mu.Unlock()
		lb.backends[i].Healthy = false
		return
	}
	defer res.Body.Close()
	if res.StatusCode == 200 {
		fmt.Printf("[HEALTH] %s -> healthy\n", lb.backends[i].URL)
		lb.mu.Lock()
		defer lb.mu.Unlock()
		lb.backends[i].Healthy = true
		return
	}
	fmt.Printf("[HEALTH] %s -> unhealthy\n", lb.backends[i].URL)
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.backends[i].Healthy = false
}

func (lb *LoadBalancer) startHealthChecks() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		lb.checkAllBackends()
	}
}

func (lb *LoadBalancer) checkAllBackends() {
	for i := range lb.backends {
		lb.checkHealthy(i)
	}
}

func (lb *LoadBalancer) handleRequest(w http.ResponseWriter, r *http.Request) {
	idx, backend, err := lb.getNextBackend()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
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
		lb.mu.Lock()
		defer lb.mu.Unlock()
		lb.backends[idx].Healthy = false
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
		backends: []Backend{{
			URL:     "http://localhost:9091",
			Healthy: true,
		},
			{
				URL:     "http://localhost:9092",
				Healthy: true,
			}},
		counter: 0,
	}
	lb.checkAllBackends()
	go lb.startHealthChecks()
	http.HandleFunc("/", lb.handleRequest)
	fmt.Println("Server listening on port: 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
}
