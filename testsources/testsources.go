package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type row struct {
	Price int `json:"price"`
}

func main() {
	listenAddr := flag.String("addr", ":8080", "http listen address")
	flag.Parse()

	http.HandleFunc("/primes", handler([]int{2, 3, 5, 7, 11, 13, 17, 19, 23}))
	http.HandleFunc("/fibo", handler([]int{1, 1, 2, 3, 5, 8, 13, 21}))
	http.HandleFunc("/fact", handler([]int{1, 2, 6, 24}))
	http.HandleFunc("/rand", handler([]int{5, 17, 3, 19, 76, 24, 1, 5, 10, 34, 8, 27, 7}))
	http.HandleFunc("/empty", handler([]int{}))

	fmt.Printf("Listen on %s", *listenAddr)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}

func handler(weights []int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		lat := rand.Intn(1000)
		time.Sleep(time.Duration(lat) * time.Millisecond)

		x := rand.Intn(100)
		if x < 10 {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Randomisation
		r := rand.New(rand.NewSource(time.Now().Unix()))

		resp := make([]row, 0, len(weights))
		for _, k := range r.Perm(len(weights)) {
			resp = append(resp, row{weights[k]})
		}

		json.NewEncoder(w).Encode(resp)
	}
}
