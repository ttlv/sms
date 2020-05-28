SHELL = /bin/bash

build:
	@GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -o sms-grpc-server main.go
	@docker build -t gopherlv/sms-grpc-server .
	@rm sms-grpc-server

push: build
	@$(eval REV := $(shell git rev-parse HEAD|cut -c 1-8))
	@docker tag gopherlv/sms-grpc-server gopherlv/sms_service:sms-grpc-server-$(REV)
	@docker push gopherlv/sms_service:sms-grpc-server-$(REV)
