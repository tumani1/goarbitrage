package config

import (
	"os"
	"path"
	"time"

	"goarbitrage/common"
)

const (
	CONFIG_FILE = "config.json"
)

var (
	Cfg Config

	RootDir, _ = os.Getwd()
	CfgDir     = path.Join(RootDir, "configs")
)

type (
	Config struct {
		Telegram  Telegram            `json: telegram`
		Exchanges map[string]Exchange `json: exchanges`
		Settings  Settings            `json: settings`
	}

	Settings struct {
		RefreshInterval time.Duration `json: refresh_interval`
		MaxTxVolume     float64       `json: mat_tx_volume`
	}

	Telegram struct {
		Enable bool   `json: enable`
		ApiKey string `json: api_key`
		ChatId int64  `json: chat_id`
		Debug  bool   `json: debug`
	}

	Exchange struct {
		Name                    string `json: name`
		Enabled                 bool   `json: enabled`
		Verbose                 bool   `json: verbose`
		RESTPollingDelay        time.Duration
		AuthenticatedAPISupport bool   `json: auth_api_support`
		APIKey                  string `json: api_key`
		APISecret               string `json: api_secret`
		ClientID                string `json: client_id`
		Symbol                  string `json: symbol`
	}
)

func (c *Config) LoadFile() error {
	defaultPath := path.Join(CfgDir, CONFIG_FILE)
	file, err := common.ReadFile(defaultPath)
	if err != nil {
		return err
	}

	return common.JSONDecode(file, &c)
}

func GetConfig() *Config {
	return &Cfg
}
