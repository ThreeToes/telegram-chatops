package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
)

// Compare against the detailType field
const (
	PipelineChangeType = "CodePipeline Pipeline Execution State Change"
	StageChangeType = "CodePipeline Stage Execution State Change"
	ActionChangeType = "CodePipeline Action Execution State Change"
)

// Pipeline state changed statuses
const (
	PipelineCanceled = "CANCELED"
	PipelineFailed = "FAILED"
	PipelineResumed = "RESUMED"
	PipelineStarted = "STARTED"
	PipelineStopped = "STOPPED"
	PipelineStopping = "STOPPING"
	PipelineSucceeded = "SUCCEEDED"
	PipelineSuperseded = "SUPERSEDED"
)

// Stage state change statuses
const (
	StageCanceled = "CANCELED"
	StageFailed = "FAILED"
	StageResumed = "RESUMED"
	StageStarted = "STARTED"
	StageStopped = "STOPPED"
	StageStopping = "STOPPING"
	StageSucceeded = "SUCCEEDED"
)

const (
	ActionAbandoned = "ABANDONED"
	ActionCanceled = "CANCELED"
	ActionFailed = "FAILED"
	ActionStarted = "STARTED"
	ActionSucceeded = "SUCCEEDED"
)

var tg *tgbotapi.BotAPI

// main function. Required environment variables
// - TELEGRAM_TOKEN: Telegram bot API token
// - CHANNEL_ID: Telegram channel to send notifications to
func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
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
		var ev CodePipelineEvent
		err := json.Unmarshal([]byte(r.SNS.Message), &ev)
		if err != nil {
			logrus.Errorf("Could not convert message: %v", err)
			logrus.Errorf("Offending message: %v", r.SNS.Message)
			continue
		}
		var msg string
		//msg := fmt.Sprintf("<strong>CodePipeline Event for pipeline %s:</strong> %s\n", ev.Detail.Pipeline, ev.DetailType) +
		//	fmt.Sprintf("%s has changed to %s", ev.Detail.Stage, ev.Detail.State)
		switch ev.DetailType {
		case PipelineChangeType:
			msg = fmt.Sprintf("<strong>Pipeline state change event</strong>\n<pre>%s</pre> changed to <pre>%s</pre>",
				ev.Detail.Pipeline, ev.Detail.State)
		case StageChangeType:
			msg = fmt.Sprintf("<strong>Stage state change event</strong>\n"+
				"Stage <pre>%s</pre> in <pre>%s</pre> changed to <pre>%s</pre>",
				ev.Detail.Stage, ev.Detail.Pipeline, ev.Detail.State)
		case ActionChangeType:
			msg = fmt.Sprintf("<strong>Action state change event</strong>\n"+
				"Action <pre>%s</pre>, stage <pre>%s</pre> in <pre>%s</pre> changed to <pre>%s</pre>",
				ev.Detail.Action, ev.Detail.Stage, ev.Detail.Pipeline, ev.Detail.State)
		}
		logrus.Infof("Sending TG message: %s", msg)
		message := tgbotapi.NewMessage(h.chatId, msg)
		message.ParseMode = tgbotapi.ModeHTML
		_, err = tg.Send(message)
		if err != nil {
			logrus.Errorf("Could not send notification: %v", err)
		}
	}
	return nil
}