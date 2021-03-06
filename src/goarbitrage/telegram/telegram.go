package telegram

import (
	"fmt"

	"github.com/mgutz/logxi/v1"
	"gopkg.in/telegram-bot-api.v4"

	"goarbitrage/config"
)

var (
	Bot *tgbotapi.BotAPI
)

func Init() error {
	var (
		err error
	)

	c := config.Cfg
	Bot, err = tgbotapi.NewBotAPI(c.Telegram.ApiKey)
	if err != nil {
		return fmt.Errorf("Init bot api", err.Error())
	}

	if Bot.Self.UserName == "" {
		return fmt.Errorf("Error connecting to Telegram!")
	}

	log.Info("Authorized on account", "info", Bot.Self.UserName)
	Bot.Debug = c.Telegram.Debug

	return nil
}

func SendTelegramMessage(message string) error {
	c := config.Cfg
	msg := tgbotapi.NewMessage(c.Telegram.ChatId, message)
	_, err := Bot.Send(msg)
	if err != nil {
		return fmt.Errorf("Error send message:", err.Error())
	}

	return nil
}

func StartUpBot() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	update, err := Bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal("Could not get chan for updates", "fatal", err.Error())
	}

	for {
		select {
		case data := <-update:
			log.Info("Get data:", "info", data)
		}
	}
}
