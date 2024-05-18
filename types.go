package tvsocket

import (
	"fmt"
	"strings"
	"time"
)

// SocketInterface ...
type SocketInterface interface {
	AddSymbol(symbol string) error
	RemoveSymbol(symbol string) error
	Init(fields ...string) error
	Close() error
	RequestQuotes(symbol string, bars int, interval string, resultCallback OnReceiveQuoteCallback) error
}

// SocketMessage ...
type SocketMessage struct {
	Message string      `json:"m"`
	Payload interface{} `json:"p"`
}

// QuoteMessage ...
type QuoteMessage struct {
	Symbol string     `mapstructure:"n"`
	Status string     `mapstructure:"s"`
	Data   *QuoteData `mapstructure:"v"`
}

//	getSocketMessage("quote_set_fields", []string{s.quoteSessionID, "lp", "volume", "bid", "ask", "ch", "chp"}),
//
// QuoteData ...
type QuoteData struct {
	Price             *float64 `mapstructure:"lp"`
	PrevClosePrice    *float64 `mapstructure:"prev_close_price"`
	RegularClosePrice *float64 `mapstructure:"regular_close_price"`
	RegularCloseTime  *int64   `mapstructure:"regular_close_time"`
	HighPrice         *float64 `mapstructure:"high_price"`
	LowPrice          *float64 `mapstructure:"low_price"`
	OpenPrice         *float64 `mapstructure:"open_price"`
	OpenTime          *int64   `mapstructure:"open_time"`
	Volume            *float64 `mapstructure:"volume"`
	Bid               *float64 `mapstructure:"bid"`
	Ask               *float64 `mapstructure:"ask"`
	Change            *float64 `mapstructure:"ch"`
	Time              *int64   `mapstructure:"lp_time"`
}

// Flags ...
type Flags struct {
	Flags []string `json:"flags"`
}

// OnReceiveDataCallback ...
type OnReceiveDataCallback func(symbol string, data *QuoteData)

type OnReceiveQuoteCallback func(symbol string, hloc []HLOC)

// OnErrorCallback ...
type OnErrorCallback func(err error, context string)

func (q *QuoteData) String() string {
	sb := new(strings.Builder)
	if q.OpenPrice != nil {
		sb.WriteString(fmt.Sprintf("open price: %7.2f | ", *q.OpenPrice))
	}
	if q.Price != nil {
		sb.WriteString(fmt.Sprintf("price: %7.2f | ", *q.Price))
	}
	if q.Time != nil {
		//tm := chrono.UnixToMarketTime(*q.Time)
		tm := time.Unix(*q.Time, 0)
		sb.WriteString(fmt.Sprintf("time: %v | ", tm))
	}
	if q.Bid != nil {
		sb.WriteString(fmt.Sprintf("bid %7.2f | ", *q.Bid))
	}
	if q.Ask != nil {
		sb.WriteString(fmt.Sprintf("ask: %7.2f | ", *q.Ask))
	}
	if q.OpenTime != nil {
		tm := time.Unix(*q.OpenTime, 0)
		sb.WriteString(fmt.Sprintf("open time: %v | ", tm))
	}
	if q.RegularClosePrice != nil {
		sb.WriteString(fmt.Sprintf("regular close price: %7.2f | ", *q.RegularClosePrice))
	}
	if q.RegularCloseTime != nil {
		tm := time.Unix(*q.RegularCloseTime, 0)
		sb.WriteString(fmt.Sprintf("regular close time: %v | ", tm))
	}
	if q.HighPrice != nil {
		sb.WriteString(fmt.Sprintf("high price: %7.2f | ", *q.HighPrice))
	}
	if q.LowPrice != nil {
		sb.WriteString(fmt.Sprintf(" low price: %7.2f | ", *q.LowPrice))
	}
	if q.PrevClosePrice != nil {
		sb.WriteString(fmt.Sprintf("prev.close price: %7.2f | ", *q.PrevClosePrice))
	}
	if q.Change != nil {
		sb.WriteString(fmt.Sprintf("change: %7.2f | ", *q.Change))
	}
	if q.Volume != nil {
		sb.WriteString(fmt.Sprintf("volume: %7.2fM | ", *q.Volume/1000000.))
	}
	return sb.String()
}
