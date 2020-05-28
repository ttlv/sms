SHELL = /bin/bash

build:
	@GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -o sms-monitor main.go
	@docker build -t gopherlv/sms-monitor .
	@rm sms-monitor

push: build
	@$(eval REV := $(shell git rev-parse HEAD|cut -c 1-8))
	@docker tag gopherlv/sms-monitor gopherlv/sms_service:sms-monitor-$(REV)
	@docker push gopherlv/sms_service:sms-monitor-$(REV)
