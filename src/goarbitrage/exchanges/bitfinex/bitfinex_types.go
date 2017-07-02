package bitfinex

type (
	BitfinexBookStructure struct {
		Price     float64 `json:"price,string"`
		Amount    float64 `json:"amount,string"`
		Timestamp float64 `json:"timestamp,string"`
	}

	BitfinexOrderBook struct {
		Bids []BitfinexBookStructure `json:"bids"`
		Asks []BitfinexBookStructure `json:"asks"`
	}

	BitfinexOrder struct {
		ID                    int64
		Symbol                string
		Exchange              string
		Price                 float64 `json:"price,string"`
		AverageExecutionPrice float64 `json:"avg_execution_price,string"`
		Side                  string
		Type                  string
		Timestamp             string
		IsLive                bool    `json:"is_live"`
		IsCancelled           bool    `json:"is_cancelled"`
		IsHidden              bool    `json:"is_hidden"`
		WasForced             bool    `json:"was_forced"`
		OriginalAmount        float64 `json:"original_amount,string"`
		RemainingAmount       float64 `json:"remaining_amount,string"`
		ExecutedAmount        float64 `json:"executed_amount,string"`
		OrderID               int64   `json:"order_id"`
	}
)
