package gemini

type (
	GeminiOrderbookEntry struct {
		Price     float64 `json:"price,string"`
		Amount    float64 `json:"amount,string"`
		Timestamp float64 `json:"timestamp,string"`
	}

	GeminiOrderBook struct {
		Bids []GeminiOrderbookEntry `json:"bids"`
		Asks []GeminiOrderbookEntry `json:"asks"`
	}

	GeminiOrder struct {
		OrderID           int64   `json:"order_id"`
		ClientOrderID     string  `json:"client_order_id"`
		Symbol            string  `json:"symbol"`
		Exchange          string  `json:"exchange"`
		Price             float64 `json:"price,string"`
		AvgExecutionPrice float64 `json:"avg_execution_price,string"`
		Side              string  `json:"side"`
		Type              string  `json:"type"`
		Timestamp         int64   `json:"timestamp"`
		TimestampMS       int64   `json:"timestampms"`
		IsLive            bool    `json:"is_live"`
		IsCancelled       bool    `json:"is_cancelled"`
		WasForced         bool    `json:"was_forced"`
		ExecutedAmount    float64 `json:"executed_amount,string"`
		RemainingAmount   float64 `json:"remaining_amount,string"`
		OriginalAmount    float64 `json:"original_amount,string"`
	}
)
