.PHONY: build build-status

build: build-status

build-status:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o build/status ./cmd/pipeline-status/...
	cd build/ && rm -f status.zip && zip -r status.zip status

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
		--parameter-overrides TelegramToken="$(TELEGRAM_TOKEN)" TelegramChatId="$(TELEGRAM_CHAT_ID)"