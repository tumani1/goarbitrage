package bitfinex

import (
	"fmt"
	"net/url"

	"github.com/mgutz/logxi/v1"

	"goarbitrage/exchanges"
)

func (b *Bitfinex) UpdateDepth(done chan struct{}, resp chan exchange.TaskResponse) {
	var (
		values url.Values
	)

	if b.Verbose {
		log.Info(fmt.Sprintf("%s polling delay: %ds.\n", b.GetName(), b.RESTPollingDelay))
		log.Info(fmt.Sprintf("%s currencies enabled: %s.\n", b.GetName(), b.Symbol))
	}

	book, err := b.GetOrderBook(b.GetSymbol(), values)
	if err != nil {
		log.Error(fmt.Sprintf("Error get order book %s(%s)", b.GetName(), b.Symbol), "error", err.Error())
		return
	}

	var t exchange.OrderBook
	for _, i := range book.Bids {
		t.Bids = append(t.Bids, exchange.ItemBook(i))
	}
	for _, i := range book.Asks {
		t.Asks = append(t.Asks, exchange.ItemBook(i))
	}

	resp <- exchange.TaskResponse{
		Name:      b.Name,
		OrderBook: t,
	}

	return
}
