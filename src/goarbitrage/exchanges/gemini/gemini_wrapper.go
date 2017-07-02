package gemini

import (
	"fmt"
	"net/url"

	"github.com/mgutz/logxi/v1"

	"goarbitrage/exchanges"
)

func (g *Gemini) UpdateDepth(done chan struct{}, resp chan exchange.TaskResponse) {
	var (
		values url.Values
	)

	if g.Verbose {
		log.Info(fmt.Sprintf("%s polling delay: %ds.\n", g.GetName(), g.RESTPollingDelay))
		log.Info(fmt.Sprintf("%s currencies enabled: %s.\n", g.GetName(), g.Symbol))
	}

	book, err := g.GetOrderBook(g.GetSymbol(), values)
	if err != nil {
		log.Error(fmt.Sprintf("Error get order book %s(%s)", g.GetName(), g.Symbol), "error", err.Error())
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
		Name:      g.Name,
		OrderBook: t,
	}

	return
}
