package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	"github.com/aws/aws-sdk-go-v2/service/codestarnotifications"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type commandHandler struct {
	tg *tgbotapi.BotAPI
	chatId int64
	codePipeline *codepipeline.Client
	notifications *codestarnotifications.Client
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.Infof("Starting lambda")
	token := os.Getenv("TELEGRAM_TOKEN")
	chatId, err := strconv.ParseInt(os.Getenv("CHAT_ID"), 10, 64)
	if err != nil {
		logrus.Fatalf("Could not parse chat ID: %v", err)
	}
	tg, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		logrus.Fatalf("Could not instantiate Telegram API: %v", err)
	}
	conf, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		logrus.Fatalf("Could not instantiate AWS config: %v", err)
	}
	pipeline := codepipeline.NewFromConfig(conf)
	notifications := codestarnotifications.NewFromConfig(conf)
	c := &commandHandler{
		tg:     tg,
		chatId: chatId,
		codePipeline: pipeline,
		notifications: notifications,
	}
	lambda.Start(c.handler)
}

func (c *commandHandler) handler(ctx context.Context, event *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var telegramEvent TelegramMessage
	err := json.Unmarshal([]byte(event.Body), &telegramEvent)
	logrus.Infof(event.Body)
	if err != nil {
		return &events.APIGatewayProxyResponse{ StatusCode: http.StatusInternalServerError}, err
	}
	if telegramEvent.Message.Chat.ID != c.chatId {
		return &events.APIGatewayProxyResponse{StatusCode: http.StatusUnauthorized}, fmt.Errorf("chat ID doesn't match what was configured")
	}
	return &events.APIGatewayProxyResponse{ StatusCode: http.StatusOK}, nil
}

func (c *commandHandler) listPipelines(ctx context.Context) error {
	pipelines, err := c.codePipeline.ListPipelines(ctx, &codepipeline.ListPipelinesInput{})
	if err != nil {
		return err
	}
	sb := &strings.Builder{}
	for _, p := range pipelines.Pipelines {
		sb.WriteString(fmt.Sprintf("* <code>%s</code>\n", *p.Name))
	}
	msg := tgbotapi.NewMessage(c.chatId, fmt.Sprintf("Pipelines:\n%s", sb.String()))
	msg.ParseMode = tgbotapi.ModeHTML
	_, err = c.tg.Send(msg)
	return err
}