package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestWinnerMaker(t *testing.T) {
	var wm WinnerMaker
	wm.Add(&SourceResult{5, "fibo"})
	wm.Add(&SourceResult{3, "fibo"})
	wm.Add(&SourceResult{8, "fibo"})
	wm.Add(&SourceResult{3, "primes"})
	wm.Add(&SourceResult{5, "primes"})
	wm.Add(&SourceResult{7, "primes"})

	w := wm.Make()
	if w.Price != 7 {
		t.Errorf("Wrong prime: %d", w.Price)
	}
	if w.Source != "fibo" {
		t.Errorf("Wrong source: %s", w.Source)
	}
}

func TestGetWinner(t *testing.T) {
	type args struct {
		sources []string
		sb      SenderBuilder
	}
	tests := []struct {
		name string
		args args
		want *Winner
	}{
		// TODO: Add test cases.
		{"base", args{sources: []string{"fibo", "primes"}, sb: SenderBuilderFunc(func() HTTPSender {
			return HTTPSenderFunc(func(req *http.Request) (*http.Response, error) {
				body := map[string]string{
					"fibo":   `[{"price": 5}, {"price": 3}, {"price": 8}]`,
					"primes": `[{"price": 3}, {"price": 5}, {"price": 7}]`,
				}[req.URL.RequestURI()]
				resp := &http.Response{
					Body:       ioutil.NopCloser(strings.NewReader(body)),
					StatusCode: 200,
				}
				return resp, nil
			})
		})}, &Winner{Price: 7, Source: "fibo"}},
		{"empty", args{sb: SenderBuilderFunc(func() HTTPSender {
			return HTTPSenderFunc(func(req *http.Request) (*http.Response, error) {
				resp := &http.Response{
					Body: ioutil.NopCloser(strings.NewReader(`[]`)),
				}
				return resp, nil
			})
		})}, &Winner{}},
		{"error", args{sources: []string{"fibo", "primes"}, sb: SenderBuilderFunc(func() HTTPSender {
			return HTTPSenderFunc(func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("error")
			})
		})}, &Winner{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetWinner(tt.args.sources, tt.args.sb)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetWinner() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetWinner_Timeout(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.RequestURI {
		case "/max":
			time.Sleep(time.Second)
			w.Write([]byte(`[{"price": 14}, {"price": 16}, {"price": 15}]`))
		case "/fibo":
			w.Write([]byte(`[{"price": 5}, {"price": 3}, {"price": 8}]`))
		case "/primes":
			w.Write([]byte(`[{"price": 3}, {"price": 5}, {"price": 7}]`))
		}
	}))

	w := GetWinner([]string{s.URL + "/fibo", s.URL + "/primes", s.URL + "/max"}, DefaultSenderBuilderFunc(10*time.Millisecond))
	if w.Price != 7 {
		t.Errorf("Wrong winner price: %d", w.Price)
	}
	if w.Source != s.URL+"/fibo" {
		t.Errorf("Wrong winner source: %s", w.Source)
	}
}
