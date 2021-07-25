package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
)

var tg *tgbotapi.BotAPI

// main function. Required environment variables
// - TELEGRAM_TOKEN: Telegram bot API token
// - CHANNEL_ID: Telegram channel to send notifications to
func main() {
	token := os.Getenv("TELEGRAM_TOKEN")
	chatId, err := strconv.ParseInt(os.Getenv("CHAT_ID"), 10, 64)
	if err != nil {
		logrus.Fatalf("Could not parse chat ID: %v", err)
	}
	tg, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		logrus.Fatalf("Could not instantiate Telegram API: %v", err)
	}
	h := &handler{
		chatId: chatId,
		tg:     tg,
	}
	logrus.Info("Starting lambda")
	lambda.Start(h.handleSns)
}

// handler is the lambda function itself
type handler struct {
	chatId int64
	tg *tgbotapi.BotAPI
}

func (h *handler) handleSns(_ context.Context, e *events.SNSEvent) error {
	for _, r := range e.Records {
		_, err := tg.Send(tgbotapi.NewMessage(h.chatId, r.SNS.Message))
		if err != nil {
			logrus.Errorf("Could not send notification: %v", err)
		}
	}
	return nil
}