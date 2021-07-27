.PHONY: build build-status

build: mod build-status build-commands

mod:
	go mod tidy

build-status:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o build/status ./cmd/pipeline-status/...
	cd build/ && rm -f status.zip && zip -r status.zip status

build-commands:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o build/commands ./cmd/command/...
	cd build/ && rm -f commands.zip && zip -r commands.zip commands

bot_subdomain := "pipelinebot"
ifdef $(BOT_SUBDOMAIN)
bot_subdomain = $(BOT_SUBDOMAIN)
endif
# Required environment variables:
#   * CLOUDFORMATION_BUCKET - Cloudformation artifact bucket
#   * TELEGRAM_TOKEN - Telegram API token
#   * TELEGRAM_CHAT_ID - Telegram chat ID to put the status updates into
#	* BOT_DOMAIN - Domain to put the bot on
#	* HOSTED_ZONE - ID of the hosted zone to put the bot on
# Optional environment variables:
#   * BOT_SUBDOMAIN - Bot subdomain. Telegram expects a somewhat public API (annoyingly)
deploy: build
	@aws cloudformation package \
    	--template-file ./cfn/chatstack.yml \
        --output-template-file build/stack.yml \
        --s3-bucket $(CLOUDFORMATION_BUCKET)
	aws cloudformation deploy \
		--stack-name telegram-status \
		--template-file ./build/stack.yml \
		--no-fail-on-empty-changeset \
		--capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM \
		--parameter-overrides TelegramToken="$(TELEGRAM_TOKEN)" TelegramChatId="$(TELEGRAM_CHAT_ID)" \
			Route53ZoneId="$(HOSTED_ZONE)" BotDomain="$(BOT_DOMAIN)" BotSubDomain="$(BOT_SUBDOMAIN)"

# Create bot webhook
create-webhook:
	curl https://api.telegram.org/bot$(TELEGRAM_TOKEN)/setWebhook?url=https://$(BOT_SUBDOMAIN).$(BOT_DOMAIN)/

delete-webhook:
	curl -X DELETE https://api.telegram.org/bot$(TELEGRAM_TOKEN)/setWebhook?url=https://$(BOT_SUBDOMAIN).$(BOT_DOMAIN)/