package tvsocket

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"strings"
	"time"

	"testing"
)

func TestSocket_RequestQuotesADD(t *testing.T) {
	var quote QuoteData
	symbol := "USI:ADD"
	//fields := []string{"close_price", "open_price", "high_price", "low_price"}
	tv, err := Connect(func(sym string, q *QuoteData) {
		//fmt.Printf("*** Received %s: %s\n", sym, q.String())
		if sym == symbol {
			quote = *q
		}
	}, func(err error, ctx string) {
		if !strings.Contains(err.Error(), "closed") {
			fmt.Printf("*** Error: %v ctx %s\n", err, ctx)
		}
	})
	require.NoError(t, err)
	defer tv.Close()
	_ = quote
	bars := 5
	interval := "1H"
	err = tv.RequestQuotes(symbol, bars, interval, OnReceiveQuote)
	require.NoError(t, err)
	repeat := 2
	for i := 0; i < repeat; i++ {
		time.Sleep(1 * time.Second)
	}
	tv.Close()
}

func TestSocket_RequestQuotesSTOCK(t *testing.T) {
	var quote QuoteData
	symbol := "NASDAQ:NVDA"
	//fields := []string{"close_price", "open_price", "high_price", "low_price"}
	tv, err := Connect(func(sym string, q *QuoteData) {
		//fmt.Printf("*** Received %s: %s\n", sym, q.String())
		if sym == symbol {
			quote = *q
		}
	}, func(err error, ctx string) {
		if !strings.Contains(err.Error(), "closed") {
			fmt.Printf("*** Error: %v ctx %s\n", err, ctx)
		}
	})
	require.NoError(t, err)
	defer tv.Close()
	_ = quote
	bars := 5
	interval := "5m"
	err = tv.RequestQuotes(symbol, bars, interval, OnReceiveQuote)
	require.NoError(t, err)
	repeat := 10
	for i := 0; i < repeat; i++ {
		time.Sleep(10 * time.Second)
	}
	tv.Close()
}

func OnReceiveQuote(symbol string, hloc []TOHLCV) {
	fmt.Printf(">>> OnReceiveQuote\n")
	for _, v := range hloc {
		tm := time.Unix(v.Time, 0)
		fmt.Printf("Time: %v - Value: %7.2f\n", tm, v.Close)
	}
}
