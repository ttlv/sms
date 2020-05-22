SHELL = /bin/bash

build:
	@GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -o sms_producer main.go
	@docker build -t hashtech/sms_producer .
	@rm sms_producer

push: build
	@$(eval REV := $(shell git rev-parse HEAD|cut -c 1-8))
	@docker tag hashtech/sms_producer registry.gitlab.com/hashtech/common:sms-producer-$(REV)
	@docker tag hashtech/sms_producer registry.gitlab.com/hashtech/common:sms-producer
	@docker push registry.gitlab.com/hashtech/common:sms-producer-$(REV)
	@docker push registry.gitlab.com/hashtech/common:sms-producer
