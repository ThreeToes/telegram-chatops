.PHONY: build build-status

build: build-status

build-status:
	GOOS=linux GOARCH=x86_64 go build -o build/status ./cmd/pipeline-status/main.go .
	cd build/ && zip -r status status.zip

deploy: build
	aws cloudformation deploy \
		--stack-name telegram-status \
		--template-file ./cfn/chatstack.yml \
		--no-fail-on-empty-changeset