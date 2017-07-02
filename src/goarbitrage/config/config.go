package config

import (
	"os"
	"path"
	"time"

	"github.com/mgutz/logxi/v1"

	"encoding/json"
	"goarbitrage/common"
)

const (
	CONFIG_FILE = "config.json"
)

var (
	Cfg        = &Config{}
	RootDir, _ = os.Getwd()
	CfgDir     = path.Join(RootDir, "configs")
)

type (
	Config struct {
		Telegram  Telegram            `json: "telegram"`
		Exchanges map[string]Exchange `json: "exchanges"`
		Settings  Settings            `json: "settings"`
	}

	Settings struct {
		RefreshRate        time.Duration `json: "refresh_rate"`
		MaxTxVolume        float64       `json: "max_tx_volume"`
		MinTxVolume        float64       `json: "min_tx_volume"`
		ProfitThresh       float64       `json: "profit_thresh"`
		PercThresh         float64       `json: "perc_thresh"`
		ArbitrageBuyQueue  int           `json: "arbitrage_buy_queue"`
		ArbitrageSellQueue int           `json: "arbitrage_sell_queue"`
	}

	Telegram struct {
		Enable bool   `json: "enable"`
		ApiKey string `json: "api_key"`
		ChatId int64  `json: "chat_id"`
		Debug  bool   `json: "debug"`
	}

	Exchange struct {
		Name                    string `json: "name"`
		Enabled                 bool   `json: "enabled"`
		Verbose                 bool   `json: "verbose"`
		RESTPollingDelay        time.Duration
		AuthenticatedAPISupport bool   `json: "auth_api_support"`
		APIKey                  string `json: "api_key"`
		APISecret               string `json: "api_secret"`
		ClientID                string `json: "client_id"`
		Symbol                  string `json: "symbol"`
	}
)

func Init() {
	// Try load config from file system
	err := LoadFile()
	if err != nil {
		log.Fatal("Error load config file:", "fatal", err.Error())
	}
}

func LoadFile() error {
	defaultPath := path.Join(CfgDir, CONFIG_FILE)
	file, err := common.ReadFile(defaultPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(file, Cfg)
}
