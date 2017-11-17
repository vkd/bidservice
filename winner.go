package main

import (
	"log"
	"sync"
)

// Winner - winner of sources bids
type Winner struct {
	Price  int    `json:"price"`
	Source string `json:"source"`
}

// GetWinner - get winner of bids from sources
//
// Result:
// Winner.Source - source with the highest bid
// Winner.Price - the second highest price of all sources
func GetWinner(sources []string, sb SenderBuilder) *Winner {
	resultChan := make(chan *SourceResult)

	var wg sync.WaitGroup
	wg.Add(len(sources))

	go func() {
		// close result chan after all goroutines
		wg.Wait()
		close(resultChan)
	}()

	for _, s := range sources {
		go func(source string) {
			defer wg.Done()
			bids, err := GetBids(sb.HTTPSender(), source)
			if err != nil {
				log.Printf("Error on get bids for source (%q): %v", source, err)
				return
			}
			bids = Filter2HigherBids(bids)
			for i := range bids {
				resultChan <- &SourceResult{bids[i].Price, source}
			}
		}(s)
	}

	var wm WinnerMaker
	for sr := range resultChan {
		wm.Add(sr)
	}

	return wm.Make()
}

// SenderBuilder - builder of HTTPSenders
//
// Create new instance of sender in every goroutine
type SenderBuilder interface {
	HTTPSender() HTTPSender
}

// SenderBuilderFunc - implement SenderBuilder by func
type SenderBuilderFunc func() HTTPSender

// HTTPSender - implement SenderBuilder
func (f SenderBuilderFunc) HTTPSender() HTTPSender {
	return f()
}

// WinnerMaker - create winner by SourceResults
type WinnerMaker struct {
	first, second *SourceResult
}

// Add - add SourceResult
func (w *WinnerMaker) Add(sr *SourceResult) {
	if w.first == nil {
		w.first = sr
		return
	}
	if sr.Price > w.first.Price {
		w.second = w.first
		w.first = sr
		return
	}
	if w.second == nil {
		w.second = sr
		return
	}
	if sr.Price > w.second.Price {
		w.second = sr
		return
	}
}

// Make - make winner by all previous SourceResults
func (w *WinnerMaker) Make() *Winner {
	var res Winner
	if w.first != nil {
		res.Source = w.first.Source
	}
	if w.second != nil {
		res.Price = w.second.Price
	}
	return &res
}

// SourceResult - bid price from source
type SourceResult struct {
	Price  int
	Source string
}
