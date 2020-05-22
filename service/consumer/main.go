package main

import (
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/ttlv/sms"
	"github.com/ttlv/sms/config"
	"github.com/ttlv/sms/provider/amazon"
	"github.com/ttlv/sms/provider/chuanglan"
	"github.com/ttlv/sms/provider/emay"
	"github.com/ttlv/sms/provider/twilio"
	"github.com/ttlv/sms/provider/yunpian"
	"github.com/ttlv/sms/server"
)

func main() {
	cfg := config.MustGetConfig()
	db, err := gorm.Open("mysql", cfg.DB)
	if err != nil {
		panic(err)
	}
	consumer := server.NewConsumer(db, []sms.SmsProvider{
		yunpian.New(),
		emay.New(),
		amazon.New(),
		twilio.New(),
		chuanglan.New(),
	})

	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for _ = range ticker.C {
			consumer.SortProviders()
		}
	}()

	consumer.Run()
}
