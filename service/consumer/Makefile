SHELL = /bin/bash

build:
	@GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -o sms-consumer main.go
	@docker build -t gopherlv/sms-consumer .
	@rm sms-consumer

push: build
	@$(eval REV := $(shell git rev-parse HEAD|cut -c 1-8))
	@docker tag gopherlv/sms-consumer gopherlv/sms_service:sms-consumer-$(REV)
	@docker push gopherlv/sms_service:sms-consumer-$(REV)
