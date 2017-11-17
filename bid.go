package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Bid - bid from external source
type Bid struct {
	Price int `json:"price"`
}

// GetBids - get bids from external source
//
// s - request sender
// source - url of external source
func GetBids(s HTTPSender, source string) ([]*Bid, error) {
	req, err := http.NewRequest("GET", source, nil)
	if err != nil {
		return nil, fmt.Errorf("error on create request: %v", err)
	}

	resp, err := s.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close() // TODO check error
	}
	if err != nil {
		return nil, fmt.Errorf("error on send request: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("wrong status: %d", resp.StatusCode)
	}

	var res []*Bid
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, fmt.Errorf("error on decode json: %v", err) // TODO print raw body
	}
	return res, nil
}

// Filter2HigherBids - return two higher bids in order by asc
func Filter2HigherBids(bids []*Bid) []*Bid {
	switch len(bids) {
	case 0:
		return nil
	case 1:
		return bids
	case 2:
		if bids[0].Price < bids[1].Price {
			bids[0], bids[1] = bids[1], bids[0]
		}
		return bids
	}

	var first = bids[0]
	var second = bids[1]
	if second.Price > first.Price {
		first, second = second, first
	}

	var b *Bid
	for bi := 2; bi < len(bids); bi++ { // start from third elem
		b = bids[bi]
		if b.Price > first.Price {
			second = first
			first = bids[bi]
		} else if b.Price > second.Price {
			second = bids[bi]
		}
	}
	return []*Bid{first, second}
}
