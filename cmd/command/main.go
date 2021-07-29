package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
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
	logrus.Infof("Using chat ID %d", chatId)
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

func (c *commandHandler) handler(ctx context.Context, event *events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	var telegramEvent TelegramMessage
	err := json.Unmarshal([]byte(event.Body), &telegramEvent)
	logrus.Infof(event.Body)
	if err != nil {
		return &events.APIGatewayV2HTTPResponse{ StatusCode: http.StatusInternalServerError}, err
	}
	if telegramEvent.Message.Chat.ID != c.chatId {
		return &events.APIGatewayV2HTTPResponse{StatusCode: http.StatusUnauthorized}, fmt.Errorf("chat ID doesn't match what was configured")
	}
	cmdSplit := strings.SplitN(telegramEvent.Message.Text, " ", 2)
	logrus.Infof("Message %s, Command split %v", telegramEvent.Message.Text, cmdSplit)
	if len(cmdSplit) > 0 {
		switch cmdSplit[0] {
		case "/list":
			// List pipelines
			c.listPipelines(ctx, telegramEvent.Message.MessageID)
		case "/status":
			if len(cmdSplit) < 2{
				msg := tgbotapi.NewMessage(c.chatId, "Please provide a pipeline name")
				msg.ReplyToMessageID = telegramEvent.Message.MessageID
				c.tg.Send(msg)
				break
			}
			err := c.getPipelineStatus(ctx, cmdSplit[1], telegramEvent.Message.MessageID)
			if err != nil {
				logrus.Errorf("could not send message: %v", err)
			}
		case "/subscriptions":
			// Manage subscriptions
			break
		}
	}
	return &events.APIGatewayV2HTTPResponse{ StatusCode: http.StatusOK}, nil
}

func (c *commandHandler) listPipelines(ctx context.Context, replyTo int) error {
	pipelines, err := c.codePipeline.ListPipelines(ctx, &codepipeline.ListPipelinesInput{})
	if err != nil {
		logrus.Errorf("Could not list pipelines: %v", err)
		msg := tgbotapi.NewMessage(c.chatId, "Could not list pipelines, check cloudwatch logs")
		c.tg.Send(msg)
		return err
	}
	sb := &strings.Builder{}
	for _, p := range pipelines.Pipelines {
		sb.WriteString(fmt.Sprintf("* <code>%s</code>\n", *p.Name))
	}
	msg := tgbotapi.NewMessage(c.chatId, fmt.Sprintf("Pipelines:\n%s", sb.String()))
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyToMessageID = replyTo
	_, err = c.tg.Send(msg)
	return err
}

func (c *commandHandler) getPipelineStatus(ctx context.Context, pipelineName string, replyTo int) error {
	p, err := c.codePipeline.GetPipelineState(ctx, &codepipeline.GetPipelineStateInput{
		Name:    aws.String(pipelineName),
	})
	if err != nil {
		logrus.Errorf("Could not get pipeline: %v", err)
		msg := tgbotapi.NewMessage(c.chatId, fmt.Sprintf("Could not get pipeline `%s`'s status", pipelineName))
		msg.ParseMode = tgbotapi.ModeMarkdown
		c.tg.Send(msg)
		return err
	}
	sb := &strings.Builder{}
	sb.WriteString(fmt.Sprintf("Pipeline <code>%s</code>:\n", pipelineName))
	for _, s := range p.StageStates {
		sb.WriteString(fmt.Sprintf("\t* Stage <code>%s</code> - <code>%s</code>\n", *s.StageName, s.LatestExecution.Status))
	}
	msg := tgbotapi.NewMessage(c.chatId, sb.String())
	msg.ReplyToMessageID = replyTo
	msg.ParseMode = tgbotapi.ModeHTML
	_, err = c.tg.Send(msg)
	return err
}