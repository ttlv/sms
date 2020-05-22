SHELL = /bin/bash

build:
	@GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -o sms_webhook main.go
	@docker build -t hashtech/sms_webhook .
	@rm sms_webhook

push: build
	@$(eval REV := $(shell git rev-parse HEAD|cut -c 1-8))
	@docker tag hashtech/sms_webhook registry.gitlab.com/hashtech/common:sms-webhook-$(REV)
	@docker tag hashtech/sms_webhook registry.gitlab.com/hashtech/common:sms-webhook
	@docker push registry.gitlab.com/hashtech/common:sms-webhook-$(REV)
	@docker push registry.gitlab.com/hashtech/common:sms-webhook
