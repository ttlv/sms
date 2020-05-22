package server

import (
	"github.com/jinzhu/gorm"
	"github.com/ttlv/sms"
	"github.com/ttlv/sms/internal"
)

func New(db *gorm.DB, queue sms.SmsQueue, providers []sms.SmsProvider) (serv internal.SmsServer, err error) {
	return internal.New(db, queue, providers)
}

func NewConsumer(db *gorm.DB, providers []sms.SmsProvider) Consumer {
	db.AutoMigrate(&sms.SmsRecord{}, &sms.SmsFailureRecord{}, &sms.SmsBrand{}, &sms.SmsSetting{})
	return Consumer{Providers: providers}
}
