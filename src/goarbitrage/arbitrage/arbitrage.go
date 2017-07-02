package arbitrage

import (
	"fmt"
	"math"
	"sync"
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

	ProfitStruct struct {
		Profit            float64
		Volume            float64
		WeightedBuyPrice  float64
		WeightedSellPrice float64
		BuyPrice          float64
		SellPrice         float64
	}
)

func New() *ArbitrageStrategy {
	return &ArbitrageStrategy{
		Depths: map[string]exchange.OrderBook{},
	}
}

func (a *ArbitrageStrategy) updateDepths() {
	wg := sync.WaitGroup{}
	done := make(chan struct{})
	resp := make(chan exchange.TaskResponse, len(a.Exchanges))

	for _, v := range a.Exchanges {
		wg.Add(1)
		go v.UpdateDepth(&wg, done, resp)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		cnt := 0
		timeout := time.After(5 * time.Second)

		for {
			select {
			case <-timeout:
				close(done)
				return
			case data := <-resp:
				cnt++
				log.Info("name:", "info", data.Name)
				a.Depths[data.Name] = data.OrderBook

				if cnt == len(a.Exchanges) {
					return
				}
			}
		}
	}()

	wg.Wait()
}

func (a *ArbitrageStrategy) tick() {
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
	r := a.arbitrageDepthOpportunity(kask, kbid)
	if r.Volume == 0 || r.BuyPrice == 0 {
		return
	}

	perc := (r.WeightedSellPrice - r.WeightedBuyPrice) / r.BuyPrice * 100
	log.Info("Percent:", "info", perc)

	s := config.Cfg.Settings
	if r.Profit > s.ProfitThresh && perc > s.PercThresh {
		log.Info(
			fmt.Sprintf(
				"profit: %f CNY with volume: %f BTC - buy at %.4f (%s) sell at %.4f (%s) ~%.2f%%",
				r.Profit, r.Volume, r.BuyPrice, kask, r.SellPrice, kbid, perc,
			), "info",
		)
	}

	return
}

func (a *ArbitrageStrategy) arbitrageDepthOpportunity(kask, kbid string) ProfitStruct {
	var (
		profit                 ProfitStruct
		bestAskPos, bestBidPos int
	)

	askPos, bidPos := a.getMaxDepth(kask, kbid)
	for i := 0; i < askPos+1; i++ {
		for j := 0; j < bidPos+1; j++ {
			tempProfit := a.getProfitFor(i, j, kask, kbid)

			if tempProfit.Profit > 0 && tempProfit.Profit >= profit.Profit {
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

func (a *ArbitrageStrategy) getProfitFor(askPos, bidPos int, kask, kbid string) ProfitStruct {
	if a.Depths[kask].Asks[askPos].Price >= a.Depths[kbid].Bids[bidPos].Price {
		return ProfitStruct{}
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

	//log.Info("Volume", "info", config.Cfg.Settings.MaxTxVolume)
	maxAmount := math.Min(math.Min(maxAmountBuy, maxAmountSell), float64(1))

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
	return ProfitStruct{
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

		log.Info("Refrash rate:", "info", config.Cfg.Settings.RefreshRate)
		time.Sleep(time.Second * 5)
	}

	return
}
