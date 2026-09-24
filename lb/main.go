package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
)

func main() {
	port := flag.Int("p1", 9091, "First Backend Port")
	flag.Parse()
	backendURL := fmt.Sprintf("http://localhost:%d", *port)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(backendURL + r.URL.Path)
		req, err := http.NewRequest(r.Method, backendURL+r.URL.Path, r.Body)
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
	})
	fmt.Println("Server listening on port: 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
}
