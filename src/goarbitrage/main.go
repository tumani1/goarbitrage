package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/mgutz/logxi/v1"

	"goarbitrage/arbitrage"
	"goarbitrage/common"
	"goarbitrage/config"
	"goarbitrage/exchanges"
	"goarbitrage/exchanges/bitfinex"
	"goarbitrage/exchanges/gemini"
	"goarbitrage/telegram"
)

type (
	Bot struct {
		config    *config.Config
		arbitrer  *arbitrage.ArbitrageStrategy
		exchanges map[string]exchange.IBotExchange
		shutdown  chan bool
	}
)

var (
	bot Bot
)

func HandleInterrupt() {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-s
		log.Info("Captured signal", "info", sig.String())
		Shutdown()
	}()
}

func Shutdown() {
	log.Info("Shutting down...", "info")
	os.Exit(1)
}

func main() {
	HandleInterrupt()

	// ---------------------------------------
	log.Info("Load config file...")
	config.Init()

	// ---------------------------------------
	cfg := config.Cfg
	if cfg.Telegram.Enable {
		log.Info("Load telegram notify...")
		if err := telegram.Init(); err != nil {
			log.Fatal("Error load telegram notify", "fatal", err.Error())
		}
	} else {
		log.Info("Telegram disabled", "info")
	}

	// ---------------------------------------
	log.Info("Init exchanges...")
	bot.exchanges = map[string]exchange.IBotExchange{}
	for _, i := range []exchange.IBotExchange{
		new(bitfinex.Bitfinex),
		new(gemini.Gemini),
	} {
		if i == nil {
			continue
		}

		i.SetDefaults()
		i.Setup(cfg.Exchanges[i.GetName()])
		bot.exchanges[i.GetName()] = i
		log.Info("Successfully set settings for exchange:", i.GetName())
	}

	// ---------------------------------------
	log.Info("Init arbitrage...")
	bot.arbitrer = arbitrage.New()
	bot.arbitrer.Exchanges = bot.exchanges

	// ---------------------------------------
	log.Info("Start watch loop...")
	bot.arbitrer.Loop()

	// ---------------------------------------
	<-common.QuitChan
	Shutdown()
}
