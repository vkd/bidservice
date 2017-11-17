package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"
)

var (
	timeout = flag.Int("timeout", 100, "Timeout of requests to source (msec)")
	addr    = flag.String("addr", ":8081", "Server address")
)

func main() {
	flag.Parse()
	sb := DefaultSenderBuilderFunc(time.Duration(*timeout) * time.Millisecond)

	mux := http.NewServeMux()
	mux.HandleFunc("/winner", WinnerHandler(func(sources []string) *Winner {
		return GetWinner(sources, sb)
	}))

	s := http.Server{
		Addr: *addr,
		// TODO set timeouts
		// TODO graceful shutdown
		Handler: mux,
	}
	fmt.Printf("Server starting on %s ...\n", *addr)
	err := s.ListenAndServe()
	fmt.Printf("Server is stopped: %v\n", err)
}

// DefaultSenderBuilderFunc - return SenderBuilder by *http.Client
func DefaultSenderBuilderFunc(timeout time.Duration) SenderBuilderFunc {
	return func() HTTPSender {
		return &http.Client{Timeout: timeout}
	}
}

// WinnerHandler - handler for `GET /winner` method
func WinnerHandler(getWinnerFn func(sources []string) *Winner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}

		sources := r.URL.Query()["s"] // TODO check input source
		if len(sources) == 0 {
			http.Error(w, "'s' query is empty", http.StatusBadRequest)
			return
		}

		winner := getWinnerFn(sources)

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(winner) // TODO check error
	}
}
