SHELL = /bin/bash

build:
	@GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -o sms_monitor main.go
	@docker build -t hashtech/sms_monitor .
	@rm sms_monitor

push: build
	@$(eval REV := $(shell git rev-parse HEAD|cut -c 1-8))
	@docker tag hashtech/sms_monitor registry.gitlab.com/hashtech/common:sms-monitor-$(REV)
	@docker tag hashtech/sms_monitor registry.gitlab.com/hashtech/common:sms-monitor
	@docker push registry.gitlab.com/hashtech/common:sms-monitor-$(REV)
	@docker push registry.gitlab.com/hashtech/common:sms-monitor
