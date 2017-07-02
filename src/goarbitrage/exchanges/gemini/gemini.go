package gemini

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
	GEMINI_API_URL     = "https://api.gemini.com"
	GEMINI_API_VERSION = "1"

	GEMINI_SYMBOLS      = "symbols"
	GEMINI_ORDERBOOK    = "book"
	GEMINI_ORDERS       = "orders"
	GEMINI_ORDER_NEW    = "order/new"
	GEMINI_ORDER_CANCEL = "order/cancel"
	GEMINI_ORDER_STATUS = "order/status"
)

type Gemini struct {
	exchange.ExchangeBase
}

func (g *Gemini) SetDefaults() {
	g.Name = "Gemini"
	g.Enabled = false
	g.Verbose = false
	g.RESTPollingDelay = 10
}

func (g *Gemini) Setup(exch config.Exchange) {
	if !exch.Enabled {
		g.SetEnabled(false)
		return
	}

	g.Enabled = true
	g.AuthenticatedAPISupport = exch.AuthenticatedAPISupport
	g.SetAPIKeys(exch.APIKey, exch.APISecret, "", false)
	g.RESTPollingDelay = exch.RESTPollingDelay
	g.Verbose = exch.Verbose
	g.Symbol = exch.Symbol
}

func (g *Gemini) GetSymbols() ([]string, error) {
	symbols := []string{}
	path := fmt.Sprintf("%s/v%s/%s", GEMINI_API_URL, GEMINI_API_VERSION, GEMINI_SYMBOLS)
	err := common.SendHTTPGetRequest(path, true, &symbols)
	if err != nil {
		return nil, err
	}
	return symbols, nil
}

func (g *Gemini) GetOrderBook(currency string, params url.Values) (GeminiOrderBook, error) {
	var response GeminiOrderBook
	path := common.EncodeURLValues(fmt.Sprintf("%s/v%s/%s/%s", GEMINI_API_URL, GEMINI_API_VERSION, GEMINI_ORDERBOOK, currency), params)

	err := common.SendHTTPGetRequest(path, true, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

func (g *Gemini) NewOrder(symbol string, amount, price float64, side, orderType string) (int64, error) {
	request := make(map[string]interface{})
	request["symbol"] = symbol
	request["amount"] = strconv.FormatFloat(amount, 'f', -1, 64)
	request["price"] = strconv.FormatFloat(price, 'f', -1, 64)
	request["side"] = side
	request["type"] = orderType

	response := GeminiOrder{}
	err := g.SendAuthenticatedHTTPRequest("POST", GEMINI_ORDER_NEW, request, &response)
	if err != nil {
		return 0, err
	}
	return response.OrderID, nil
}

func (g *Gemini) CancelOrder(OrderID int64) (GeminiOrder, error) {
	request := make(map[string]interface{})
	request["order_id"] = OrderID

	response := GeminiOrder{}
	err := g.SendAuthenticatedHTTPRequest("POST", GEMINI_ORDER_CANCEL, request, &response)
	if err != nil {
		return GeminiOrder{}, err
	}
	return response, nil
}

func (g *Gemini) GetOrderStatus(orderID int64) (GeminiOrder, error) {
	request := make(map[string]interface{})
	request["order_id"] = orderID

	response := GeminiOrder{}
	err := g.SendAuthenticatedHTTPRequest("POST", GEMINI_ORDER_STATUS, request, &response)
	if err != nil {
		return GeminiOrder{}, err
	}

	return response, nil
}

func (g *Gemini) GetOrders() ([]GeminiOrder, error) {
	response := []GeminiOrder{}
	err := g.SendAuthenticatedHTTPRequest("POST", GEMINI_ORDERS, nil, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (g *Gemini) SendAuthenticatedHTTPRequest(method, path string, params map[string]interface{}, result interface{}) (err error) {
	request := make(map[string]interface{})
	request["request"] = fmt.Sprintf("/v%s/%s", GEMINI_API_VERSION, path)
	request["nonce"] = time.Now().UnixNano()

	if params != nil {
		for key, value := range params {
			request[key] = value
		}
	}

	PayloadJson, err := common.JSONEncode(request)
	if err != nil {
		return errors.New("SendAuthenticatedHTTPRequest: Unable to JSON request")
	}

	if g.Verbose {
		log.Info("Request JSON:", "info", PayloadJson)
	}

	PayloadBase64 := common.Base64Encode(PayloadJson)
	hmac := common.GetHMAC(common.HASH_SHA512_384, []byte(PayloadBase64), []byte(g.APISecret))
	headers := make(map[string]string)
	headers["X-GEMINI-APIKEY"] = g.APIKey
	headers["X-GEMINI-PAYLOAD"] = PayloadBase64
	headers["X-GEMINI-SIGNATURE"] = common.HexEncodeToString(hmac)

	resp, err := common.SendHTTPRequest(method, GEMINI_API_URL+path, headers, strings.NewReader(""))
	if g.Verbose {
		log.Info("Recieved raw:", "info", resp)
	}

	err = common.JSONDecode([]byte(resp), &result)
	if err != nil {
		return errors.New("Unable to JSON Unmarshal response.")
	}

	return nil
}
