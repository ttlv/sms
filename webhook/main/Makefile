SHELL = /bin/bash

build:
	@GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -o sms-webhook main.go
	@docker build -t gopherlv/sms-webhook .
	@rm sms-webhook

push: build
	@$(eval REV := $(shell git rev-parse HEAD|cut -c 1-8))
	@docker tag gopherlv/sms-webhook gopherlv/sms_service:sms-webhook-$(REV)
	@docker push gopherlv/sms_service:sms-webhook-$(REV)
