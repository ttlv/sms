package server_test

import (
	"github.com/jinzhu/gorm"
	"github.com/ttlv/common_utils/utils"
	"github.com/ttlv/sms"
	"github.com/ttlv/sms/config"
	"github.com/ttlv/sms/internal"
	"github.com/ttlv/sms/provider/dummy"
	"github.com/ttlv/sms/server"
)

var SmsServer internal.SmsServer
var DB *gorm.DB

func setup(smsQueue sms.SmsQueue) {
	var err error
	cfg := config.MustGetConfig()
	DB, err = gorm.Open("mysql", cfg.DB)
	if err != nil {
		panic(err)
	}
	DB.AutoMigrate(&sms.SmsRecord{}, &sms.SmsFailureRecord{}, &sms.SmsSetting{}, &sms.SmsBrand{})
	utils.RunSQL(DB, `
	  TRUNCATE TABLE sms_records;
	  TRUNCATE TABLE sms_failure_records;
	  TRUNCATE TABLE sms_settings;
	  TRUNCATE TABLE sms_brands;
	`)
	providers := []sms.SmsProvider{
		dummy.DummyProvider{Code: "YunPian", Countries: []string{"CN"}, SendFunc: func(params sms.SendParams) (string, string, error) {
			return "", "", nil
		}},
		dummy.DummyProvider{Code: "Emay", Countries: []string{"CN"}, SendFunc: func(params sms.SendParams) (string, string, error) {
			return "", "", nil
		}},
		dummy.DummyProvider{Code: "Amazon", Countries: []string{""}, SendFunc: func(params sms.SendParams) (string, string, error) {
			return "", "", nil
		}},
		dummy.DummyProvider{Code: "Twilio", Countries: []string{""}, SendFunc: func(params sms.SendParams) (string, string, error) {
			return "", "", nil
		}},
	}
	SmsServer, _ = server.New(DB, smsQueue, providers)
}
