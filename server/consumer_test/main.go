package main

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/ttlv/sms"
	"github.com/ttlv/sms/config"
	"github.com/ttlv/sms/provider/dummy"
	"github.com/ttlv/sms/server"
)

func main() {
	cfg := config.MustGetConfig()
	db, err := gorm.Open("mysql", cfg.DB)
	db.AutoMigrate(&sms.SmsRecord{}, &sms.SmsFailureRecord{}, &sms.SmsBrand{}, &sms.SmsSetting{})
	if err != nil {
		panic(err)
	}
	consumer := server.NewConsumer(db, []sms.SmsProvider{
		dummy.DummyProvider{DB: db, Code: "YunPian", Countries: []string{"CN"}, SendFunc: func(params sms.SendParams) (string, string, error) {
			if params.Phone == "+86 186 0000 0001" {
				return "success", "", nil
			} else if params.Phone == "+86 186 0000 0002" {
				return "success", "", nil
			} else if params.Phone == "+86 186 0000 0003" {
				return "success", "", nil
			} else if params.Phone == "+86 186 0000 0004" {
				return "success", "", nil
			} else if params.Phone == "+86 196 0000 0009" {
				return "", "", fmt.Errorf("%s", "发送不出去")
			} else if params.Phone == "+86 196 0000 0007" {
				return "", "", fmt.Errorf("%s", "发送不出去")
			} else if params.Phone == "+86 196 0000 0001" {
				return "success", "4000", nil
			} else if params.Phone == "+86 196 0000 0002" {
				return "success", "4001", nil
			} else if params.Phone == "+86 196 0000 0003" {
				return "success", "4002", nil
			} else if params.Phone == "+86 196 0000 0004" {
				return "success", "4003", nil
			}
			return "", "", fmt.Errorf("%s", "发送失败")
		}},
		dummy.DummyProvider{DB: db, Code: "Emay", Countries: []string{"CN"}, SendFunc: func(params sms.SendParams) (string, string, error) {
			if params.Phone == "+86 186 0000 0006" {
				return "success", "", nil
			} else if params.Phone == "+86 186 0000 0001" {
				return "success", "", nil
			} else if params.Phone == "+86 196 0000 0002" {
				return "success", "5001", nil
			} else if params.Phone == "+86 196 0000 0008" {
				return "success", "", fmt.Errorf("%s", "发送不出去")
			} else if params.Phone == "+86 196 0000 0003" {
				return "failed", "5002", fmt.Errorf("error")
			} else if params.Phone == "+86 196 0000 0004" {
				return "success", "5003", nil
			} else if params.Phone == "+86 196 0000 0009" {
				return "success", "", nil
			} else if params.Phone == "+86 196 0000 0007" {
				return "", "", fmt.Errorf("%s", "发送不出去")
			}
			return "", "", fmt.Errorf("%s", "发送失败")
		}},
		dummy.DummyProvider{DB: db, Code: "Twilio", Countries: []string{}, SendFunc: func(params sms.SendParams) (string, string, error) {
			if params.Phone == "+86 186 0000 0007" {
				return "success", "", nil
			} else if params.Phone == "+60 10-514 6182" {
				return "success", "", nil
			} else if params.Phone == "+86 196 0000 0007" {
				return "", "", nil
			}
			return "", "", fmt.Errorf("%s", "发送失败")
		}},
		dummy.DummyProvider{DB: db, Code: "Amazon", Countries: []string{}, SendFunc: func(params sms.SendParams) (string, string, error) {
			return "", "", fmt.Errorf("%s", "发送失败")
		}},
	})
	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for _ = range ticker.C {
			consumer.SortProviders()
		}
	}()
	consumer.Run()
}
