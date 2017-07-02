package exchange

import (
	"log"
	"time"

	"goarbitrage/common"
	"goarbitrage/config"
	"sync"
)

const (
	WarningBase64DecryptSecretKeyFailed = "WARNING -- Exchange %s unable to base64 decode secret key.. Disabling Authenticated API support."
)

type (
	ExchangeBase struct {
		Name                        string
		Enabled                     bool
		Verbose                     bool
		RESTPollingDelay            time.Duration
		AuthenticatedAPISupport     bool
		APISecret, APIKey, ClientID string
		TakerFee, MakerFee, Fee     float64
		Symbol                      string
		APIUrl                      string
	}

	ItemBook struct {
		Price     float64
		Amount    float64
		Timestamp float64
	}

	OrderBook struct {
		Bids []ItemBook
		Asks []ItemBook
	}

	TaskResponse struct {
		Name      string
		OrderBook OrderBook
	}

	IBotExchange interface {
		Setup(exch config.Exchange)
		UpdateDepth(wg *sync.WaitGroup, done chan struct{}, resp chan TaskResponse)
		SetDefaults()
		GetName() string
		GetSymbol() string
		IsEnabled() bool
	}
)

func (e *ExchangeBase) GetName() string {
	return e.Name
}
func (e *ExchangeBase) GetSymbol() string {
	return e.Symbol
}
func (e *ExchangeBase) SetEnabled(enabled bool) {
	e.Enabled = enabled
}

func (e *ExchangeBase) IsEnabled() bool {
	return e.Enabled
}

func (e *ExchangeBase) SetAPIKeys(APIKey, APISecret, ClientID string, b64Decode bool) {
	e.APIKey = APIKey
	e.ClientID = ClientID

	if b64Decode {
		result, err := common.Base64Decode(APISecret)
		if err != nil {
			e.AuthenticatedAPISupport = false
			log.Printf(WarningBase64DecryptSecretKeyFailed, e.Name)
		}
		e.APISecret = string(result)
	} else {
		e.APISecret = APISecret
	}
}
