FROM alpine
RUN apk --update upgrade && \
    apk add ca-certificates && \
    apk add tzdata && \
    rm -rf /var/cache/apk/*

ADD sms-webhook /bin/
CMD /bin/sms-webhook
