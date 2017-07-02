package bitfinex

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mgutz/logxi/v1"

	"goarbitrage/common"
	"goarbitrage/config"
	"goarbitrage/exchanges"
)

const (
	BITFINEX_API_URL      = "https://api.bitfinex.com/v1/"
	BITFINEX_API_VERSION  = "1"
	BITFINEX_ORDERBOOK    = "book/"
	BITFINEX_ORDER_NEW    = "order/new"
	BITFINEX_ORDER_CANCEL = "order/cancel"
	BITFINEX_ORDER_STATUS = "order/status"
	BITFINEX_SYMBOLS      = "symbols/"
)

type Bitfinex struct {
	exchange.ExchangeBase
}

func (b *Bitfinex) SetDefaults() {
	b.Name = "Bitfinex"
	b.Enabled = false
	b.Verbose = false
	b.RESTPollingDelay = 10
}

func (b *Bitfinex) Setup(exch config.Exchange) {
	if !exch.Enabled {
		b.SetEnabled(false)
		return
	}

	b.Enabled = true
	b.AuthenticatedAPISupport = exch.AuthenticatedAPISupport
	b.SetAPIKeys(exch.APIKey, exch.APISecret, "", false)
	b.RESTPollingDelay = exch.RESTPollingDelay
	b.Verbose = exch.Verbose
	b.Symbol = exch.Symbol
}

func (b *Bitfinex) GetOrderBook(symbol string, values url.Values) (BitfinexOrderBook, error) {
	var response BitfinexOrderBook
	path := common.EncodeURLValues(BITFINEX_API_URL+BITFINEX_ORDERBOOK+symbol, values)

	log.Info("Path:", "info", path)
	err := common.SendHTTPGetRequest(path, true, &response)
	if err != nil {
		return response, err
	}

	log.Info("Resp:", "info", "ok")

	return response, nil
}

func (b *Bitfinex) NewOrder(Symbol string, Amount float64, Price float64, Buy bool, Type string, Hidden bool) (BitfinexOrder, error) {
	request := make(map[string]interface{})
	request["symbol"] = Symbol
	request["amount"] = strconv.FormatFloat(Amount, 'f', -1, 64)
	request["price"] = strconv.FormatFloat(Price, 'f', -1, 64)
	request["exchange"] = "bitfinex"
	request["type"] = Type
	request["side"] = "sell"

	if Buy {
		request["side"] = "buy"
	}

	response := BitfinexOrder{}
	err := b.SendAuthenticatedHTTPRequest("POST", BITFINEX_ORDER_NEW, request, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

func (b *Bitfinex) CancelOrder(OrderID int64) (BitfinexOrder, error) {
	request := make(map[string]interface{})
	request["order_id"] = OrderID
	response := BitfinexOrder{}

	err := b.SendAuthenticatedHTTPRequest("POST", BITFINEX_ORDER_CANCEL, request, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

func (b *Bitfinex) GetOrderStatus(OrderID int64) (BitfinexOrder, error) {
	request := make(map[string]interface{})
	request["order_id"] = OrderID
	orderStatus := BitfinexOrder{}

	err := b.SendAuthenticatedHTTPRequest("POST", BITFINEX_ORDER_STATUS, request, &orderStatus)
	if err != nil {
		return orderStatus, err
	}

	return orderStatus, err
}

func (b *Bitfinex) GetSymbols() ([]string, error) {
	products := []string{}
	err := common.SendHTTPGetRequest(BITFINEX_API_URL+BITFINEX_SYMBOLS, true, &products)
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (b *Bitfinex) SendAuthenticatedHTTPRequest(method, path string, params map[string]interface{}, result interface{}) error {
	if len(b.APIKey) == 0 {
		return errors.New("SendAuthenticatedHTTPRequest: Invalid API key")
	}

	request := make(map[string]interface{})
	request["request"] = fmt.Sprintf("/v%s/%s", BITFINEX_API_VERSION, path)
	request["nonce"] = strconv.FormatInt(time.Now().UnixNano(), 10)

	if params != nil {
		for key, value := range params {
			request[key] = value
		}
	}

	PayloadJson, err := common.JSONEncode(request)
	if err != nil {
		return errors.New("SendAuthenticatedHTTPRequest: Unable to JSON request")
	}

	if b.Verbose {
		log.Info("Request JSON:", "info", PayloadJson)
	}

	PayloadBase64 := common.Base64Encode(PayloadJson)
	hmac := common.GetHMAC(common.HASH_SHA512_384, []byte(PayloadBase64), []byte(b.APISecret))
	headers := make(map[string]string)
	headers["X-BFX-APIKEY"] = b.APIKey
	headers["X-BFX-PAYLOAD"] = PayloadBase64
	headers["X-BFX-SIGNATURE"] = common.HexEncodeToString(hmac)

	resp, err := common.SendHTTPRequest(method, BITFINEX_API_URL+path, headers, strings.NewReader(""))
	if err != nil {
		return err
	}

	if strings.Contains(resp, "message") {
		return errors.New("SendAuthenticatedHTTPRequest: " + resp[11:])
	}

	if b.Verbose {
		log.Info("Recieved raw:", "info", resp)
	}

	err = common.JSONDecode([]byte(resp), &result)
	if err != nil {
		return errors.New("SendAuthenticatedHTTPRequest: Unable to JSON Unmarshal response.")
	}

	return nil
}
