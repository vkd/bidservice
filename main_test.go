package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func Test_main_NotFound(t *testing.T) {
	testTimeout := 10
	timeout = &testTimeout
	go main()
	time.Sleep(10 * time.Millisecond)

	req, err := http.NewRequest("GET", "http://localhost"+*addr, nil)
	if err != nil {
		t.Fatalf("Error on create request: %v", err)
	}
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("Error on send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Fatalf("Wrong not found: %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error on read response body: %v", err)
	}
	if string(body) != "404 page not found\n" {
		t.Errorf("Wrong response body: %q", string(body))
	}
}

func Test_main_winner(t *testing.T) {
	address := ":8083"
	addr = &address
	testTimeout := 10
	timeout = &testTimeout
	go main()
	time.Sleep(10 * time.Millisecond)

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

	url := "http://localhost" + *addr + "/winner"
	url += "?s=" + s.URL + "/primes"
	url += "&s=" + s.URL + "/max"
	url += "&s=" + s.URL + "/fibo"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("Error on create request: %v", err)
	}
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("Error on send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("Wrong status: %d", resp.StatusCode)
	}

	res := struct {
		Price  int    `json:"price"`
		Source string `json:"source"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		t.Fatalf("Error on decode response: %v", err)
	}

	if res.Price != 7 {
		t.Errorf("Wrong price: %d", res.Price)
	}
	if res.Source != s.URL+"/fibo" {
		t.Errorf("Wrong source: %s", res.Source)
	}
}

func TestWinnerHandler(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		sources  []string
		winner   *Winner
		status   int
		wantBody string
	}{
		// TODO: Add test cases.
		{"2 sources", "/winner?s=/fibo&s=/primes", []string{"/fibo", "/primes"}, &Winner{7, "/primes"}, 200, "{\"price\":7,\"source\":\"/primes\"}\n"},
		{"one x param", "/winner?s=/fibo&x=/max&s=/primes", []string{"/fibo", "/primes"}, &Winner{7, "/primes"}, 200, "{\"price\":7,\"source\":\"/primes\"}\n"},
		{"one empty", "/winner?s=", []string{""}, &Winner{7, ""}, 200, "{\"price\":7,\"source\":\"\"}\n"},
		{"error empty query", "/winner", nil, nil, 400, "'s' query is empty\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := http.NewRequest("GET", tt.url, nil)
			if err != nil {
				t.Fatalf("Error on create request: %v", err)
			}
			w := httptest.NewRecorder()
			WinnerHandler(func(ss []string) *Winner {
				if !reflect.DeepEqual(ss, tt.sources) {
					t.Errorf("Wrong input sources: %v (want: %v)", ss, tt.sources)
				}
				return tt.winner
			}).ServeHTTP(w, r)
			if w.Code != tt.status {
				t.Errorf("Wrong status: %d (want: %d)", w.Code, tt.status)
			}
			if w.Body.String() != tt.wantBody {
				t.Errorf("Expect resp body:    %s", tt.wantBody)
				t.Errorf("Wrong response body: %s", w.Body.String())
			}
		})
	}
}

func TestWinnerHandler_WrongMethod(t *testing.T) {
	r, err := http.NewRequest("POST", "/winner", nil)
	if err != nil {
		t.Fatalf("Error on create request: %v", err)
	}
	w := httptest.NewRecorder()
	WinnerHandler(nil).ServeHTTP(w, r)
	if w.Code != 404 {
		t.Errorf("Wrong status: %d (want: %d)", w.Code, 404)
	}
	if w.Body.String() != "404 page not found\n" {
		t.Errorf("Wrong response body: %s", w.Body.String())
	}
}
