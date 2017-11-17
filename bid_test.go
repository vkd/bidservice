package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestGetBids(t *testing.T) {
	type args struct {
		s      HTTPSender
		source string
	}
	tests := []struct {
		name    string
		args    args
		want    []*Bid
		wantErr bool
	}{
		// TODO: Add test cases.
		{"bad url", args{source: "://wrong_scheme"}, nil, true},
		{"bad request", args{s: HTTPSenderFunc(func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("error")
		})}, nil, true},
		{"bad status", args{s: HTTPSenderFunc(func(req *http.Request) (*http.Response, error) {
			resp := &http.Response{StatusCode: 400}
			return resp, nil
		})}, nil, true},
		{"bad json", args{s: HTTPSenderFunc(func(req *http.Request) (*http.Response, error) {
			resp := &http.Response{}
			resp.Body = ioutil.NopCloser(strings.NewReader(`{]`))
			resp.StatusCode = 200
			return resp, nil
		})}, nil, true},
		{"ok", args{s: HTTPSenderFunc(func(req *http.Request) (*http.Response, error) {
			resp := &http.Response{
				Body:       ioutil.NopCloser(strings.NewReader(`[{"price": 5}, {"price": 3}, {"price": 8}]`)),
				StatusCode: 200,
			}
			return resp, nil
		})}, []*Bid{{5}, {3}, {8}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBids(tt.args.s, tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBids() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBids() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter2HigherBids(t *testing.T) {
	tests := []struct {
		name string
		bids []*Bid
		want []*Bid
	}{
		// TODO: Add test cases.
		{"0 len", nil, nil},
		{"1", []*Bid{{3}}, []*Bid{{3}}},
		{"2 same", []*Bid{{6}, {4}}, []*Bid{{6}, {4}}},
		{"2 swap", []*Bid{{4}, {6}}, []*Bid{{6}, {4}}},
		{"3 skip", []*Bid{{4}, {6}, {2}}, []*Bid{{6}, {4}}},
		{"3 mid", []*Bid{{4}, {6}, {5}}, []*Bid{{6}, {5}}},
		{"3 first", []*Bid{{4}, {6}, {8}}, []*Bid{{8}, {6}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Filter2HigherBids(tt.bids)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter2HigherBids() = %v, want %v", got, tt.want)
				t.Errorf("Filter2HigherBids() = %v, %v, want %v, %v", got[0], got[1], tt.want[0], tt.want[1])
			}
		})
	}
}
