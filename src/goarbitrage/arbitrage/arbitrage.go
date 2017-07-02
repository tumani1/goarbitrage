package arbitrage

import (
	"fmt"
	"math"
	"time"

	"github.com/mgutz/logxi/v1"

	"goarbitrage/config"
	"goarbitrage/exchanges"
)

type (
	ArbitrageStrategy struct {
		Exchanges map[string]exchange.IBotExchange
		Depths    map[string]exchange.OrderBook
		shutdown  chan struct{}
	}

	Profit struct {
		Profit            float64
		Volume            float64
		WeightedBuyPrice  float64
		WeightedSellPrice float64
		BuyPrice          float64
		SellPrice         float64
	}
)

func (a *ArbitrageStrategy) updateDepths() {
	var (
		done chan struct{}
	)

	if a.Depths == nil {
		a.Depths = map[string]exchange.OrderBook{}
	}

	resp := make(chan exchange.TaskResponse, len(a.Exchanges))
	for _, v := range a.Exchanges {
		log.Info("Send", "info", v.GetName())
		go v.UpdateDepth(done, resp)
	}

	for n := 0; n < len(a.Exchanges); n++ {
		select {
		case data := <-resp:
			log.Info("Resp:", "info", data.Name)
			a.Depths[data.Name] = data.OrderBook
		}
	}
}

func (a *ArbitrageStrategy) tick() {
	log.Info("tick function", "info")

	for k1, _ := range a.Depths {
		for k2, _ := range a.Depths {
			if k1 == k2 {
				continue
			}

			ex1 := a.Depths[k1]
			ex2 := a.Depths[k2]

			if ex1.Asks[0].Price < ex2.Bids[0].Price {
				a.arbitrageOpportunity(k1, k2)
			}
		}
	}
}

func (a *ArbitrageStrategy) arbitrageOpportunity(kask, kbid string) {
	result := a.arbitrageDepthOpportunity(kask, kbid)
	if result.Volume == 0 || result.BuyPrice == 0 {
		return
	}

	perc := (result.WeightedSellPrice - result.WeightedBuyPrice) / result.BuyPrice * 100
	log.Info("Percent:", "info", perc)
}

func (a *ArbitrageStrategy) arbitrageDepthOpportunity(kask, kbid string) Profit {
	var (
		profit                 Profit
		bestAskPos, bestBidPos int
	)

	askPos, bidPos := a.getMaxDepth(kask, kbid)
	for i := 0; i < askPos+1; i++ {
		for j := 0; j < bidPos+1; j++ {
			tempProfit := a.getProfitFor(i, j, kask, kbid)

			if tempProfit.Profit >= 0 && tempProfit.Profit >= profit.Profit {
				profit = tempProfit
				bestAskPos, bestBidPos = i, j
			}
		}
	}

	profit.BuyPrice = a.Depths[kask].Asks[bestAskPos].Price
	profit.SellPrice = a.Depths[kbid].Bids[bestBidPos].Price
	return profit
}

func (a *ArbitrageStrategy) getMaxDepth(kask, kbid string) (int, int) {
	var (
		askPos, bidPos int
	)

	if len(a.Depths[kbid].Bids) > 0 && len(a.Depths[kask].Asks) > 0 {
		for a.Depths[kask].Asks[askPos].Price < a.Depths[kbid].Bids[0].Price {
			if askPos >= len(a.Depths[kask].Asks)-1 {
				break
			}

			askPos += 1
		}

		for a.Depths[kask].Asks[0].Price < a.Depths[kbid].Bids[bidPos].Price {
			if bidPos >= len(a.Depths[kbid].Bids)-1 {
				break
			}

			bidPos += 1
		}
	}

	return askPos, bidPos
}

func (a *ArbitrageStrategy) getProfitFor(askPos, bidPos int, kask, kbid string) Profit {
	if a.Depths[kask].Asks[askPos].Price >= a.Depths[kbid].Bids[bidPos].Price {
		return Profit{}
	}

	var (
		maxAmountBuy, maxAmountSell float64
	)

	for i := 0; i < askPos+1; i++ {
		maxAmountBuy += a.Depths[kask].Asks[i].Amount
	}

	for j := 0; j < bidPos+1; j++ {
		maxAmountSell += a.Depths[kbid].Bids[j].Amount
	}

	maxAmount := math.Min(math.Min(maxAmountBuy, maxAmountSell), config.Cfg.Settings.MaxTxVolume)

	var (
		buyTotal, weightedBuyPrice float64
	)
	for i := 0; i < askPos+1; i++ {
		price := a.Depths[kask].Asks[i].Price
		amount := math.Min(maxAmount, buyTotal+a.Depths[kask].Asks[i].Amount) - buyTotal
		if amount <= .0 {
			break
		}

		buyTotal += amount
		if weightedBuyPrice == .0 {
			weightedBuyPrice = price
		} else {
			weightedBuyPrice = (weightedBuyPrice*(buyTotal-amount) + price*amount) / buyTotal
		}
	}

	var (
		sellTotal, weightedSellPrice float64
	)
	for j := 0; j < bidPos+1; j++ {
		price := a.Depths[kbid].Bids[j].Price
		amount := math.Min(maxAmount, sellTotal+a.Depths[kbid].Bids[j].Amount) - sellTotal
		if amount <= .0 {
			break
		}

		sellTotal += amount
		if weightedSellPrice == .0 || sellTotal == .0 {
			weightedSellPrice = price
		} else {
			weightedSellPrice = (weightedSellPrice*(sellTotal-amount) + price*amount) / sellTotal
		}
	}

	if math.Abs(sellTotal-buyTotal) > float64(0.00001) {
		log.Warn(fmt.Sprintf("sell_total=%v,buy_total=%v", sellTotal, buyTotal), "warn")
	}

	profit := sellTotal*weightedSellPrice - buyTotal*weightedBuyPrice
	return Profit{
		Profit:            profit,
		Volume:            sellTotal,
		WeightedSellPrice: weightedSellPrice,
		WeightedBuyPrice:  weightedBuyPrice,
	}
}

func (a *ArbitrageStrategy) Loop() {
	for {
		a.updateDepths()
		a.tick()

		time.Sleep(time.Second * 10)
	}
}
